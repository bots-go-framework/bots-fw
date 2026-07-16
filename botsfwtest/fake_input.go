// Package botsfwtest provides a fake in-memory WebhookContext implementation for
// testing bot command handlers without network, database, or platform dependencies.
//
// Usage:
//
//	whc := botsfwtest.NewFakeWebhookContext(
//	    botsfwtest.WithTextMessage("hello"),
//	    botsfwtest.WithBotCode("mybot"),
//	)
//	m, err := myCommand.Action(whc)
//	// inspect m, err and whc.SentMessages()
package botsfwtest

import (
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
)

// fakeUser is a minimal botinput.User implementation.
type fakeUser struct {
	id        string
	firstName string
}

func (u *fakeUser) Platform() string     { return "fake" }
func (u *fakeUser) GetID() any           { return u.id }
func (u *fakeUser) IsBotUser() bool      { return false }
func (u *fakeUser) GetFirstName() string { return u.firstName }
func (u *fakeUser) GetLastName() string  { return "" }
func (u *fakeUser) GetUserName() string  { return u.id }
func (u *fakeUser) GetLanguage() string  { return "en" }
func (u *fakeUser) GetAvatar() string    { return "" }
func (u *fakeUser) GetCountry() string   { return "" }

// fakeChat is a minimal botinput.Chat implementation.
type fakeChat struct {
	id string
}

func (c *fakeChat) GetID() string     { return c.id }
func (c *fakeChat) GetType() string   { return "private" }
func (c *fakeChat) IsGroupChat() bool { return false }

// fakeInputMessage is a shared base for all fake input messages.
type fakeInputMessage struct {
	user   *fakeUser
	chat   *fakeChat
	t      time.Time
	chatID string
}

func (m *fakeInputMessage) GetSender() botinput.User         { return m.user }
func (m *fakeInputMessage) GetRecipient() botinput.Recipient { return nil }
func (m *fakeInputMessage) GetTime() time.Time               { return m.t }
func (m *fakeInputMessage) MessageIntID() int                { return 0 }
func (m *fakeInputMessage) MessageStringID() string          { return "" }
func (m *fakeInputMessage) BotChatID() (string, error)       { return m.chat.id, nil }
func (m *fakeInputMessage) Chat() botinput.Chat              { return m.chat }
func (m *fakeInputMessage) LogRequest()                      {}

// InputTypeCallbackQuery is re-exported for use in test assertions.
const InputTypeCallbackQuery = botinput.TypeCallbackQuery

// FakeTextMessage implements botinput.TextMessage and is exported so test code
// can type-assert whc.Input() to access the Text() method directly.
type FakeTextMessage struct {
	fakeInputMessage
	text string
}

var _ botinput.TextMessage = (*FakeTextMessage)(nil)

func (m *FakeTextMessage) InputType() botinput.Type { return botinput.TypeText }

// Text returns the configured text of the fake message.
func (m *FakeTextMessage) Text() string   { return m.text }
func (m *FakeTextMessage) IsEdited() bool { return false }

// fakeCallbackQuery implements botinput.CallbackQuery.
type fakeCallbackQuery struct {
	fakeInputMessage
	id   string
	data string
}

var _ botinput.CallbackQuery = (*fakeCallbackQuery)(nil)

func (q *fakeCallbackQuery) InputType() botinput.Type     { return botinput.TypeCallbackQuery }
func (q *fakeCallbackQuery) GetID() string                { return q.id }
func (q *fakeCallbackQuery) GetFrom() botinput.Sender     { return q.user }
func (q *fakeCallbackQuery) GetMessage() botinput.Message { return nil }
func (q *fakeCallbackQuery) GetData() string              { return q.data }
func (q *fakeCallbackQuery) Chat() botinput.Chat          { return q.chat }
