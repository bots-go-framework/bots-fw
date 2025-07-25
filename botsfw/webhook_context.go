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

// WebhookInlineQueryContext provides context for inline query (TODO: check & document)
type WebhookInlineQueryContext interface {
}

// ExecutionContext TODO: either specify clear purpose and added value or remove
type ExecutionContext interface {
	Context() context.Context
}

// WebhookContext provides context for current request from user to bot
type WebhookContext interface { // TODO: Make interface much smaller?
	//dal.TransactionCoordinator
	Environment() string
	BotInputProvider
	BotPlatform() BotPlatform

	Request() *http.Request

	//Init(w http.ResponseWriter, r *http.Request) error

	// Context return context
	Context() context.Context

	// SetContext sets context
	SetContext(c context.Context)

	ExecutionContext() ExecutionContext

	AppContext() AppContext

	BotContext() BotContext

	MustBotChatID() string

	// GetBotCode returns bot code. This is a shortcut to BotSettings().Code
	GetBotCode() string

	// GetBotToken returns bot token. This is a shortcut to BotSettings().Token
	// Deprecated: use BotSettings().Token instead
	//GetBotToken() string

	// GetBotSettings returns bot settings
	GetBotSettings() *BotSettings

	// DB is a reference to database used to store data of current bot
	DB() dal.DB

	// Tx is a reference to database transaction used to get/save data of current bot
	//Tx() dal.ReadwriteTransaction

	// ChatData returns data of current bot chat without ID/Key
	ChatData() botsfwmodels.BotChatData // Formerly ChatEntity()

	// BotUser returns record of current bot user
	GetBotUser() (botUser botsdal.BotUser, err error)
	GetBotUserForUpdate(ctx context.Context, tx dal.ReadwriteTransaction) (botUser botsdal.BotUser, err error)

	GetBotUserID() string

	// IsInGroup indicates if message was received in a group botChat
	IsInGroup() (bool, error) // We  need to return an error as well (for Telegram chat instance).

	// CommandText TODO: needs to be documented
	CommandText(title, icon string) string

	//DefaultLocale() strongoapp.ByLocale

	// SetLocale sets Locale for current session
	SetLocale(code5 string) error

	NewMessage(text string) botmsg.MessageFromBot
	NewMessageByCode(messageCode string, a ...interface{}) botmsg.MessageFromBot
	NewEditMessage(text string, format botmsg.Format) (botmsg.MessageFromBot, error)
	//NewEditMessageKeyboard(kbMarkup tgbotapi.InlineKeyboardMarkup) MessageFromBot

	UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error

	AppUserID() string
	SetUser(id string, data botsfwmodels.AppUserData)

	// AppUserInt64ID Deprecate: use AppUserID() instead
	//AppUserInt64ID() int64

	AppUserData() (botsfwmodels.AppUserData, error)
	//SaveAppUser(appUserID int64, appUserEntity BotAppUser) error

	BotState

	//Store() botsfwdal.DataAccess

	// SaveBotChat // It is dangerous to allow user to pass context to this func as if it's a transactional context it might lead to deadlock
	// Previously: takes context as we might want to add timeout or cancellation or something else.
	SaveBotChat() error

	//RecordsMaker() botsfwmodels.BotRecordsMaker

	// RecordsFieldsSetter returns a helper that sets fields of bot related records
	RecordsFieldsSetter() BotRecordsFieldsSetter

	//botinput.InputMessage // TODO: Should be removed!!!
	i18n.SingleLocaleTranslator
	GetTranslator(locale string) i18n.SingleLocaleTranslator

	Responder() WebhookResponder

	Analytics() WebhookAnalytics
}

// BotState provides state of the bot (TODO: document how is used)
type BotState interface {
	IsNewerThen(chatEntity botsfwmodels.BotChatData) bool
}

// BotInputProvider provides an input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
type BotInputProvider interface {
	// Input returns a webhook input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
	Input() botinput.InputMessage
}
