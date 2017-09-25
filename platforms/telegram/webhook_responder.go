package telegram_bot

import (
	"encoding/json"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"net/http"
	"github.com/strongo/app/log"
	"bytes"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
)

type TelegramWebhookResponder struct {
	w   http.ResponseWriter
	whc *TelegramWebhookContext
}

var _ bots.WebhookResponder = (*TelegramWebhookResponder)(nil)

func NewTelegramWebhookResponder(w http.ResponseWriter, whc *TelegramWebhookContext) TelegramWebhookResponder {
	responder := TelegramWebhookResponder{w: w, whc: whc}
	whc.responder = responder
	return responder
}

func (r TelegramWebhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotApiSendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	log.Debugf(c, "TelegramWebhookResponder.SendMessage(channel=%v, isEdit=%v) => m: %v", channel, m.IsEdit, m)
	if channel != bots.BotApiSendMessageOverHTTPS && channel != bots.BotApiSendMessageOverResponse {
		panic(fmt.Sprintf("Unknown channel: [%v]. Expected either 'https' or 'response'.", channel))
	}
	//ctx := tc.Context()

	var chattable tgbotapi.Chattable

	parseMode := func() string {
		switch m.Format {
		case bots.MessageFormatHTML:
			return "HTML"
		case bots.MessageFormatMarkdown:
			return "Markdown"
		}
		return ""
	}

	tgUpdate := r.whc.Input().(TelegramWebhookUpdateProvider).TgUpdate()

	if m.TelegramCallbackAnswer != nil {
		log.Debugf(c, "Inline answer")
		if m.TelegramCallbackAnswer.CallbackQueryID == "" && tgUpdate.CallbackQuery != nil {
			m.TelegramCallbackAnswer.CallbackQueryID = tgUpdate.CallbackQuery.ID
		}

		chattable = m.TelegramCallbackAnswer
	} else if editMsgTxt := m.TelegramEditMessageText; editMsgTxt != nil {
		log.Debugf(c, "m.TelegramEditMessageMarkup != nil")
		if editMsgTxt.ReplyMarkup == nil && m.TelegramKeyboard != nil {
			editMsgTxt.ReplyMarkup = m.TelegramKeyboard.(*tgbotapi.InlineKeyboardMarkup)
		}
		if editMsgTxt.InlineMessageID == "" || editMsgTxt.ChatID == 0 || editMsgTxt.MessageID == 0 {
			inlineMessageID, chatID, messageID := getTgMessageIDs(tgUpdate)
			switch {
			case inlineMessageID != "":
				editMsgTxt.InlineMessageID = inlineMessageID
				editMsgTxt.ChatID = 0
				editMsgTxt.MessageID = 0
			case chatID != 0 && messageID != 0:
				editMsgTxt.ChatID = chatID
				editMsgTxt.MessageID = messageID
				editMsgTxt.InlineMessageID = ""
			default:
				err = errors.New("Can't edit Telegram message as inlineMessageID is empty && chatID == 0 && messageID == 0")
			}
		} else if editMsgTxt.InlineMessageID != "" && editMsgTxt.ChatID != 0 && editMsgTxt.MessageID != 0 {
			panic("m.TelegramEditMessageText => InlineMessageID is NOT empty && ChatID != 0 && MessageID != 0")
		}
		chattable = editMsgTxt
	} else if editMsgMarkup := m.TelegramEditMessageMarkup; editMsgMarkup != nil {
		log.Debugf(c, "m.TelegramEditMessageMarkup != nil")
		if (editMsgMarkup.ReplyMarkup == nil || len(editMsgMarkup.ReplyMarkup.InlineKeyboard) == 0) && m.TelegramKeyboard != nil {
			if replyMarkup, ok := m.TelegramKeyboard.(*tgbotapi.InlineKeyboardMarkup); ok {
				editMsgMarkup.ReplyMarkup = replyMarkup
			} else {
				panic(fmt.Sprintf("m.TelegramKeyboard is not *tgbotapi.InlineKeyboardMarkup but %T", m.TelegramKeyboard))
			}
		}
		chattable = m.TelegramEditMessageMarkup
	} else if m.TelegramInlineConfig != nil {
		chattable = m.TelegramInlineConfig
	} else if m.Text == bots.NoMessageToSend {
		log.Debugf(c, bots.NoMessageToSend)
		return
	} else if m.IsEdit || (tgUpdate.CallbackQuery != nil && tgUpdate.CallbackQuery.InlineMessageID != "" && m.TelegramChatID == 0) {
		// Edit message
		inlineMessageID, chatID, messageID := getTgMessageIDs(tgUpdate)
		log.Debugf(c, "Edit message => inlineMessageID: %v, chatID: %d, messageID: %d", inlineMessageID, chatID, messageID)
		if inlineMessageID == "" && chatID == 0 && messageID == 0 {
			err = errors.New("Can't edit Telegram message as inlineMessageID is empty && chatID == 0 && messageID == 0")
			return
		}
		if m.Text == "" && m.TelegramKeyboard != nil {
			chattable = tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, inlineMessageID, m.TelegramKeyboard.(*tgbotapi.InlineKeyboardMarkup))
		} else if m.Text != "" {
			editMessageTextConfig := tgbotapi.NewEditMessageText(chatID, messageID, inlineMessageID, m.Text)
			editMessageTextConfig.ParseMode = parseMode()
			editMessageTextConfig.DisableWebPagePreview = m.DisableWebPagePreview
			editMessageTextConfig.ReplyMarkup = m.TelegramKeyboard.(*tgbotapi.InlineKeyboardMarkup)

			chattable = editMessageTextConfig
		} else {
			err = fmt.Errorf("can't edit telegram message as got unknown output: %v", m)
			return
		}
	} else if m.Text != "" {
		messageConfig := r.whc.NewTgMessage(m.Text)
		if m.TelegramChatID != 0 {
			messageConfig.ChatID = m.TelegramChatID
		}
		messageConfig.DisableWebPagePreview = m.DisableWebPagePreview
		messageConfig.DisableNotification = m.DisableNotification
		messageConfig.ReplyMarkup = m.TelegramKeyboard
		messageConfig.ParseMode = parseMode()

		chattable = messageConfig
	} else {
		switch r.whc.InputType() {
		case bots.WebhookInputInlineQuery: // pass
			log.Debugf(c, "No response to WebhookInputInlineQuery")
		case bots.WebhookInputChosenInlineResult: // pass
		default:
			mBytes, err := ffjson.Marshal(m)
			if err != nil {
				log.Errorf(c, "Failed to marshal MessageFromBot to JSON: %v", err)
			}
			inputTypeName := bots.WebhookInputTypeNames[r.whc.InputType()]
			log.Debugf(c, "Not inline answer, Not inline, Not edit inline, Text is empty. r.whc.InputType(): %v\nMessageFromBot:\n%v", inputTypeName, string(mBytes))
			ffjson.Pool(mBytes)
		}
		return
	}

	if jsonStr, err := ffjson.Marshal(chattable); err != nil {
		log.Errorf(c, "Failed to marshal message config to json: %v\n\tJSON: %v\n\tchattable: %v", err, jsonStr, chattable)
		ffjson.Pool(jsonStr)
		return resp, err
	} else {
		var indentedJson bytes.Buffer
		var indentedJsonStr string
		if indentedErr := json.Indent(&indentedJson, jsonStr, "", "\t"); indentedErr == nil {
			indentedJsonStr = indentedJson.String()
		} else {
			indentedJsonStr = string(jsonStr)
		}
		ffjson.Pool(jsonStr)
		//vals, err := chattable.Values()
		//if err != nil {
		//	//pass?
		//}
		//log.Debugf(c, "Sending to Telegram, Text: %v\n------------------------\nAs JSON: %v------------------------\nAs URL values: %v", m.Text, indentedJsonStr, vals.Encode())
		log.Debugf(c, "Sending to Telegram, Text: %v\n------------------------\nAs JSON: %v", m.Text, indentedJsonStr)
	}

	//if values, err := chattable.Values(); err != nil {
	//	log.Errorf(c, "Failed to marshal message config to url.Values: %v", err)
	//	return resp, err
	//} else {
	//	log.Debugf(c, "Message for sending to Telegram as URL values: %v", values)
	//}

	switch channel {
	case bots.BotApiSendMessageOverResponse:
		if _, err := tgbotapi.ReplyToResponse(chattable, r.w); err != nil {
			log.Errorf(c, "Failed to send message to Telegram throw HTTP response: %v", err)
		}
		return resp, err
	case bots.BotApiSendMessageOverHTTPS:
		botApi := tgbotapi.NewBotAPIWithClient(
			r.whc.BotContext.BotSettings.Token,
			r.whc.GetHttpClient(),
		)
		botApi.EnableDebug(c)
		if message, err := botApi.Send(chattable); err != nil {
			log.Errorf(c, errors.Wrap(err, "Failed to send message to Telegram using HTTPS API").Error())
			return resp, err
		} else {
			if message.MessageID != 0 {
				log.Debugf(c, "Telegram API: MessageID=%v", message.MessageID)
			} else {
				if messageJson, err := ffjson.Marshal(message); err != nil {
					log.Warningf(c, "Telegram API response as raw: %v", message)
					ffjson.Pool(messageJson)
				} else {
					log.Debugf(c, "Telegram API response as JSON: %v", string(messageJson))
					ffjson.Pool(messageJson)
				}
			}

			return bots.OnMessageSentResponse{TelegramMessage: message}, nil
		}
	default:
		panic(fmt.Sprintf("Unknown channel: %v", channel))
	}
}
