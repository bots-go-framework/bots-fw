package botsfw

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/strongo/app/user"
	"net/url"
	"strings"
	"time"
)

// BotEntity holds properties common to al bot entities
type BotEntity struct {
	AccessGranted bool
	user.OwnedByUserWithIntID
}

// IsAccessGranted indicates if access to the bot has been granted
func (e *BotEntity) IsAccessGranted() bool {
	return e.AccessGranted
}

// SetAccessGranted mark that access has been granted
func (e *BotEntity) SetAccessGranted(value bool) bool {
	if e.AccessGranted != value {
		e.AccessGranted = value
		return true
	}
	return false
}

// BotUserData hold common properties for bot user entities
type BotUserData struct {
	BotEntity
	user.LastLogin

	FirstName string // required
	LastName  string // optional
	UserName  string // optional
}

// BotChatData hold common properties for bot chat entities not specific to any platform
type BotChatData struct {
	BotEntity
	AppUserIntIDs []int64
	BotID         string `datastore:",noindex"`
	//
	IsGroup bool   `datastore:",noindex,omitempty"`
	Type    string `datastore:",noindex,omitempty"`
	Title   string `datastore:",noindex,omitempty"`
	//
	AwaitingReplyTo   string    `datastore:",noindex,omitempty"`
	PreferredLanguage string    `datastore:",noindex,omitempty"`
	GaClientID        []byte    `datastore:",noindex,omitempty"`
	DtLastInteraction time.Time `datastore:",omitempty"`
	InteractionsCount int       `datastore:",omitempty"`
	DtForbidden       time.Time `datastore:",omitempty"`
	DtForbiddenLast   time.Time `datastore:",noindex,omitempty"`
	LanguageCodes     []string  `datastore:",noindex"` // UI languages
}

var _ BotChat = (*BotChatData)(nil)

// GetAppUserStrID returns app user ID as string
func (e *BotChatData) GetAppUserStrID() string {
	return fmt.Sprintf("%d", e.GetAppUserIntID())
}

// GetBotID returns bot ID
func (e *BotChatData) GetBotID() string {
	return e.BotID
}

// IsGroupChat indicates if it is a group chat
func (e *BotChatData) IsGroupChat() bool {
	return e.IsGroup
}

// SetIsGroupChat marks chat as a group one
func (e *BotChatData) SetIsGroupChat(v bool) {
	e.IsGroup = v
}

// SetBotID sets bot ID
func (e *BotChatData) SetBotID(botID string) {
	e.BotID = botID
}

// AddClientLanguage adds client UI language
func (e *BotChatData) AddClientLanguage(languageCode string) (changed bool) {
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

// func (e *BotChatData) GetBotUserIntID() int {
// 	panic("Should be overwritten in subclass")
// }
//
// func (e *BotChatData) GetBotUserStringID() string {
// 	panic("Should be overwritten in subclass")
// }

// SetBotUserID sets bot user ID
func (e *BotChatData) SetBotUserID(id interface{}) {
	panic(fmt.Sprintf("Should be overwritten in subclass, got: %T=%v", id, id))
}

// SetDtLastInteraction sets date time of last interaction
func (e *BotChatData) SetDtLastInteraction(v time.Time) {
	e.DtLastInteraction = v
	e.InteractionsCount++
}

// GetGaClientID returns Google Analytics client UUID
func (e *BotChatData) GetGaClientID() string {
	var v uuid.UUID
	var err error
	if len(e.GaClientID) == 0 {
		var err error
		v, err = uuid.NewRandom()
		if err != nil {
			panic(fmt.Sprintf("Failed to create UUID: %v", err))
		}
		e.GaClientID = v[:]
	} else if v, err = uuid.FromBytes(e.GaClientID); err != nil {
		panic(fmt.Sprintf("Failed to create UUID from bytes: len(%v)=%v", e.GaClientID, len(e.GaClientID)))
	}
	return v.String()
}

// SetDtUpdateToNow mark entity updated with now
func (e *BotChatData) SetDtUpdateToNow() {
	e.DtUpdated = time.Now()
}

// GetAwaitingReplyTo returns current state
func (e *BotChatData) GetAwaitingReplyTo() string {
	return e.AwaitingReplyTo
}

// SetAwaitingReplyTo sets current state
func (e *BotChatData) SetAwaitingReplyTo(value string) {
	e.AwaitingReplyTo = strings.TrimLeft(value, "/")
}

// GetPreferredLanguage returns preferred language
func (e *BotChatData) GetPreferredLanguage() string {
	return e.PreferredLanguage
}

// SetPreferredLanguage sets preferred language
func (e *BotChatData) SetPreferredLanguage(value string) {
	e.PreferredLanguage = value
}

// IsAwaitingReplyTo returns true if bot us awaiting reply to a specific command
func (e *BotChatData) IsAwaitingReplyTo(code string) bool {
	awaitingReplyToPath := e.getAwaitingReplyToPath()
	return awaitingReplyToPath == code || strings.HasSuffix(awaitingReplyToPath, AwaitingReplyToPathSeparator+code)
}

func (e *BotChatData) getAwaitingReplyToPath() string {
	pathAndQuery := strings.SplitN(e.AwaitingReplyTo, AwaitingReplyToPath2QuerySeparator, 2)
	if len(pathAndQuery) > 1 {
		return pathAndQuery[0]
	}
	return e.AwaitingReplyTo
}

// PopStepsFromAwaitingReplyUpToSpecificParent go back in state
func (e *BotChatData) PopStepsFromAwaitingReplyUpToSpecificParent(step string) {
	awaitingReplyTo := e.AwaitingReplyTo
	pathAndQuery := strings.SplitN(awaitingReplyTo, AwaitingReplyToPath2QuerySeparator, 2)
	path := pathAndQuery[0]
	steps := strings.Split(path, AwaitingReplyToPathSeparator)
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i] == step {
			if i < len(steps)-1 {
				path = strings.Join(steps[:i+1], AwaitingReplyToPathSeparator)
				if len(pathAndQuery) > 1 {
					query := pathAndQuery[1]
					e.SetAwaitingReplyTo(path + AwaitingReplyToPath2QuerySeparator + query)
				} else {
					e.SetAwaitingReplyTo(path)
				}
			}
			//steps = steps[:i]
			break
			// } else {
			// log.Infof(c, "steps[%v]: %v != %v:", i, steps[i], step)
		}
	}
}

// PushStepToAwaitingReplyTo go down in state
func (e *BotChatData) PushStepToAwaitingReplyTo(step string) {
	awaitingReplyTo := e.AwaitingReplyTo
	pathAndQuery := strings.SplitN(awaitingReplyTo, AwaitingReplyToPath2QuerySeparator, 2)
	if len(pathAndQuery) > 1 { // Has query part - something after "?" character
		if !e.IsAwaitingReplyTo(step) {
			path := pathAndQuery[0]
			query := pathAndQuery[1]
			awaitingReplyTo = strings.Join([]string{path, AwaitingReplyToPathSeparator, step, AwaitingReplyToPath2QuerySeparator, query}, "")
			e.SetAwaitingReplyTo(awaitingReplyTo)
		}
	} else { // Has no query - no "?" character
		if !e.IsAwaitingReplyTo(step) {
			awaitingReplyTo = awaitingReplyTo + AwaitingReplyToPathSeparator + step
			e.SetAwaitingReplyTo(awaitingReplyTo)
		}
	}
}

// AddWizardParam adds context param to state
func (e *BotChatData) AddWizardParam(key, value string) {
	awaitingReplyTo := e.GetAwaitingReplyTo()
	awaitingURL, err := url.Parse(awaitingReplyTo)
	if err != nil {
		panic(fmt.Sprintf("Failed to call url.Parse(awaitingReplyTo=%v)", awaitingReplyTo))
	}
	query := awaitingURL.Query()
	query.Set(key, value)
	awaitingURL.RawQuery = query.Encode()
	e.SetAwaitingReplyTo(awaitingURL.String())
}

// GetWizardParam returns state param value
func (e *BotChatData) GetWizardParam(key string) string {
	u, err := url.Parse(e.GetAwaitingReplyTo())
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}
