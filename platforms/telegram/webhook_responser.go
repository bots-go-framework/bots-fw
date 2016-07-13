package telegram_bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/http"
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

func (r TelegramWebhookResponder) SendMessage(m bots.MessageFromBot, channel bots.BotApiSendMessageChannel) error {
	if channel != bots.BotApiSendMessageOverHTTPS && channel != bots.BotApiSendMessageOverResponse {
		panic(fmt.Sprintf("Unknown channel: [%v]. Expected either 'https' or 'response'.", channel))
	}
	//ctx := tc.Context()
	logger := r.whc.GetLogger()

	var chattable tgbotapi.Chattable
	botApi := &tgbotapi.BotAPI{
		Token:  r.whc.BotContext.BotSettings.Token,
		Debug:  true,
		Client: r.whc.GetHttpClient(),
	}
	if m.TelegramInlineAnswer != nil {
		logger.Debugf("Inline answer")
		chattable = m.TelegramInlineAnswer
		inlineAnswer := *m.TelegramInlineAnswer
		input, ok := r.whc.WebhookInput.(TelegramWebhookInput)
		if !ok {
			return errors.New(fmt.Sprintf("Expected TelegramWebhookInput, got %T", r.whc.WebhookInput))
		}
		inlineAnswer.InlineQueryID = input.update.InlineQuery.ID

		jsonStr, err := json.Marshal(inlineAnswer)
		if err == nil {
			logger.Infof("Message for sending to Telegram: %v", string(jsonStr))
		} else {
			logger.Errorf("Failed to marshal message config to json: %v\n\tInput: %v", err, inlineAnswer)
		}

		apiResponse, err := botApi.AnswerInlineQuery(inlineAnswer)

		if err != nil {
			s, err := json.Marshal(apiResponse)
			if err != nil {
				logger.Debugf("apiResponse: %v", s)
			}
		}
		return err
	} else if m.Text != "" {
		logger.Debugf("Not inline answer")
		messageConfig := r.whc.NewTgMessage(m.Text)
		messageConfig.DisableWebPagePreview = m.DisableWebPagePreview
		switch m.Format {
		case bots.MessageFormatHTML:
			messageConfig.ParseMode = "HTML"
		case bots.MessageFormatMarkdown:
			messageConfig.ParseMode = "Markdown"
		}
		messageConfig.ReplyMarkup = m.TelegramKeyboard

		chattable = messageConfig

		jsonStr, err := json.Marshal(chattable)
		if err == nil {
			logger.Infof("Message for sending to Telegram: %v", string(jsonStr))
		} else {
			logger.Errorf("Failed to marshal message config to json: %v\n\tInput: %v", err, jsonStr)
		}
		switch channel {
		case bots.BotApiSendMessageOverResponse:
			_, err := tgbotapi.ReplyToResponse(chattable, r.w)
			//logger.Debugf("Sent to response: %v", s)
			return err
		case bots.BotApiSendMessageOverHTTPS:
			if _, err := botApi.Send(chattable); err != nil {
				logger.Errorf("Failed to send message to Telegram using HTTPS API: %v, %v", err)
			}
			return err
		}
	} else {
		logger.Warningf("Not inline and text is empty.")
	}
	return nil
}
