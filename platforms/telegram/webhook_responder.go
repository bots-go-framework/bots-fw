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
			return  "Markdown"
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
	} else if m.TelegramEditMessageText != nil {
		if m.TelegramEditMessageText.ReplyMarkup == nil && m.TelegramKeyboard != nil {
			m.TelegramEditMessageText.ReplyMarkup = m.TelegramKeyboard.(*tgbotapi.InlineKeyboardMarkup)
		}
		chattable = m.TelegramEditMessageText
	} else  if m.TelegramEditMessageMarkup != nil {
		chattable = m.TelegramEditMessageMarkup
	} else if m.TelegramInlineConfig != nil {
		chattable = m.TelegramInlineConfig
	} else if m.Text != "" {
		if tgUpdate.CallbackQuery != nil && tgUpdate.CallbackQuery.InlineMessageID != "" && m.TelegramChatID == 0 {
			editMessageTextConfig := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{InlineMessageID: tgUpdate.CallbackQuery.InlineMessageID},
				Text: m.Text,
				ParseMode: parseMode(),
				DisableWebPagePreview: m.DisableWebPagePreview,
			}
			editMessageTextConfig.ReplyMarkup, _ = m.TelegramKeyboard.(*tgbotapi.InlineKeyboardMarkup)
			chattable = editMessageTextConfig
		} else {
			if m.Text == bots.NoMessageToSend {
				return
			}
			messageConfig := r.whc.NewTgMessage(m.Text)
			if m.TelegramChatID != 0 {
				messageConfig.ChatID = m.TelegramChatID
			}
			messageConfig.DisableWebPagePreview = m.DisableWebPagePreview
			messageConfig.DisableNotification = m.DisableNotification
			messageConfig.ReplyMarkup = m.TelegramKeyboard
			messageConfig.ParseMode = parseMode()

			chattable = messageConfig
		}
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
			log.Debugf(c, "Telegram API: MessageID=%v", message.MessageID)
			//if messageJson, err := ffjson.Marshal(message); err != nil {
			//	log.Warningf(c, "Telegram API response as raw: %v", message)
			// messageJson.Pool(messageJson)
			//} else {
			//	log.Debugf(c, "Telegram API: MessageID=%v", message.MessageID)
			//	log.Debugf(c, "Telegram API response as JSON: %v", string(messageJson))
			// messageJson.Pool(messageJson)
			//}
			return bots.OnMessageSentResponse{TelegramMessage: message}, nil
		}
	default:
		panic(fmt.Sprintf("Unknown channel: %v", channel))
	}
}
