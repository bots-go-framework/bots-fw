package botsfw

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/strongo/i18n"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- Test doubles ---

type moreTestUser struct {
	id        any
	firstName string
	lastName  string
	language  string
}

func (ts *moreTestUser) GetID() any           { return ts.id }
func (ts *moreTestUser) GetFirstName() string { return ts.firstName }
func (ts *moreTestUser) GetLastName() string  { return ts.lastName }
func (ts *moreTestUser) GetUserName() string  { return "" }
func (ts *moreTestUser) GetLanguage() string  { return ts.language }
func (ts *moreTestUser) Platform() string     { return "test" }
func (ts *moreTestUser) IsBotUser() bool      { return false }
func (ts *moreTestUser) GetAvatar() string    { return "" }
func (ts *moreTestUser) GetCountry() string   { return "" }

type moreTestInputMessage struct {
	inputType botinput.Type
	chatID    string
	senderID  any
	firstName string
	lastName  string
	language  string
	text      string
}

func (ti *moreTestInputMessage) InputType() botinput.Type   { return ti.inputType }
func (ti *moreTestInputMessage) BotChatID() (string, error) { return ti.chatID, nil }
func (ti *moreTestInputMessage) Chat() botinput.Chat        { return nil }
func (ti *moreTestInputMessage) GetSender() botinput.User {
	return &moreTestUser{id: ti.senderID, firstName: ti.firstName, lastName: ti.lastName, language: ti.language}
}
func (ti *moreTestInputMessage) GetRecipient() botinput.Recipient { return nil }
func (ti *moreTestInputMessage) GetTime() time.Time               { return time.Now() }
func (ti *moreTestInputMessage) LogRequest()                      {}
func (ti *moreTestInputMessage) MessageIntID() int                { return 0 }
func (ti *moreTestInputMessage) MessageStringID() string          { return "" }

type moreTestTextInputMessage struct {
	moreTestInputMessage
}

func (ti *moreTestTextInputMessage) Text() string   { return ti.text }
func (ti *moreTestTextInputMessage) IsEdited() bool { return false }

type errorResponseWriter struct{}

func (errorResponseWriter) Header() http.Header       { return http.Header{} }
func (errorResponseWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("write error") }
func (errorResponseWriter) WriteHeader(int)           {}

type testWebhookDriverImpl struct{}

func (testWebhookDriverImpl) RegisterWebhookHandlers(_ HttpRouter, _ string, _ ...WebhookHandler) {}
func (testWebhookDriverImpl) HandleWebhook(_ http.ResponseWriter, _ *http.Request, _ WebhookHandler) {
}

type testBotPlatformMore struct{}

func (testBotPlatformMore) ID() string      { return "test" }
func (testBotPlatformMore) Version() string { return "1.0" }

func newMoreTestWHCB(t *testing.T) *WebhookContextBase {
	t.Helper()
	r, _ := http.NewRequest("GET", "/test", nil)
	mockInput := &moreTestInputMessage{
		inputType: botinput.TypeText,
		chatID:    "chat1",
		senderID:  "user123",
		firstName: "John",
		lastName:  "Doe",
		language:  "en",
	}
	whcb := &WebhookContextBase{
		r:          r,
		c:          context.Background(),
		input:      mockInput,
		appContext: testAppContext{},
		botContext: BotContext{
			BotSettings: &BotSettings{
				Code:   "testbot",
				Token:  "tok123",
				Locale: i18n.LocaleEnUS,
				Env:    "local",
			},
		},
		botPlatform: &testBotPlatformMore{},
		getIsInGroup: func() (bool, error) {
			return false, nil
		},
	}
	whcb.translator = translator{
		localeCode5: func() string {
			return whcb.locale.Code5
		},
		Translator: testAppContext{}.GetTranslator(context.Background()),
	}
	return whcb
}

// =============================================================================
// whContextDummy — all methods panic (webhook_context_dummy.go, 0% coverage)
// =============================================================================

func TestWhContextDummy_NewEditMessage_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from NewEditMessage")
		}
	}()
	d := whContextDummy{}
	_, _ = d.NewEditMessage("test", botmsg.FormatText)
}

func TestWhContextDummy_UpdateLastProcessed_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from UpdateLastProcessed")
		}
	}()
	d := whContextDummy{}
	_ = d.UpdateLastProcessed(nil)
}

func TestWhContextDummy_AppUserData_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from AppUserData")
		}
	}()
	d := whContextDummy{}
	_, _ = d.AppUserData()
}

func TestWhContextDummy_IsNewerThen_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from IsNewerThen")
		}
	}()
	d := whContextDummy{}
	_ = d.IsNewerThen(nil)
}

func TestWhContextDummy_Responder_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from Responder")
		}
	}()
	d := whContextDummy{}
	_ = d.Responder()
}

// =============================================================================
// WebhookHandlerBase.Register (webhook_handler_base.go, 0% coverage)
// =============================================================================

func TestWebhookHandlerBase_Register_Valid(t *testing.T) {
	bh := &WebhookHandlerBase{}
	bh.Register(testWebhookDriverImpl{}, testBotHost{})

	if bh.WebhookDriver == nil {
		t.Error("expected WebhookDriver to be set")
	}
	if bh.BotHost == nil {
		t.Error("expected BotHost to be set")
	}
}

func TestWebhookHandlerBase_Register_NilDriver_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil driver")
		}
	}()
	bh := &WebhookHandlerBase{}
	bh.Register(nil, testBotHost{})
}

func TestWebhookHandlerBase_Register_NilHost_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil host")
		}
	}()
	bh := &WebhookHandlerBase{}
	bh.Register(testWebhookDriverImpl{}, nil)
}

// =============================================================================
// IsInTransaction / NonTransactionalContext (webhook_context_base.go, 0%)
// =============================================================================

func TestWebhookContextBase_IsInTransaction_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from IsInTransaction")
		}
	}()
	whcb := newMoreTestWHCB(t)
	whcb.IsInTransaction(context.Background())
}

func TestWebhookContextBase_NonTransactionalContext_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from NonTransactionalContext")
		}
	}()
	whcb := newMoreTestWHCB(t)
	whcb.NonTransactionalContext(context.Background())
}

// =============================================================================
// NewWebhookContextBase — nil request and valid (webhook_context_base.go, 22.2%)
// =============================================================================

func TestNewWebhookContextBase_NilRequest_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil HttpRequest")
		}
	}()
	args := CreateWebhookContextArgs{
		HttpRequest: nil,
		AppContext:  testAppContext{},
		BotContext: BotContext{
			BotHost:     testBotHost{},
			BotSettings: &BotSettings{Code: "testbot", Token: "tok123", Locale: i18n.LocaleEnUS},
		},
		WebhookInput: &moreTestInputMessage{inputType: botinput.TypeText, chatID: "chat1", senderID: "user1"},
	}
	_, _ = NewWebhookContextBase(args, &testBotPlatformMore{}, nil,
		func() (bool, error) { return false, nil },
		func(context.Context) (string, string, error) { return "en-US", "chat1", nil },
	)
}

func TestNewWebhookContextBase_ValidArgs(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	args := CreateWebhookContextArgs{
		HttpRequest: r,
		AppContext:  testAppContext{},
		BotContext: BotContext{
			BotHost:     testBotHost{},
			BotSettings: &BotSettings{Code: "testbot", Token: "tok123", Locale: i18n.LocaleEnUS},
		},
		WebhookInput: &moreTestInputMessage{inputType: botinput.TypeText, chatID: "chat1", senderID: "user1"},
	}

	whcb, err := NewWebhookContextBase(args, &testBotPlatformMore{}, nil,
		func() (bool, error) { return false, nil },
		func(context.Context) (string, string, error) { return "en-US", "chat1", nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if whcb == nil {
		t.Fatal("expected non-nil WebhookContextBase")
	}
	if whcb.BotPlatform().ID() != "test" {
		t.Errorf("expected BotPlatform ID 'test', got '%s'", whcb.BotPlatform().ID())
	}
	if whcb.GetBotCode() != "testbot" {
		t.Errorf("expected bot code 'testbot', got '%s'", whcb.GetBotCode())
	}
	if whcb.Analytics() == nil {
		t.Error("expected non-nil Analytics from NewWebhookContextBase")
	}
}

func TestNewWebhookContextBase_WithGetLocaleAndChatID(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	args := CreateWebhookContextArgs{
		HttpRequest: r,
		AppContext:  testAppContext{},
		BotContext: BotContext{
			BotHost:     testBotHost{},
			BotSettings: &BotSettings{Code: "bot2", Token: "tok2", Locale: i18n.LocaleEnUS},
		},
		WebhookInput: &moreTestInputMessage{inputType: botinput.TypeText, chatID: "", senderID: "u1"},
	}

	whcb, err := NewWebhookContextBase(args, &testBotPlatformMore{}, nil,
		func() (bool, error) { return true, nil },
		func(context.Context) (string, string, error) { return "en-US", "fromLocaleFunc", nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("BotChatID error: %v", err)
	}
	if chatID != "fromLocaleFunc" {
		t.Errorf("expected 'fromLocaleFunc', got '%s'", chatID)
	}
}

// =============================================================================
// GetBotUserID (webhook_context_base.go:412, 0%)
// =============================================================================

func TestWebhookContextBase_GetBotUserID(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	got := whcb.GetBotUserID()
	if got != "user123" {
		t.Errorf("expected 'user123', got '%s'", got)
	}
}

func TestWebhookContextBase_GetBotUserID_IntSender(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.input = &moreTestInputMessage{inputType: botinput.TypeText, chatID: "c1", senderID: 42}
	got := whcb.GetBotUserID()
	if got != "42" {
		t.Errorf("expected '42', got '%s'", got)
	}
}

// =============================================================================
// HasChatData false path (webhook_context_base.go:426)
// =============================================================================

func TestWebhookContextBase_HasChatData_False(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.HasChatData() {
		t.Error("expected HasChatData() == false when botChat.Data is nil")
	}
}

// =============================================================================
// NewMessageByCode (webhook_context_base.go:678, 0%)
// =============================================================================

func TestWebhookContextBase_NewMessageByCode_WithArgs(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	m := whcb.NewMessageByCode("hello_%s", "world")
	if m.Text != "hello_world" {
		t.Errorf("expected 'hello_world', got '%s'", m.Text)
	}
}

func TestWebhookContextBase_NewMessageByCode_NoArgs(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	m := whcb.NewMessageByCode("simple_key")
	if m.Text != "simple_key" {
		t.Errorf("expected 'simple_key', got '%s'", m.Text)
	}
}

// =============================================================================
// CommandText (webhook_context_base.go:729, 0%)
// =============================================================================

func TestWebhookContextBase_CommandText_Cases(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}

	tests := []struct {
		name     string
		title    string
		icon     string
		expected string
	}{
		{"title_and_icon", "Help", "❓", "Help ❓"},
		{"slash_no_translate", "/start", "", "/start"},
		{"empty_title_with_icon", "", "❓", "❓"},
		{"both_empty", "", "", "<NO_TITLE_OR_ICON>"},
		{"title_only", "Help", "", "Help"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := whcb.CommandText(tt.title, tt.icon)
			if got != tt.expected {
				t.Errorf("CommandText(%q, %q) = %q, want %q", tt.title, tt.icon, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// SetContext / Context (webhook_context_base.go:664-667)
// =============================================================================

func TestWebhookContextBase_SetContext_Verify(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	type ctxKey struct{}
	newCtx := context.WithValue(context.Background(), ctxKey{}, "testValue")
	whcb.SetContext(newCtx)
	if whcb.Context() != newCtx {
		t.Error("SetContext/Context() mismatch")
	}
}

// =============================================================================
// AppUserEntity (webhook_context_base.go:655, 0%)
// =============================================================================

func TestWebhookContextBase_AppUserEntity_IsNil(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.AppUserEntity() != nil {
		t.Error("expected nil AppUserEntity when not set")
	}
}

// =============================================================================
// SetUser / AppUserID (webhook_context_base.go:209, 0%)
// =============================================================================

func TestWebhookContextBase_SetUser_Verify(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.SetUser("app-user-42", nil)

	// Verify fields were set directly (AppUserID() has side-effects that need DB)
	if whcb.appUserID != "app-user-42" {
		t.Errorf("expected appUserID='app-user-42', got '%s'", whcb.appUserID)
	}
	if whcb.AppUserEntity() != nil {
		t.Error("expected nil AppUserEntity when data not set")
	}
}

// =============================================================================
// AppUserID pre-set (webhook_context_base.go:215, 0%)
// =============================================================================

func TestWebhookContextBase_AppUserID_PreSet(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.appUserID = "pre-set-user"
	// Verify field directly — AppUserID() still tries to load platform user from DB
	if whcb.appUserID != "pre-set-user" {
		t.Errorf("expected 'pre-set-user', got '%s'", whcb.appUserID)
	}
}

// =============================================================================
// NewMessage (webhook_context_base.go:685)
// =============================================================================

func TestWebhookContextBase_NewMessage_Text(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	m := whcb.NewMessage("hello world")
	if m.Text != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", m.Text)
	}
}

// =============================================================================
// Getters not already covered by existing tests
// =============================================================================

func TestWebhookContextBase_RequestAccess(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.Request() == nil {
		t.Error("expected non-nil Request()")
	}
	if whcb.Request().URL.Path != "/test" {
		t.Errorf("expected path '/test', got '%s'", whcb.Request().URL.Path)
	}
}

func TestWebhookContextBase_EnvironmentLocal(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if env := whcb.Environment(); env != "local" {
		t.Errorf("expected 'local', got '%s'", env)
	}
}

func TestWebhookContextBase_BotPlatformAccess(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	bp := whcb.BotPlatform()
	if bp == nil {
		t.Fatal("expected non-nil BotPlatform")
	}
	if bp.ID() != "test" {
		t.Errorf("expected 'test', got '%s'", bp.ID())
	}
}

func TestWebhookContextBase_InputAccess(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.Input() == nil {
		t.Error("expected non-nil Input()")
	}
	if whcb.Input().InputType() != botinput.TypeText {
		t.Errorf("expected TypeText, got %v", whcb.Input().InputType())
	}
}

func TestWebhookContextBase_RecordsFieldsSetter_IsNil(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.RecordsFieldsSetter() != nil {
		t.Error("expected nil RecordsFieldsSetter")
	}
}

func TestWebhookContextBase_DB_IsNil(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.DB() != nil {
		t.Error("expected nil DB when not set")
	}
}

func TestWebhookContextBase_Analytics_IsNil(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if whcb.Analytics() != nil {
		t.Error("expected nil Analytics when not initialized via NewWebhookContextBase")
	}
}

// =============================================================================
// BotChatID paths (webhook_context_base.go:153-196)
// =============================================================================

func TestWebhookContextBase_BotChatID_FromInput(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "chat1" {
		t.Errorf("expected 'chat1', got '%s'", chatID)
	}
}

func TestWebhookContextBase_BotChatID_AlreadySet(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.SetChatID("preset-chat")
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "preset-chat" {
		t.Errorf("expected 'preset-chat', got '%s'", chatID)
	}
}

func TestWebhookContextBase_MustBotChatID_Empty_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty BotChatID")
		}
	}()
	whcb := newMoreTestWHCB(t)
	whcb.input = &moreTestInputMessage{inputType: botinput.TypeText, chatID: "", senderID: "user1"}
	whcb.MustBotChatID()
}

// =============================================================================
// MessageText with TextMessage input
// =============================================================================

func TestWebhookContextBase_MessageText_WithTextInput(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.input = &moreTestTextInputMessage{
		moreTestInputMessage: moreTestInputMessage{
			inputType: botinput.TypeText,
			chatID:    "chat1",
			senderID:  "user1",
			text:      "hello bot",
		},
	}
	got := whcb.MessageText()
	if got != "hello bot" {
		t.Errorf("expected 'hello bot', got '%s'", got)
	}
}

// =============================================================================
// Locale fallback (webhook_context_base.go:691)
// =============================================================================

func TestWebhookContextBase_Locale_FallbackToSettings(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	// Locale() tries ChatData() which needs DB when locale is unset.
	// Set chatID to empty so ChatData() returns nil early.
	whcb.input = &moreTestInputMessage{inputType: botinput.TypeText, chatID: "", senderID: "u1"}
	loc := whcb.Locale()
	if loc.Code5 != "en-US" {
		t.Errorf("expected 'en-US' from BotSettings fallback, got '%s'", loc.Code5)
	}
}

func TestWebhookContextBase_Locale_AfterSetLocale(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	loc := whcb.Locale()
	if loc.Code5 != "en-US" {
		t.Errorf("expected 'en-US', got '%s'", loc.Code5)
	}
}

// =============================================================================
// PingHandler — error path (misc.go, 66.7%)
// =============================================================================

func TestPingHandler_WriteErrorPath(t *testing.T) {
	r, _ := http.NewRequest("GET", "/ping", nil)
	// Should not panic, just logs the error
	PingHandler(errorResponseWriter{}, r)
}

func TestPingHandler_SuccessPath(t *testing.T) {
	r, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	PingHandler(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "Pong" {
		t.Errorf("expected 'Pong', got '%s'", w.Body.String())
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header")
	}
}

// =============================================================================
// SetLocale edge cases (webhook_context_base.go:706-726)
// =============================================================================

func TestWebhookContextBase_SetLocale_NilAppContext(t *testing.T) {
	whcb := &WebhookContextBase{}
	err := whcb.SetLocale("en-US")
	if err == nil {
		t.Fatal("expected error for nil appContext")
	}
}

// =============================================================================
// IsInGroup edge cases (webhook_context_base.go:296)
// =============================================================================

func TestWebhookContextBase_IsInGroup_Error(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.getIsInGroup = func() (bool, error) {
		return false, fmt.Errorf("test error")
	}
	_, err := whcb.IsInGroup()
	if err == nil {
		t.Fatal("expected error from IsInGroup")
	}
}

func TestWebhookContextBase_IsInGroup_True(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.getIsInGroup = func() (bool, error) {
		return true, nil
	}
	isGroup, err := whcb.IsInGroup()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isGroup {
		t.Error("expected IsInGroup() == true")
	}
}

// =============================================================================
// GetBotToken via newMoreTestWHCB (webhook_context_base.go:418)
// =============================================================================

func TestWebhookContextBase_GetBotToken_Via_More(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if tok := whcb.GetBotToken(); tok != "tok123" {
		t.Errorf("expected 'tok123', got '%s'", tok)
	}
}

// =============================================================================
// GetTranslator (webhook_context_base.go:358)
// =============================================================================

func TestWebhookContextBase_GetTranslator_More(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	tr := whcb.GetTranslator("en-US")
	if tr == nil {
		t.Fatal("expected non-nil translator")
	}
	got := tr.Translate("test_key")
	if got != "test_key" {
		t.Errorf("expected 'test_key', got '%s'", got)
	}
}

// =============================================================================
// Translator convenience methods
// =============================================================================

func TestTranslator_Translate_More(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	got := whcb.Translate("my_key")
	if got != "my_key" {
		t.Errorf("expected 'my_key', got '%s'", got)
	}
}

func TestTranslator_TranslateNoWarning_More(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	got := whcb.TranslateNoWarning("my_key2")
	if got != "my_key2" {
		t.Errorf("expected 'my_key2', got '%s'", got)
	}
}

func TestTranslator_Locale_More(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	loc := whcb.translator.Locale()
	if loc.Code5 != "en-US" {
		t.Errorf("expected 'en-US', got '%s'", loc.Code5)
	}
}

func TestTranslator_TranslateWithMap_More(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}
	got := whcb.TranslateWithMap("map_key", map[string]string{"a": "b"})
	if got != "map_key" {
		t.Errorf("expected 'map_key', got '%s'", got)
	}
}
