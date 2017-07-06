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
	botApi := tgbotapi.NewBotAPIWithClient(
		r.whc.BotContext.BotSettings.Token,
		r.whc.GetHttpClient(),
	)

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

	botApi.EnableDebug(c)
	if m.TelegramCallbackAnswer != nil {
		log.Debugf(c, "Inline answer")
		m.TelegramCallbackAnswer.CallbackQueryID = tgUpdate.CallbackQuery.ID

		chattable = m.TelegramCallbackAnswer
		jsonStr, err := json.Marshal(chattable)
		if err == nil {
			log.Infof(c, "CallbackAnswer for sending to Telegram: %v", string(jsonStr))
		} else {
			log.Errorf(c, "Failed to marshal message config to json: %v\n\tInput: %v", err, chattable)
		}
		apiResponse, err := botApi.Send(chattable)

		if s, err2 := json.Marshal(apiResponse); err2 != nil {
			log.Debugf(c, "apiResponse: %v", s)
		}
		return resp, err
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
		case bots.WebhookInputChosenInlineResult: // pass
		default:
			mBytes, err := json.Marshal(m)
			if err != nil {
				log.Errorf(c, "Failed to marshal MessageFromBot to JSON: %v", err)
			}
			inputTypeName := bots.WebhookInputTypeNames[r.whc.InputType()]
			log.Warningf(c, "Not inline answer, Not inline, Not edit inline, Text is empty. r.whc.InputType(): %v\nMessageFromBot:\n%v", inputTypeName, string(mBytes))
		}
		return
	}

	if jsonStr, err := json.Marshal(chattable); err != nil {
		log.Errorf(c, "Failed to marshal message config to json: %v\n\tJSON: %v\n\tchattable: %v", err, jsonStr, chattable)
		return resp, err
	} else {
		var indentedJson bytes.Buffer
		var indentedJsonStr string
		if indentedErr := json.Indent(&indentedJson, jsonStr, "", "\t"); indentedErr == nil {
			indentedJsonStr = indentedJson.String()
		} else {
			indentedJsonStr = string(jsonStr)
		}
		//vals, err := chattable.Values()
		//if err != nil {
		//	//pass?
		//}
		//log.Infof(c, "Sending to Telegram, Text: %v\n------------------------\nAs JSON: %v------------------------\nAs URL values: %v", m.Text, indentedJsonStr, vals.Encode())
		log.Infof(c, "Sending to Telegram, Text: %v\n------------------------\nAs JSON: %v", m.Text, indentedJsonStr)
	}

	//if values, err := chattable.Values(); err != nil {
	//	log.Errorf(c, "Failed to marshal message config to url.Values: %v", err)
	//	return resp, err
	//} else {
	//	log.Infof(c, "Message for sending to Telegram as URL values: %v", values)
	//}

	switch channel {
	case bots.BotApiSendMessageOverResponse:
		if _, err := tgbotapi.ReplyToResponse(chattable, r.w); err != nil {
			log.Errorf(c, "Failed to send message to Telegram throw HTTP response: %v", err)
		}
		return resp, err
	case bots.BotApiSendMessageOverHTTPS:
		if message, err := botApi.Send(chattable); err != nil {
			log.Errorf(c, errors.Wrap(err, "Failed to send message to Telegram using HTTPS API").Error())
			return resp, err
		} else {
			log.Debugf(c, "Telegram API: MessageID=%v", message.MessageID)
			//if messageJson, err := json.Marshal(message); err != nil {
			//	log.Warningf(c, "Telegram API response as raw: %v", message)
			//} else {
			//	log.Debugf(c, "Telegram API: MessageID=%v", message.MessageID)
			//	log.Debugf(c, "Telegram API response as JSON: %v", string(messageJson))
			//}
			return bots.OnMessageSentResponse{TelegramMessage: message}, nil
		}
	default:
		panic(fmt.Sprintf("Unknown channel: %v", channel))
	}
}
