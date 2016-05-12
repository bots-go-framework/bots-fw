package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"encoding/json"
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
	case bots.MessageFormatText:
		// messageConfig.ParseMode = ""
	case bots.MessageFormatHTML:
		messageConfig.ParseMode = "HTML"
	case bots.MessageFormatMarkdown:
		messageConfig.ParseMode = "Markdown"
	}
	messageConfig.ReplyMarkup = m.TelegramKeyboard
	//if hideKeyboard, ok := m.TelegramKeyboard.(tgbotapi.ReplyKeyboardHide); ok {
	//	messageConfig.ReplyMarkup = hideKeyboard
	//} else if forceReply, ok := m.TelegramKeyboard.(bots.ForceReply); ok {
	//	messageConfig.ReplyMarkup = forceReply
	//} else if markup, ok := m.TelegramKeyboard.(bots.ReplyKeyboardMarkup); ok {
	//	buttons := make([][]string, len(markup.Buttons))
	//	for i, sourceRow := range markup.Buttons {
	//		destRow := make([]string, len(sourceRow))
	//		for j, srcBtn := range sourceRow {
	//			destRow[j] = srcBtn.Text
	//		}
	//		buttons[i] = destRow
	//	}
	//	messageConfig.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
	//		Keyboard: buttons,
	//		ResizeKeyboard:  markup.ResizeKeyboard,
	//		OneTimeKeyboard: markup.OneTimeKeyboard,
	//		Selective:       markup.Selective,
	//	}
	//} else if inline, ok := m.Keyboard.(tgbotapi.InlineKeyboardMarkup); ok {
	//	messageConfig.ReplyMarkup = inline
	//}
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


