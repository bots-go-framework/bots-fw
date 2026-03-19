package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/i18n"
	"net/http"
)

// WebhookInlineQueryContext provides context for inline query
// Deprecated: not used; will be removed in a future version.
type WebhookInlineQueryContext interface {
}

// ExecutionContext wraps Context() and adds no independent value.
// Deprecated: use WebhookRequestContext directly.
type ExecutionContext interface {
	Context() context.Context
}

// --- Sub-interfaces ---

// WebhookRequestContext provides identity and infrastructure access for the current request.
type WebhookRequestContext interface {
	// Context returns the Go context for this request.
	Context() context.Context

	// SetContext replaces the request context (e.g. after adding values or a deadline).
	SetContext(c context.Context)

	// Request returns the raw HTTP request.
	Request() *http.Request

	// Environment returns the deployment environment (e.g. "local", "production").
	Environment() string

	// BotPlatform returns the platform this request arrived on (Telegram, Viber, FBM, …).
	BotPlatform() BotPlatform

	// BotContext returns settings and host information for the current bot.
	BotContext() BotContext

	// GetBotCode is a convenience shortcut for BotContext().BotSettings.Code.
	GetBotCode() string

	// GetBotSettings is a convenience shortcut for BotContext().BotSettings.
	GetBotSettings() *BotSettings

	// DB returns the database handle assigned to this bot.
	DB() dal.DB

	// AppContext returns application-level context (i18n, DAL, etc.).
	AppContext() AppContext

	// ExecutionContext returns the execution context.
	// Deprecated: use Context() directly.
	ExecutionContext() ExecutionContext
}

// WebhookInputContext provides access to the incoming message from the user.
type WebhookInputContext interface {
	BotInputProvider

	// GetBotUserID returns the platform-specific user ID of the sender as a string.
	GetBotUserID() string

	// MustBotChatID returns the chat ID or panics if it cannot be determined.
	MustBotChatID() string

	// IsInGroup reports whether the message was received in a group chat.
	IsInGroup() (bool, error)
}

// WebhookUserData provides read/write access to the persistent state of the current
// bot user, app user, and chat.
type WebhookUserData interface {
	// ChatData returns the current bot chat's persistent data.
	// Returns nil for input types that have no associated chat (e.g. InlineQuery).
	ChatData() botsfwmodels.BotChatData

	// SaveBotChat persists the current chat data to the database.
	SaveBotChat() error

	// GetBotUser returns the current platform user record.
	GetBotUser() (botUser botsdal.BotUser, err error)

	// GetBotUserForUpdate returns the platform user record inside a write transaction.
	GetBotUserForUpdate(ctx context.Context, tx dal.ReadwriteTransaction) (botUser botsdal.BotUser, err error)

	// AppUserID returns the application-layer user ID linked to this bot user.
	AppUserID() string

	// SetUser caches the resolved app user ID and data into the context.
	SetUser(id string, data botsfwmodels.AppUserData)

	// AppUserData loads and returns the app user's persistent data.
	AppUserData() (botsfwmodels.AppUserData, error)

	// RecordsFieldsSetter returns the helper used to populate new bot/chat/user records.
	RecordsFieldsSetter() BotRecordsFieldsSetter

	// UpdateLastProcessed records the message sequence number / timestamp on the chat entity.
	UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error

	// IsNewerThen reports whether the current message is newer than the chat entity's
	// last-processed sequence number (used to detect and discard duplicate deliveries).
	IsNewerThen(chatEntity botsfwmodels.BotChatData) bool
}

// WebhookI18n provides localisation support for the current request.
type WebhookI18n interface {
	i18n.SingleLocaleTranslator

	// SetLocale switches the active locale for this request.
	SetLocale(code5 string) error

	// GetTranslator returns a translator pinned to the given locale code.
	GetTranslator(locale string) i18n.SingleLocaleTranslator

	// CommandText formats a command title and icon into a display string.
	CommandText(title, icon string) string
}

// WebhookMessaging provides helpers to construct and send messages back to the user.
type WebhookMessaging interface {
	// NewMessage creates a plain-text MessageFromBot.
	NewMessage(text string) botmsg.MessageFromBot

	// NewMessageByCode creates a MessageFromBot from an i18n key, formatting it with args.
	NewMessageByCode(messageCode string, a ...interface{}) botmsg.MessageFromBot

	// NewEditMessage creates a MessageFromBot that edits the previously sent message.
	NewEditMessage(text string, format botmsg.Format) (botmsg.MessageFromBot, error)

	// Responder returns the WebhookResponder used to deliver messages to the platform.
	Responder() WebhookResponder
}

// WebhookTelemetry provides access to the analytics pipeline.
type WebhookTelemetry interface {
	Analytics() WebhookAnalytics
}

// --- Composed interface ---

// WebhookContext is the full request context passed to every command action handler.
// It is a composition of focused sub-interfaces. Prefer accepting the narrowest
// sub-interface that covers your function's actual needs.
type WebhookContext interface {
	WebhookRequestContext
	WebhookInputContext
	WebhookUserData
	WebhookI18n
	WebhookMessaging
	WebhookTelemetry
}

// BotState provides state of the bot.
// Deprecated: use WebhookUserData.IsNewerThen instead.
type BotState interface {
	IsNewerThen(chatEntity botsfwmodels.BotChatData) bool
}

// BotInputProvider provides an input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
type BotInputProvider interface {
	// Input returns a webhook input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
	Input() botinput.InputMessage
}
