package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"encoding/json"
	"net/http"
	"github.com/strongo/bots-api-telegram"
	"errors"
	"fmt"
)

type TelegramWebhookResponder struct {
	w http.ResponseWriter
	whc *TelegramWebhookContext
}

var _ bots.WebhookResponder = (*TelegramWebhookResponder)(nil)

func NewTelegramWebhookResponder(w http.ResponseWriter, whc *TelegramWebhookContext) TelegramWebhookResponder {
	return TelegramWebhookResponder{w: w, whc: whc}
}


func (r TelegramWebhookResponder) SendMessage(m bots.MessageFromBot) error {
	//ctx := tc.Context()
	logger := r.whc.GetLogger()

	var chattable tgbotapi.Chattable
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

		botApi := &tgbotapi.BotAPI{
			Token: r.whc.BotContext.BotSettings.Token,
			Debug: true,
			Client: r.whc.GetHttpClient(),
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
		s, err := tgbotapi.ReplyToResponse(chattable, r.w)
		logger.Debugf("Sent to response: %v", s)
		return err
	}
	return nil
}


