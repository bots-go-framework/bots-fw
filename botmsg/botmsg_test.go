package botmsg_test

import (
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

func TestAnswerCallbackQuery_BotMessageType(t *testing.T) {
	if got := (botmsg.AnswerCallbackQuery{}).BotMessageType(); got != botmsg.TypeCallbackAnswer {
		t.Errorf("BotMessageType() = %v, want %v", got, botmsg.TypeCallbackAnswer)
	}
}

func TestAnswerCallbackQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   botmsg.AnswerCallbackQuery
		wantErr bool
	}{
		{name: "valid_minimal", query: botmsg.AnswerCallbackQuery{CallbackQueryID: "q1"}},
		{name: "valid_full", query: botmsg.AnswerCallbackQuery{CallbackQueryID: "q1", Text: "ok", CacheTime: 5, URL: "https://t.me/bot?start=x"}},
		{name: "empty_id", query: botmsg.AnswerCallbackQuery{}, wantErr: true},
		{name: "text_too_long", query: botmsg.AnswerCallbackQuery{CallbackQueryID: "q1", Text: strings.Repeat("x", 201)}, wantErr: true},
		{name: "negative_cache_time", query: botmsg.AnswerCallbackQuery{CallbackQueryID: "q1", CacheTime: -1}, wantErr: true},
		{name: "invalid_url", query: botmsg.AnswerCallbackQuery{CallbackQueryID: "q1", URL: "\x7f"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatIntID_ChatUID(t *testing.T) {
	if got := botmsg.ChatIntID(42).ChatUID(); got != "42" {
		t.Errorf("ChatUID() = %q, want %q", got, "42")
	}
}

func TestTextMessageFromBot(t *testing.T) {
	m := &botmsg.TextMessageFromBot{}

	if got := m.BotEndpoint(); got != "sendMessage" {
		t.Errorf("BotEndpoint() = %q, want %q", got, "sendMessage")
	}

	if got := m.BotMessageType(); got != botmsg.TypeText {
		t.Errorf("BotMessageType() = %v, want %v", got, botmsg.TypeText)
	}

	m.IsEdit = true
	if got := m.BotMessageType(); got != botmsg.TypeEditMessage {
		t.Errorf("BotMessageType() (IsEdit) = %v, want %v", got, botmsg.TypeEditMessage)
	}
}
