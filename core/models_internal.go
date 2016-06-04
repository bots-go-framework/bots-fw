package bots

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type OwnedByUser struct {
	AppUserID int64 // TODO: Rename to AppUserID?
	DtCreated time.Time
	DtUpdated time.Time
}

func (e *OwnedByUser) GetAppUserID() int64 {
	return e.AppUserID
}

func (e *OwnedByUser) SetAppUserID(appUserID int64) {
	e.AppUserID = appUserID
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

	FirstName string // required
	LastName  string // optional
	UserName  string // optional
}

type BotChatEntity struct {
	BotEntity
	//
	Type  string `datastore:",noindex"`
	Title string `datastore:",noindex"`
	//
	AwaitingReplyTo   string `datastore:",noindex"`
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

func (e *BotChatEntity) IsAwaitingReplyTo(code string, logger Logger) bool {
	awaitingReplyToPath := e.getAwaitingReplyToPath()
	logger.Debugf("IsAwaitingReplyTo(%v), awaitingReplyToPath: %v", code, awaitingReplyToPath)
	return awaitingReplyToPath == code || strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR+code)
}

func (e *BotChatEntity) getAwaitingReplyToPath() string {
	pathAndQuery := strings.SplitN(e.AwaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, 2)
	if len(pathAndQuery) > 1 {
		return pathAndQuery[0]
	}
	return e.AwaitingReplyTo
}

func (e *BotChatEntity) AddStepToAwaitingReplyTo(step string) {
	awaitingReplyTo := e.AwaitingReplyTo
	pathAndQuery := strings.SplitN(awaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, 2)
	if len(pathAndQuery) > 1 {
		path := pathAndQuery[0]
		suffix := AWAITING_REPLY_TO_PATH_SEPARATOR + step
		if !strings.HasSuffix(path, suffix) {
			e.SetAwaitingReplyTo(fmt.Sprintf("%v%v?%v", path, suffix, pathAndQuery[1]))
		}
	} else {
		e.SetAwaitingReplyTo(awaitingReplyTo + AWAITING_REPLY_TO_PATH_SEPARATOR + step)
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
