package bots

import (
	"fmt"
	"github.com/satori/go.uuid"
	"net/url"
	"strings"
	"time"
	"github.com/strongo/app/user"
)


type BotEntity struct {
	AccessGranted bool
	user.OwnedByUser
}

func (e *BotEntity) IsAccessGranted() bool {
	return e.AccessGranted
}

func (e *BotEntity) SetAccessGranted(value bool) bool {
	if e.AccessGranted != value {
		e.AccessGranted = value
		return true
	}
	return false
}

type BotUserEntity struct {
	BotEntity
	user.LastLogin

	FirstName string // required
	LastName  string // optional
	UserName  string // optional
}

type BotChatEntity struct {
	BotEntity
	BotID string `datastore:",noindex"`
	//
	IsGroup bool `datastore:",noindex"`
	Type    string `datastore:",noindex"`
	Title   string `datastore:",noindex"`
	//
	AwaitingReplyTo   string `datastore:",noindex"`
	PreferredLanguage string `datastore:",noindex"`
	GaClientID        []byte `datastore:",noindex"`
	DtLastInteraction time.Time
	InteractionsCount int
	DtForbidden       time.Time
	DtForbiddenLast   time.Time `datastore:",noindex"`
	LanguageCodes     []string  `datastore:",noindex"` // UI languages
}

var _ BotChat = (*BotChatEntity)(nil)

func (e *BotChatEntity) GetBotID() string {
	return e.BotID
}

func (e *BotChatEntity) IsGroupChat() bool {
	return e.IsGroup
}

func (e *BotChatEntity) SetIsGroupChat(v bool) {
	e.IsGroup = v
}

func (e *BotChatEntity) SetBotID(botID string) {
	e.BotID = botID
}

func (e *BotChatEntity) AddClientLanguage(languageCode string) (changed bool) {
	if languageCode == "" || languageCode == "root" {
		return false
	}
	for _, lc := range e.LanguageCodes {
		if lc == languageCode {
			return false
		}
	}
	e.LanguageCodes = append(e.LanguageCodes, languageCode)
	return false
}

//func (e *BotChatEntity) GetBotUserIntID() int {
//	panic("Should be overwritted in subclass")
//}
//
//func (e *BotChatEntity) GetBotUserStringID() string {
//	panic("Should be overwritted in subclass")
//}

func (e *BotChatEntity) SetBotUserID(id interface{}) {
	panic(fmt.Sprintf("Should be overwritted in subclass, got: %T=%v", id, id))
}

func (e *BotChatEntity) SetDtLastInteraction(v time.Time) {
	e.DtLastInteraction = v
	e.InteractionsCount += 1
}

func (e *BotChatEntity) GetGaClientID() uuid.UUID {
	var v uuid.UUID
	var err error
	if len(e.GaClientID) == 0 {
		v = uuid.NewV4()
		e.GaClientID = v.Bytes()
	} else if v, err = uuid.FromBytes(e.GaClientID); err != nil {
		panic(fmt.Sprintf("Failed to create UUID from bytes: len(%v)=%v", e.GaClientID, len(e.GaClientID)))
	}
	return v
}

func (e *BotChatEntity) SetDtUpdateToNow() {
	e.DtUpdated = time.Now()
}
func (e *BotChatEntity) GetAwaitingReplyTo() string {
	return e.AwaitingReplyTo
}

func (e *BotChatEntity) SetAwaitingReplyTo(value string) {
	e.AwaitingReplyTo = strings.TrimLeft(value, "/")
}

func (e *BotChatEntity) GetPreferredLanguage() string {
	return e.PreferredLanguage
}

func (e *BotChatEntity) SetPreferredLanguage(value string) {
	e.PreferredLanguage = value
}

func (e *BotChatEntity) IsAwaitingReplyTo(code string) bool {
	awaitingReplyToPath := e.getAwaitingReplyToPath()
	return awaitingReplyToPath == code || strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR+code)
}

func (e *BotChatEntity) getAwaitingReplyToPath() string {
	pathAndQuery := strings.SplitN(e.AwaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, 2)
	if len(pathAndQuery) > 1 {
		return pathAndQuery[0]
	}
	return e.AwaitingReplyTo
}

func (e *BotChatEntity) PopStepsFromAwaitingReplyUpToSpecificParent(step string) {
	awaitingReplyTo := e.AwaitingReplyTo
	pathAndQuery := strings.SplitN(awaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, 2)
	path := pathAndQuery[0]
	steps := strings.Split(path, AWAITING_REPLY_TO_PATH_SEPARATOR)
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i] == step {
			if i < len(steps)-1 {
				path = strings.Join(steps[:i+1], AWAITING_REPLY_TO_PATH_SEPARATOR)
				if len(pathAndQuery) > 1 {
					query := pathAndQuery[1]
					e.SetAwaitingReplyTo(path + AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR + query)
				} else {
					e.SetAwaitingReplyTo(path)
				}
			}
			steps = steps[:i]
			break
			//} else {
			//log.Infof(c, "steps[%v]: %v != %v:", i, steps[i], step)
		}
	}
}

func (e *BotChatEntity) PushStepToAwaitingReplyTo(step string) {
	awaitingReplyTo := e.AwaitingReplyTo
	pathAndQuery := strings.SplitN(awaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, 2)
	if len(pathAndQuery) > 1 { // Has query part - something after "?" character
		if !e.IsAwaitingReplyTo(step) {
			path := pathAndQuery[0]
			query := pathAndQuery[1]
			awaitingReplyTo = strings.Join([]string{path, AWAITING_REPLY_TO_PATH_SEPARATOR, step, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR, query}, "")
			e.SetAwaitingReplyTo(awaitingReplyTo)
		}
	} else { // Has no query - no "?" character
		if !e.IsAwaitingReplyTo(step) {
			awaitingReplyTo = awaitingReplyTo + AWAITING_REPLY_TO_PATH_SEPARATOR + step
			e.SetAwaitingReplyTo(awaitingReplyTo)
		}
	}
}

func (e *BotChatEntity) AddWizardParam(key, value string) {
	awaitingReplyTo := e.GetAwaitingReplyTo()
	awaitingUrl, err := url.Parse(awaitingReplyTo)
	if err != nil {
		panic(fmt.Sprintf("Failed to call url.Parse(awaitingReplyTo=%v)", awaitingReplyTo))
	}
	query := awaitingUrl.Query()
	query.Set(key, value)
	awaitingUrl.RawQuery = query.Encode()
	e.SetAwaitingReplyTo(awaitingUrl.String())
}

func (e *BotChatEntity) GetWizardParam(key string) string {
	if u, err := url.Parse(e.GetAwaitingReplyTo()); err != nil {
		return ""
	} else {
		return u.Query().Get(key)
	}
}
