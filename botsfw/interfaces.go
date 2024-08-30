package botsfw

import (
	"context"
	"net/http"
)

// BotPlatform describes current bot platform
type BotPlatform interface { // TODO: Change to a struct

	// ID returns bot platform ID like 'telegram', 'fbmessenger', 'viber', etc.
	ID() string

	// Version returns a version of a bot platform adapter. It is used for debugging purposes.
	Version() string
}

// BotHost describes current bot app host environment
type BotHost interface {

	// Context returns a context.Context for a request. We need this as some platforms (as Google App Engine Standard)
	// require usage of a context with a specific wrapper
	Context(r *http.Request) context.Context

	// GetHTTPClient returns HTTP client for current host
	// We need this as some platforms (as Google App Engine Standard) require setting http client in a specific way.
	GetHTTPClient(c context.Context) *http.Client
}

// NewBotContext creates current bot host & settings
func NewBotContext(botHost BotHost, botSettings *BotSettings) *BotContext {
	if botHost == nil {
		panic("required argument botHost is nil")
	}
	if botSettings == nil {
		panic("required argument botSettings is nil")
	}
	if botSettings.Code == "" {
		panic("ReferredTo botSettings.Code is empty string")
	}
	return &BotContext{
		BotHost:     botHost,
		BotSettings: botSettings,
	}
}

// MessengerResponse represents response from a messenger
type MessengerResponse interface {
}

// OnMessageSentResponse represents response on message sent event
type OnMessageSentResponse struct {
	StatusCode      int
	TelegramMessage MessengerResponse // TODO: change to some interface
}

// WebhookResponder is an API provider to send messages through a messenger
type WebhookResponder interface {
	SendMessage(c context.Context, m MessageFromBot, channel BotAPISendMessageChannel) (OnMessageSentResponse, error)
}

// InputMessage represents single input message
type InputMessage interface {
	Text() string
}

// BotAPISendMessageChannel specifies messenger channel
type BotAPISendMessageChannel string

const (
	// BotAPISendMessageOverHTTPS indicates message should be sent over HTTPS
	BotAPISendMessageOverHTTPS = BotAPISendMessageChannel("https")

	// BotAPISendMessageOverResponse indicates message should be sent in HTTP response
	BotAPISendMessageOverResponse = BotAPISendMessageChannel("response")
)
