package bots

import (
	"time"
	"strings"
	"fmt"
	"net/url"
)

type OwnedByUser struct {
	UserID int64
	DtCreated     time.Time
	DtUpdated	  time.Time
}

func (e *OwnedByUser) GetUserID() int64 {
	return e.UserID
}

func (whc *OwnedByUser) SetDtUpdatedToNow() {
	whc.DtUpdated = time.Now()
}

type BotEntity struct {
	AccessGranted bool
	OwnedByUser
}

func (e *BotEntity) IsAccessGranted() bool {
	return e.AccessGranted
}

func (e *BotEntity) SetAccessGranted(value bool) {
	e.AccessGranted = value
}

type BotUserEntity struct {
	BotEntity

	FirstName     string // required
	LastName      string // optional
	UserName      string // optional
}

type BotChatEntity struct {
	BotEntity
	//
	Type string `datastore:",noindex"`
	Title string `datastore:",noindex"`
	//
	AwaitingReplyTo string `datastore:",noindex"`
	PreferredLanguage string `datastore:",noindex"`
}

func (e *BotChatEntity) SetDtUpdateToNow() {
	e.DtUpdated = time.Now()
}
func (e *BotChatEntity) GetAwaitingReplyTo() string {
	return e.AwaitingReplyTo
}

func (e *BotChatEntity) SetAwaitingReplyTo(value string) {
	e.AwaitingReplyTo = value
}

func (e *BotChatEntity) GetPreferredLanguage() string {
	return e.PreferredLanguage
}

func (e *BotChatEntity) SetPreferredLanguage(value string) {
	e.PreferredLanguage = value
}

func (e *BotChatEntity) IsAwaitingReplyTo(code string) bool {
	return strings.HasSuffix(e.AwaitingReplyTo, code)
}

func (e *BotChatEntity) AddStepToAwaitingReplyTo(step string) {
	pathAndQuery := strings.SplitN(e.AwaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, 2)
	if len(pathAndQuery) > 1 {
		e.SetAwaitingReplyTo(fmt.Sprintf("%v%v%v?%v", pathAndQuery[0], AWAITING_REPLY_TO_PATH_SEPARATOR, step, pathAndQuery[1]))
	} else {
		e.SetAwaitingReplyTo(e.GetAwaitingReplyTo() + AWAITING_REPLY_TO_PATH_SEPARATOR + step)
	}
}

func (e *BotChatEntity) AddWizardParam(name, value string) {
	s := fmt.Sprintf("%v=%v", name, url.QueryEscape(value))
	awaitignReplyTo := e.GetAwaitingReplyTo()
	if strings.Contains(awaitignReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR) {
		e.SetAwaitingReplyTo(awaitignReplyTo + AWAITING_REPLY_TO_PARAMS_SEPARATOR + s)
	} else {
		e.SetAwaitingReplyTo(awaitignReplyTo + AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR + s)
	}

}
