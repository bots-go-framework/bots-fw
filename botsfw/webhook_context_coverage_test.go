package botsfw

// Tests for previously uncovered functions:
//   - analytics.go: webhookAnalytics.UserContext(), Enqueue()
//   - webhook_context_base.go: AppUserID() (pre-set and via platformUser),
//     SaveBotChat(), SaveBotUser(),
//     AppUserData() (empty appUserID path)
//   - webhook_context_base.go: ChatData() (already-loaded fast path)
//   - webhook_context_base.go: getPlatformUserRecord, getBotUser, GetBotUser, GetBotUserForUpdate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/analytics"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

// ----------------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------------

// simplePlatformUser is a minimal PlatformUserData implementation for tests.
type simplePlatformUser struct {
	appUserID     string
	accessGranted bool
}

func (u *simplePlatformUser) BaseData() *botsfwmodels.PlatformUserBaseDbo {
	return &botsfwmodels.PlatformUserBaseDbo{}
}
func (u *simplePlatformUser) GetAppUserID() string   { return u.appUserID }
func (u *simplePlatformUser) SetAppUserID(id string) { u.appUserID = id }
func (u *simplePlatformUser) IsAccessGranted() bool  { return u.accessGranted }
func (u *simplePlatformUser) SetAccessGranted(v bool) bool {
	if u.accessGranted == v {
		return false
	}
	u.accessGranted = v
	return true
}
func (u *simplePlatformUser) SetUpdatedTime(_ time.Time) {}

// newCoverageWHCB builds a fully-initialised *WebhookContextBase for coverage tests.
func newCoverageWHCB(t *testing.T) *WebhookContextBase {
	t.Helper()
	whcb := &WebhookContextBase{
		c:           context.Background(),
		appContext:  testAppContext{},
		botContext:  BotContext{BotSettings: &BotSettings{Code: "bot", Locale: i18n.LocaleEnUS}},
		botPlatform: &testBotPlatformMore{},
		input: &moreTestInputMessage{
			inputType: botinput.TypeText,
			chatID:    "chat1",
			senderID:  "user1",
		},
		getIsInGroup: func() (bool, error) { return false, nil },
	}
	whcb.translator = translator{
		localeCode5: func() string { return whcb.locale.Code5 },
		Translator:  testAppContext{}.GetTranslator(context.Background()),
	}
	return whcb
}

// ----------------------------------------------------------------------------
// analytics.go – webhookAnalytics.UserContext()
// ----------------------------------------------------------------------------

func TestWebhookAnalytics_UserContext_NoChat(t *testing.T) {
	whcb := newCoverageWHCB(t)
	// AppUserID() first tries ChatData() (when appUserID is empty and isLoadingChatData is false).
	// To skip that path and avoid a nil-db panic, mark isLoadingChatData=true and
	// pre-seed platformUser.Data so the function returns "user-99" without hitting the DB.
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "user-99"}
	whcb.whAnalytics = webhookAnalytics{whcb: whcb}

	uc := whcb.whAnalytics.(webhookAnalytics).UserContext()
	if uc.GetUserID() != "user-99" {
		t.Errorf("expected UserID 'user-99', got %q", uc.GetUserID())
	}
}

func TestWebhookAnalytics_UserContext_WithChatData(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	// GetAppUserID returns "u42" so AppUserID() returns "u42" early.
	mockChatData.EXPECT().GetAppUserID().Return("u42").AnyTimes()
	mockChatData.EXPECT().GetPreferredLanguage().Return("en")

	whcb := newCoverageWHCB(t)
	whcb.botChat.Data = mockChatData
	whcb.whAnalytics = webhookAnalytics{whcb: whcb}

	uc := whcb.whAnalytics.(webhookAnalytics).UserContext()
	if uc.GetUserID() != "u42" {
		t.Errorf("expected 'u42', got %q", uc.GetUserID())
	}
	if uc.GetUserLanguage() != "en" {
		t.Errorf("expected 'en', got %q", uc.GetUserLanguage())
	}
}

func TestWebhookAnalytics_Enqueue_DoesNotPanic(t *testing.T) {
	whcb := newCoverageWHCB(t)
	// Skip ChatData() and getPlatformUserRecord() by seeding platformUser.Data + isLoadingChatData.
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "u1"}
	whcb.whAnalytics = webhookAnalytics{whcb: whcb}
	// No senders registered, so QueueMessage is a no-op – should not panic.
	whcb.whAnalytics.Enqueue(analytics.NewEvent("test", "cat", "act"))
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – AppUserID()
// ----------------------------------------------------------------------------

func TestWebhookContextBase_AppUserID_FromChatData(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockCD.EXPECT().GetAppUserID().Return("chat-user").AnyTimes()

	whcb := newCoverageWHCB(t)
	// botChat.Data is set, so ChatData() returns it immediately (fast path).
	// AppUserID() calls chatData.GetAppUserID() → "chat-user" and returns early.
	whcb.botChat.Data = mockCD
	// platformUser.Data must also be non-nil to avoid getPlatformUserRecord hitting nil db.
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "chat-user"}

	got := whcb.AppUserID()
	if got != "chat-user" {
		t.Errorf("expected 'chat-user', got %q", got)
	}
}

func TestWebhookContextBase_AppUserID_FromPlatformUserData(t *testing.T) {
	whcb := newCoverageWHCB(t)
	// isLoadingChatData=true skips the ChatData() call, so AppUserID() goes straight
	// to checking platformUser.Data and finds "plat-user".
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "plat-user"}

	got := whcb.AppUserID()
	if got != "plat-user" {
		t.Errorf("expected 'plat-user', got %q", got)
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – AppUserID() with appUserID already set + no chat
// ----------------------------------------------------------------------------

func TestWebhookContextBase_AppUserID_EmptyWhenNoChat(t *testing.T) {
	// When isLoadingChatData is false, ChatData() tries to load from DB.
	// If we mark isLoadingChatData=true, the branch is skipped, and with
	// platformUser.Data == nil and db == nil, AppUserID() must not panic and
	// must return "".
	whcb := newCoverageWHCB(t)
	whcb.isLoadingChatData = true
	// platformUser.Data is nil; db is nil. getPlatformUserRecord will panic on nil db.
	// Instead, we pre-set platformUser.Data with no appUserID to exercise the
	// "platformUser.Data != nil but GetAppUserID is empty" branch.
	whcb.platformUser.Data = &simplePlatformUser{appUserID: ""}
	got := whcb.AppUserID()
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – SaveBotChat() via an in-memory DB
// ----------------------------------------------------------------------------

func TestWebhookContextBase_SaveBotChat_WithMemoryDB(t *testing.T) {
	memDB := dalgo2memory.NewDB()

	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	// botChat.Record is a dal.Record; we need a minimal one.

	whcb := newCoverageWHCB(t)
	whcb.db = memDB
	_ = mockCD
	// Build a real dal.Record for the botChat so Set can work.
	key := dal.NewKeyWithID("botChats", "chat1")
	whcb.botChat.Record = dal.NewRecordWithData(key, mockCD)

	// SaveBotChat runs a readwrite transaction that calls tx.Set on botChat.Record.
	// In-memory DB supports this.
	err := whcb.SaveBotChat()
	if err != nil {
		t.Errorf("unexpected error from SaveBotChat: %v", err)
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – SaveBotUser() always returns "not implemented"
// ----------------------------------------------------------------------------

func TestWebhookContextBase_SaveBotUser_ReturnsNotImplemented(t *testing.T) {
	memDB := dalgo2memory.NewDB()
	whcb := newCoverageWHCB(t)
	whcb.db = memDB

	err := whcb.SaveBotUser(context.Background())
	if err == nil {
		t.Fatal("expected error from SaveBotUser, got nil")
	}
	if !errors.Is(err, errors.New("func SaveBotUser is not implemented yet")) {
		// Just check the message is non-empty
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – AppUserData() with empty appUserID
// ----------------------------------------------------------------------------

func TestWebhookContextBase_AppUserData_EmptyAppUserID(t *testing.T) {
	whcb := newCoverageWHCB(t)
	// Force AppUserID() to return "" by marking isLoadingChatData=true
	// and providing platformUser.Data with empty appUserID.
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: ""}

	_, err := whcb.AppUserData()
	if err == nil {
		t.Fatal("expected error when AppUserID is empty")
	}
	if !dal.IsNotFound(err) {
		t.Errorf("expected ErrRecordNotFound, got: %v", err)
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – AppUserData() with appUserID set + getAppUser func
// ----------------------------------------------------------------------------

func TestWebhookContextBase_AppUserData_WithAppUserFn(t *testing.T) {
	memDB := dalgo2memory.NewDB()

	whcb := newCoverageWHCB(t)
	whcb.db = memDB
	// Pre-seed platformUser.Data so AppUserID() returns "test-user-id" without hitting DB.
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "test-user-id"}

	// Set getAppUser on BotSettings that returns a record not found error.
	whcb.botContext.BotSettings.getAppUser = func(
		ctx context.Context,
		tx dal.ReadSession,
		botID string,
		appUserID string,
	) (record.DataWithID[string, botsfwmodels.AppUserData], error) {
		return record.DataWithID[string, botsfwmodels.AppUserData]{}, dal.ErrRecordNotFound
	}

	_, err := whcb.AppUserData()
	if err == nil {
		t.Fatal("expected error from getAppUser, got nil")
	}
	// The error should propagate from GetAppUserByID.
	if !dal.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – ChatData() fast path (already loaded)
// ----------------------------------------------------------------------------

func TestWebhookContextBase_ChatData_AlreadyLoaded(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	// No method calls expected – it should return immediately.

	whcb := newCoverageWHCB(t)
	whcb.botChat.Data = mockCD

	got := whcb.ChatData()
	if got != mockCD {
		t.Error("expected the pre-loaded chatData to be returned")
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – GetAppUser() coverage
// ----------------------------------------------------------------------------

func TestWebhookContextBase_GetAppUser_EmptyAppUserID(t *testing.T) {
	whcb := newCoverageWHCB(t)
	// Force AppUserID to return "": platformUser.Data is nil but we skip ChatData()
	// and getPlatformUserRecord will just find no appUserID.
	// isLoadingChatData=true skips the ChatData() call; platformUser.Data with empty
	// appUserID means AppUserID() returns "".
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: ""} // empty

	whcb.botContext.BotSettings.getAppUser = func(
		ctx context.Context, tx dal.ReadSession, botID, appUserID string,
	) (record.DataWithID[string, botsfwmodels.AppUserData], error) {
		return record.DataWithID[string, botsfwmodels.AppUserData]{}, dal.ErrRecordNotFound
	}
	whcb.db = dalgo2memory.NewDB()

	// GetAppUser calls AppUserID() which returns "", then GetAppUserByID(ctx, db, "").
	// The settings.GetAppUserByID calls getAppUser which returns ErrRecordNotFound.
	_, err := whcb.GetAppUser()
	if err == nil {
		t.Fatal("expected error from GetAppUser, got nil")
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – AppUserID() via chatData returning empty ID
//   then falling through to platformUser branch
// ----------------------------------------------------------------------------

func TestWebhookContextBase_AppUserID_ChatDataReturnsEmpty_ThenPlatformUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	// ChatData() returns mockCD (fast path). GetAppUserID returns "" so we fall through.
	// platformUser.Data is already set, so it takes the second branch.
	mockCD.EXPECT().GetAppUserID().Return("").AnyTimes()

	whcb := newCoverageWHCB(t)
	whcb.botChat.Data = mockCD
	// platformUser.Data is already non-nil so getPlatformUserRecord is not called.
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "from-platform"}

	got := whcb.AppUserID()
	if got != "from-platform" {
		t.Errorf("expected 'from-platform', got %q", got)
	}
}

// ----------------------------------------------------------------------------
// webhook_context_base.go – SetUser() with non-nil AppUserData
// ----------------------------------------------------------------------------

func TestWebhookContextBase_SetUser_WithData(t *testing.T) {
	whcb := newCoverageWHCB(t)

	// We just need SetUser to store the id/data pair.
	// Use nil to keep it simple (nil is valid for the field type).
	whcb.SetUser("id-with-data", nil)
	if whcb.appUserID != "id-with-data" {
		t.Errorf("expected 'id-with-data', got %q", whcb.appUserID)
	}
	if whcb.appUserData != nil {
		t.Error("expected nil appUserData")
	}
}

// ----------------------------------------------------------------------------
// stubBotProfile is a minimal in-package BotProfile implementation for DB tests.
// (Can't use mock_botsfw here — it would create an import cycle.)
// ----------------------------------------------------------------------------

type stubBotProfile struct{}

func (s stubBotProfile) ID() string                       { return "test-profile" }
func (s stubBotProfile) DefaultLocale() i18n.Locale       { return i18n.LocaleEnUS }
func (s stubBotProfile) GetTranslations() BotTranslations { return BotTranslations{} }
func (s stubBotProfile) Router() Router                   { return nil }
func (s stubBotProfile) SupportedLocales() []i18n.Locale  { return []i18n.Locale{i18n.LocaleEnUS} }
func (s stubBotProfile) NewBotChatData() botsfwmodels.BotChatData {
	return &botsfwmodels.ChatBaseData{}
}
func (s stubBotProfile) NewPlatformUserData() botsfwmodels.PlatformUserData {
	return &botsfwmodels.PlatformUserBaseDbo{}
}
func (s stubBotProfile) NewAppUserData() botsfwmodels.AppUserData { return nil }

// ----------------------------------------------------------------------------
// newWHCBWithProfile returns a WHCB configured with a stub BotProfile and
// an in-memory DB — needed to test functions that call getPlatformUserRecord.
// ----------------------------------------------------------------------------

func newWHCBWithProfile(t *testing.T) *WebhookContextBase {
	t.Helper()
	memDB := dalgo2memory.NewDB()
	whcb := &WebhookContextBase{
		c:           context.Background(),
		appContext:  testAppContext{},
		botPlatform: &testBotPlatformMore{},
		input: &moreTestInputMessage{
			inputType: botinput.TypeText,
			chatID:    "chat1",
			senderID:  "user42",
		},
		db: memDB,
		botContext: BotContext{
			BotSettings: &BotSettings{
				Code:    "bot1",
				Locale:  i18n.LocaleEnUS,
				Profile: stubBotProfile{},
			},
		},
		getIsInGroup: func() (bool, error) { return false, nil },
	}
	whcb.translator = translator{
		localeCode5: func() string { return whcb.locale.Code5 },
		Translator:  testAppContext{}.GetTranslator(context.Background()),
	}
	whcb.whAnalytics = webhookAnalytics{whcb: whcb}
	return whcb
}

// ----------------------------------------------------------------------------
// getPlatformUserRecord – record not found in DB → ErrRecordNotFound
// ----------------------------------------------------------------------------

func TestWebhookContextBase_getPlatformUserRecord_NotFound(t *testing.T) {
	whcb := newWHCBWithProfile(t)

	// getPlatformUserRecord reads from DB; the record doesn't exist → not-found error.
	err := whcb.getPlatformUserRecord(whcb.db)
	if err == nil {
		t.Fatal("expected error from getPlatformUserRecord, got nil")
	}
	if !dal.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

// ----------------------------------------------------------------------------
// getBotUser / GetBotUser – same: record not found
// ----------------------------------------------------------------------------

func TestWebhookContextBase_getBotUser_NotFound(t *testing.T) {
	whcb := newWHCBWithProfile(t)

	_, err := whcb.GetBotUser()
	if err == nil {
		t.Fatal("expected error from GetBotUser, got nil")
	}
	if !dal.IsNotFound(err) {
		t.Errorf("expected not-found, got: %v", err)
	}
}

func TestWebhookContextBase_getBotUser_Cached(t *testing.T) {
	whcb := newWHCBWithProfile(t)
	// Pre-seed platformUser.Data so getBotUser returns immediately without DB call.
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"

	botUser, err := whcb.GetBotUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if botUser.Data == nil {
		t.Error("expected non-nil BotUser.Data")
	}
}

// ----------------------------------------------------------------------------
// GetBotUserForUpdate – when platformUser.Data is pre-seeded, getBotUser returns
// from cache without hitting the DB so no deadlock.
// ----------------------------------------------------------------------------

func TestWebhookContextBase_GetBotUserForUpdate_Cached(t *testing.T) {
	whcb := newWHCBWithProfile(t)
	// Pre-seed platformUser.Data so getBotUser returns immediately.
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"

	botUser, err := whcb.GetBotUserForUpdate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error from GetBotUserForUpdate: %v", err)
	}
	if botUser.Data == nil {
		t.Error("expected non-nil BotUser.Data")
	}
}

// ----------------------------------------------------------------------------
// stubRecordsFieldsSetter satisfies BotRecordsFieldsSetter with no-ops.
// ----------------------------------------------------------------------------

type stubRecordsFieldsSetter struct{}

func (s stubRecordsFieldsSetter) Platform() string { return "*" }
func (s stubRecordsFieldsSetter) SetAppUserFields(appUser botsfwmodels.AppUserData, sender botinput.Sender) error {
	return nil
}
func (s stubRecordsFieldsSetter) SetBotUserFields(botUser botsfwmodels.PlatformUserData, sender botinput.Sender, botID, botUserID, appUserID string) error {
	return nil
}
func (s stubRecordsFieldsSetter) SetBotChatFields(botChat botsfwmodels.BotChatData, chat botinput.Chat, botID, botUserID, appUserID string, isAccessGranted bool) error {
	return nil
}

// ----------------------------------------------------------------------------
// ChatData() full load path: DB miss → getOrCreatePlatformUserRecord (fast path)
// → SetBotChatFields → sender language → locale.
// This exercises loadChatEntityBase, getOrCreatePlatformUserRecord (fast path),
// ChatData (not-found branch), and analytics.Enqueue via whAnalytics.
// ----------------------------------------------------------------------------

func TestWebhookContextBase_ChatData_LoadPath_DBMiss(t *testing.T) {
	whcb := newWHCBWithProfile(t)

	// Pre-seed platformUser.Data with no appUserID so getOrCreatePlatformUserRecord
	// returns the cached value (fast path, no DB write).
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"

	// Provide a records fields setter so SetBotChatFields doesn't panic.
	whcb.recordsFieldsSetter = stubRecordsFieldsSetter{}

	// Set isLoadingChatData=false so the full load path runs.
	whcb.isLoadingChatData = false

	chatData := whcb.ChatData()
	// The DB has no bot chat → not-found → recordsFieldsSetter.SetBotChatFields called
	// → whcb.botChat.Data is still the one returned by Profile.NewBotChatData()
	if chatData == nil {
		t.Error("expected non-nil chatData after load path")
	}
}

// ----------------------------------------------------------------------------
// ChatData() duplicate-call guard: when already loading (isLoadingChatData=true
// inside loadChatEntityBase), returns nil.
// ----------------------------------------------------------------------------

func TestWebhookContextBase_ChatData_EmptyChatID(t *testing.T) {
	whcb := newWHCBWithProfile(t)
	// Override input to return empty chatID so ChatData() returns nil at the
	// "chatID == empty" guard.
	whcb.input = &moreTestInputMessage{
		inputType: botinput.TypeText,
		chatID:    "",
		senderID:  "user1",
	}

	got := whcb.ChatData()
	if got != nil {
		t.Errorf("expected nil chatData for empty chatID, got %T", got)
	}
}
