package botmsg

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var _ BotMessage = AnswerCallbackQuery{}
var _ BotMessage = (*AnswerCallbackQuery)(nil)

// AnswerCallbackQuery is modeled after https://core.telegram.org/bots/api#answercallbackquery
type AnswerCallbackQuery struct {
	CallbackQueryID string `json:"callback_query_id"` // Unique identifier for the query to be answered

	// Text of the notification. If not specified, nothing will be shown to the user, 0-200 characters
	Text string `json:"text,omitempty"`

	// If True, an alert will be shown by the client instead of a notification at the top of the chat screen
	ShowAlert bool `json:"show_alert,omitempty"`

	// URL that will be opened by the user's client.
	// If you have created a Game and accepted the conditions via @BotFather,
	// specify the URL that opens your game - note that this will only work
	// if the query comes from a callback_game button.
	//
	// Otherwise, you may use links like t.me/your_bot?start=XXXX that open your bot with a parameter.
	URL string `json:"url,omitempty"`

	// The maximum amount of time in seconds that the result of the callback query may be cached client-side.
	// Telegram apps will support caching starting in version 3.14. Defaults to 0.
	CacheTime int `json:"cache_time,omitempty"`
}

func (AnswerCallbackQuery) BotMessageType() Type {
	return TypeCallbackAnswer
}

func (v AnswerCallbackQuery) Validate() error {
	if strings.TrimSpace(v.CallbackQueryID) == "" {
		return errors.New("missing required parameter CallbackQueryID")
	}
	if len(v.Text) > 200 {
		return fmt.Errorf("callback text cannot be longer than 200 characters, got %d", len(v.Text))
	}
	if v.CacheTime < 0 {
		return fmt.Errorf("callback cache time must be 0 or positive, got %d", v.CacheTime)
	}
	if _, err := url.Parse(v.URL); err != nil {
		return fmt.Errorf("URL is invalid: %w", err)
	}
	return nil
}
