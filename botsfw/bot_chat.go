package botsfw

import (
	"time"

	"github.com/satori/go.uuid"
	"github.com/strongo/app/user"
)

// BotChat provides data about bot chat
type BotChat interface {
	GetBotID() string
	SetBotID(botID string)

	GetAppUserIntID() int64
	SetAppUserIntID(id int64)

	AddClientLanguage(languageCode string) (changed bool)

	/*
		GetBotUserIntID() int
		GetBotUserStringID() string
	*/

	SetBotUserID(id interface{})
	SetIsGroupChat(bool)

	IsAccessGranted() bool
	IsGroupChat() bool
	SetAccessGranted(value bool) bool

	GetPreferredLanguage() string
	SetPreferredLanguage(value string)

	user.UpdatedTimeSetter
	SetDtLastInteraction(time time.Time)

	GetAwaitingReplyTo() string
	SetAwaitingReplyTo(path string)
	IsAwaitingReplyTo(code string) bool
	AddWizardParam(key, value string)
	GetWizardParam(key string) string
	PopStepsFromAwaitingReplyUpToSpecificParent(code string)
	PushStepToAwaitingReplyTo(code string)
	GetGaClientID() uuid.UUID
}

// NewChatID create a new bot chat ID, returns string
func NewChatID(botID, botChatID string) string {
	return botID + ":" + botChatID
}
