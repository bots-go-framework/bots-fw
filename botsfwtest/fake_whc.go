package botsfwtest

import (
	"context"
	"net/http"
	"time"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/i18n"
)

// Option configures a FakeWebhookContext.
type Option func(*FakeWebhookContext)

// WithTextMessage configures the context to present a text message as input.
func WithTextMessage(text string) Option {
	return func(f *FakeWebhookContext) {
		f.input = &FakeTextMessage{
			fakeInputMessage: f.newBaseInput(),
			text:             text,
		}
	}
}

// WithCallbackQuery configures the context to present a callback query as input.
func WithCallbackQuery(data string) Option {
	return func(f *FakeWebhookContext) {
		f.input = &fakeCallbackQuery{
			fakeInputMessage: f.newBaseInput(),
			id:               "cbq1",
			data:             data,
		}
	}
}

// WithBotCode sets the bot code (default: "testbot").
func WithBotCode(code string) Option {
	return func(f *FakeWebhookContext) {
		f.botSettings.Code = code
	}
}

// WithUserID sets the fake sender's user ID (default: "user1").
func WithUserID(id string) Option {
	return func(f *FakeWebhookContext) {
		f.user.id = id
	}
}

// WithChatID sets the fake chat ID (default: "chat1").
func WithChatID(id string) Option {
	return func(f *FakeWebhookContext) {
		f.chat.id = id
	}
}

// WithLocale sets the translator locale (default: en-US).
func WithLocale(code5 string) Option {
	return func(f *FakeWebhookContext) {
		f.localeCode5 = code5
	}
}

// FakeWebhookContext is an in-memory implementation of botsfw.WebhookContext.
// It satisfies the full interface without any network, database, or platform calls.
//
// Typical usage in a test:
//
//	whc := botsfwtest.NewFakeWebhookContext(botsfwtest.WithTextMessage("/start"))
//	m, err := myCommand.TextAction(whc, "/start")
//	assert.NoError(t, err)
//	assert.Contains(t, m.Text, "Welcome")
//	sent := whc.SentMessages()   // messages sent via whc.Responder().SendMessage(...)
type FakeWebhookContext struct {
	ctx         context.Context
	botSettings botsfw.BotSettings
	user        *fakeUser
	chat        *fakeChat
	input       botinput.InputMessage
	chatData    *botsfwmodels.ChatBaseData
	analytics   *fakeAnalytics
	localeCode5 string
	responder   *FakeResponder
}

var _ botsfw.WebhookContext = (*FakeWebhookContext)(nil)

// NewFakeWebhookContext creates a FakeWebhookContext with sensible defaults.
// The default input is a text message with an empty string.
func NewFakeWebhookContext(opts ...Option) *FakeWebhookContext {
	f := &FakeWebhookContext{
		ctx:         context.Background(),
		localeCode5: "en-US",
		user:        &fakeUser{id: "user1", firstName: "Test"},
		chat:        &fakeChat{id: "chat1"},
		chatData:    &botsfwmodels.ChatBaseData{},
		analytics:   &fakeAnalytics{},
		botSettings: botsfw.BotSettings{Code: "testbot", Env: "test"},
	}
	f.responder = &FakeResponder{}
	// Default input: empty text message (must come after user/chat are initialised).
	f.input = &FakeTextMessage{
		fakeInputMessage: f.newBaseInput(),
		text:             "",
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *FakeWebhookContext) newBaseInput() fakeInputMessage {
	return fakeInputMessage{
		user:   f.user,
		chat:   f.chat,
		t:      time.Now(),
		chatID: f.chat.id,
	}
}

// --- SentMessages is the primary assertion surface ---

// SentMessages returns all messages that were sent via this context's responder.
func (f *FakeWebhookContext) SentMessages() []botmsg.MessageFromBot {
	return f.responder.Sent
}

// Analytics returns the fake analytics recorder.
func (f *FakeWebhookContext) Analytics() botsfw.WebhookAnalytics {
	return f.analytics
}

// RecordedAnalyticsMessages returns analytics messages recorded during the test.
func (f *FakeWebhookContext) RecordedAnalyticsMessages() []interface{ Type() string } {
	return nil // analytics.Message does not expose Type in the public API — return nil for now
}

// --- botsfw.WebhookRequestContext ---

func (f *FakeWebhookContext) Context() context.Context     { return f.ctx }
func (f *FakeWebhookContext) SetContext(c context.Context) { f.ctx = c }
func (f *FakeWebhookContext) Request() *http.Request       { return nil }
func (f *FakeWebhookContext) Environment() string          { return f.botSettings.Env }
func (f *FakeWebhookContext) BotPlatform() botsfw.BotPlatform {
	return &fakePlatform{}
}
func (f *FakeWebhookContext) BotContext() botsfw.BotContext {
	return botsfw.BotContext{BotSettings: &f.botSettings}
}
func (f *FakeWebhookContext) GetBotCode() string                  { return f.botSettings.Code }
func (f *FakeWebhookContext) GetBotSettings() *botsfw.BotSettings { return &f.botSettings }
func (f *FakeWebhookContext) AppContext() botsfw.AppContext       { return nil }
func (f *FakeWebhookContext) ExecutionContext() botsfw.ExecutionContext {
	return fakeExecutionContext{ctx: f.ctx}
}

// --- botsfw.WebhookInputContext ---

func (f *FakeWebhookContext) Input() botinput.InputMessage { return f.input }
func (f *FakeWebhookContext) GetBotUserID() string         { return f.user.id }
func (f *FakeWebhookContext) MustBotChatID() string        { return f.chat.id }
func (f *FakeWebhookContext) IsInGroup() (bool, error)     { return false, nil }

// --- botsfw.WebhookUserData ---

func (f *FakeWebhookContext) ChatData() botsfwmodels.BotChatData { return f.chatData }
func (f *FakeWebhookContext) SaveBotChat() error                 { return nil }
func (f *FakeWebhookContext) GetBotUser() (botsfwstore.PlatformUser, error) {
	return botsfwstore.PlatformUser{ID: f.user.id}, nil
}
func (f *FakeWebhookContext) SetBotUserAccessGranted(value bool) error { return nil }
func (f *FakeWebhookContext) AppUserID() string                        { return f.chatData.GetAppUserID() }
func (f *FakeWebhookContext) SetUser(id string, data botsfwmodels.AppUserData) {
	f.chatData.SetAppUserID(id)
}
func (f *FakeWebhookContext) AppUserData() (botsfwmodels.AppUserData, error) { return nil, nil }
func (f *FakeWebhookContext) RecordsFieldsSetter() botsfw.BotRecordsFieldsSetter {
	return nil
}
func (f *FakeWebhookContext) UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error {
	return nil
}
func (f *FakeWebhookContext) IsNewerThen(chatEntity botsfwmodels.BotChatData) bool { return true }

// --- botsfw.WebhookI18n ---

func (f *FakeWebhookContext) Locale() i18n.Locale {
	return i18n.GetLocaleByCode5(f.localeCode5)
}
func (f *FakeWebhookContext) SetLocale(code5 string) error {
	f.localeCode5 = code5
	return nil
}

// Translate is a pass-through translator: it returns the key unchanged.
// Override it in your option if you need specific translated strings.
func (f *FakeWebhookContext) Translate(key string, args ...interface{}) string {
	return key
}

// TranslateNoWarning behaves like Translate.
func (f *FakeWebhookContext) TranslateNoWarning(key string, args ...interface{}) string {
	return key
}

// TranslateWithMap is a pass-through returning the key unchanged.
func (f *FakeWebhookContext) TranslateWithMap(key string, args map[string]string) string {
	return key
}

// GetTranslator returns a pass-through translator for the given locale.
func (f *FakeWebhookContext) GetTranslator(locale string) i18n.SingleLocaleTranslator {
	return &passThroughTranslator{locale: locale}
}

// CommandText concatenates title and icon the same way the framework does.
func (f *FakeWebhookContext) CommandText(title, icon string) string {
	if title == "" {
		return icon
	}
	if icon == "" {
		return title
	}
	return title + " " + icon
}

// --- botsfw.WebhookMessaging ---

func (f *FakeWebhookContext) NewMessage(text string) botmsg.MessageFromBot {
	m := botmsg.MessageFromBot{}
	m.Text = text
	return m
}

func (f *FakeWebhookContext) NewMessageByCode(code string, a ...interface{}) botmsg.MessageFromBot {
	return f.NewMessage(code)
}

func (f *FakeWebhookContext) NewEditMessage(text string, format botmsg.Format) (botmsg.MessageFromBot, error) {
	m := botmsg.MessageFromBot{}
	m.Text = text
	m.Format = format
	m.IsEdit = true
	return m, nil
}

// Responder returns the FakeResponder which records every SendMessage call.
func (f *FakeWebhookContext) Responder() botsfw.WebhookResponder {
	return f.responder
}

// --- botsfw.WebhookTelemetry ---

// Analytics is declared above.

// --- helpers ---

type fakeExecutionContext struct{ ctx context.Context }

func (e fakeExecutionContext) Context() context.Context { return e.ctx }

// passThroughTranslator returns every key unchanged.
type passThroughTranslator struct {
	locale string
}

func (p *passThroughTranslator) Locale() i18n.Locale {
	return i18n.GetLocaleByCode5(p.locale)
}
func (p *passThroughTranslator) Translate(key string, args ...interface{}) string {
	return key
}
func (p *passThroughTranslator) TranslateNoWarning(key string, args ...interface{}) string {
	return key
}
func (p *passThroughTranslator) TranslateWithMap(key string, args map[string]string) string {
	return key
}
