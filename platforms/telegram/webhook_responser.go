package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"encoding/json"
	"github.com/strongo/bots-api-telegram"
	"net/http"
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
	log := r.whc.GetLogger()
	messageConfig := r.whc.NewTgMessage(m.Text)
	switch m.Format {
	case bots.MessageFormatHTML:
		messageConfig.ParseMode = "HTML"
	case bots.MessageFormatMarkdown:
		messageConfig.ParseMode = "Markdown"
	}
	if m.Keyboard.HideKeyboard {
		if len(m.Keyboard.Buttons) > 0 {
			log.Errorf("Got both 'HideKeyboard=true' & len(m.Keyboard.Buttons):%v > 0. Buttons: %v", len(m.Keyboard.Buttons), m.Keyboard.Buttons)
		}
		messageConfig.ReplyMarkup = tgbotapi.ReplyKeyboardHide{HideKeyboard: m.Keyboard.HideKeyboard, Selective: m.Keyboard.Selective}
	} else if m.Keyboard.ForceReply {
		if len(m.Keyboard.Buttons) > 0 {
			log.Errorf("Got both 'ForceReply=true' & len(m.Keyboard.Buttons):%v > 0. Buttons: %v", len(m.Keyboard.Buttons), m.Keyboard.Buttons)
		}
		messageConfig.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: m.Keyboard.Selective}
	} else if len(m.Keyboard.Buttons) > 0 {
		messageConfig.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:        m.Keyboard.Buttons,
			ResizeKeyboard:  m.Keyboard.ResizeKeyboard,
			OneTimeKeyboard: m.Keyboard.OneTimeKeyboard,
			Selective:       m.Keyboard.Selective,
		}
	}
	mcJson, err := json.Marshal(messageConfig)
	if err == nil {
		log.Infof("Message for sending to Telegram: %v", string(mcJson))
	} else {
		log.Errorf("Failed to marshal message config to json: %v\n\tInput: %v", err, messageConfig)
	}
	s, err := messageConfig.ReplyToResponse(r.w)
	log.Infof("Sending to Telegram: %v", s)
	return err
}


