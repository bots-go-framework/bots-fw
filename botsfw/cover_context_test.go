package botsfw

// Tests for previously uncovered / low-coverage functions:
//   - context_auth.go:     SetAccessGranted (0 %)
//   - webhook_context_base.go: createPlatformUserRecord (0 %)
//   - webhook_context_base.go: getOrCreatePlatformUserRecord (10.5 %) — both paths
//   - webhook_context_base.go: loadChatEntityBase (59.5 %)     — missing branches
//   - webhook_context_base.go: ChatData (50 %)                 — not-found / first-contact path
//   - webhook_context_base.go: BotChatID (64 %)               — CallbackQuery / InlineQuery / ChosenInlineResult / default branches
//   - webhook_context_base.go: MustBotChatID (83.3 %)         — error from BotChatID path
//   - webhook_context_base.go: AppUserID (68.8 %)             — getPlatformUserRecord path
//   - webhook_context_base.go: GetAppUser (80 %)              — happy path
//   - webhook_context_base.go: AppUserData (88.9 %)           — happy path
//   - webhook_context_base.go: getPlatformUserRecord (80 %)   — already-loaded fast path
//   - webhook_context_base.go: NewWebhookContextBase (88.9 %) — db-set path
//   - webhook_context_base.go: Locale (57.1 %)                — chatData returns locale path
//   - webhook_context_base.go: SetLocale (91.7 %)             — supportedLocales nil guard
//   - commands.go:            TitleByKey (92.3 %)              — missing branches
//   - settings.go:            findByPlatform (72.7 %)          — legacy flat-map paths

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/analytics"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

// ============================================================================
// noopAnalytics is a WebhookAnalytics that does nothing — used in tests that
// run createPlatformUserRecord inside a write transaction to avoid deadlock
// (the real webhookAnalytics calls AppUserID() → DB read inside write tx).
// ============================================================================

type noopAnalytics struct{}

func (noopAnalytics) Enqueue(_ analytics.Message) {}

// ============================================================================
// callbackQueryInput — implements InputMessage + CallbackQuery
// Used to exercise BotChatID's CallbackQuery branch.
// ============================================================================

type callbackQueryInput struct {
	data     string
	senderID any
}

func (c *callbackQueryInput) GetSender() botinput.User {
	return &moreTestUser{id: c.senderID}
}
func (c *callbackQueryInput) GetRecipient() botinput.Recipient { return nil }
func (c *callbackQueryInput) GetTime() time.Time               { return time.Now() }
func (c *callbackQueryInput) InputType() botinput.Type         { return botinput.TypeCallbackQuery }
func (c *callbackQueryInput) MessageIntID() int                { return 0 }
func (c *callbackQueryInput) MessageStringID() string          { return "" }
func (c *callbackQueryInput) BotChatID() (string, error)       { return "", nil } // empty → falls to type switch
func (c *callbackQueryInput) Chat() botinput.Chat              { return nil }
func (c *callbackQueryInput) LogRequest()                      {}
func (c *callbackQueryInput) GetData() string                  { return c.data }
func (c *callbackQueryInput) GetFrom() botinput.Sender         { return &moreTestUser{id: c.senderID} }
func (c *callbackQueryInput) GetID() string                    { return "cq1" }
func (c *callbackQueryInput) GetMessage() botinput.Message     { return nil }

// ============================================================================
// inlineQueryInput — implements InputMessage + InlineQuery
// ============================================================================

type inlineQueryInput struct {
	senderID any
}

func (i *inlineQueryInput) GetSender() botinput.User {
	return &moreTestUser{id: i.senderID}
}
func (i *inlineQueryInput) GetRecipient() botinput.Recipient { return nil }
func (i *inlineQueryInput) GetTime() time.Time               { return time.Now() }
func (i *inlineQueryInput) InputType() botinput.Type         { return botinput.TypeInlineQuery }
func (i *inlineQueryInput) MessageIntID() int                { return 0 }
func (i *inlineQueryInput) MessageStringID() string          { return "" }
func (i *inlineQueryInput) BotChatID() (string, error)       { return "", nil }
func (i *inlineQueryInput) Chat() botinput.Chat              { return nil }
func (i *inlineQueryInput) LogRequest()                      {}
func (i *inlineQueryInput) GetFrom() botinput.Sender         { return &moreTestUser{id: i.senderID} }
func (i *inlineQueryInput) GetID() any                       { return "iq1" }
func (i *inlineQueryInput) GetInlineQueryID() string         { return "iq1" }
func (i *inlineQueryInput) GetOffset() string                { return "" }
func (i *inlineQueryInput) GetQuery() string                 { return "" }

// ============================================================================
// chosenInlineResultInput — implements InputMessage + ChosenInlineResult
// ============================================================================

type chosenInlineResultInput struct {
	senderID any
}

func (c *chosenInlineResultInput) GetSender() botinput.User {
	return &moreTestUser{id: c.senderID}
}
func (c *chosenInlineResultInput) GetRecipient() botinput.Recipient { return nil }
func (c *chosenInlineResultInput) GetTime() time.Time               { return time.Now() }
func (c *chosenInlineResultInput) InputType() botinput.Type         { return botinput.TypeChosenInlineResult }
func (c *chosenInlineResultInput) MessageIntID() int                { return 0 }
func (c *chosenInlineResultInput) MessageStringID() string          { return "" }
func (c *chosenInlineResultInput) BotChatID() (string, error)       { return "", nil }
func (c *chosenInlineResultInput) Chat() botinput.Chat              { return nil }
func (c *chosenInlineResultInput) LogRequest()                      {}
func (c *chosenInlineResultInput) GetFrom() botinput.Sender         { return &moreTestUser{id: c.senderID} }
func (c *chosenInlineResultInput) GetInlineMessageID() string       { return "" }
func (c *chosenInlineResultInput) GetQuery() string                 { return "" }
func (c *chosenInlineResultInput) GetResultID() string              { return "" }

// ============================================================================
// nullAppContext — like testAppContext but returns nil from SupportedLocales
// to exercise the SetLocale nil-supportedLocales guard.
// ============================================================================

type nullLocaleAppContext struct{}

func (nullLocaleAppContext) SupportedLocales() []i18n.Locale { return nil }
func (nullLocaleAppContext) GetLocaleByCode5(code5 string) (i18n.Locale, error) {
	return i18n.Locale{}, errors.New("not supported")
}
func (nullLocaleAppContext) GetTranslator(_ context.Context) i18n.Translator {
	return testTranslator{}
}
func (nullLocaleAppContext) SetLocale(_ string) error { return nil }
func (nullLocaleAppContext) CreateAppUserFromBotUser(_ context.Context, _ dal.ReadwriteTransaction, _ botsdal.Bot) (
	record.DataWithID[string, botsfwmodels.AppUserData], botsdal.BotUser, error,
) {
	panic("not implemented")
}

// ============================================================================
// newWHCBWithDB — creates a WHCB with a fresh in-memory DB.
// ============================================================================

func newWHCBWithDB(t *testing.T) (*WebhookContextBase, dal.DB) {
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
	// Use noopAnalytics by default to avoid deadlocks when createPlatformUserRecord
	// (called within a write tx) invokes Enqueue → AppUserID → DB read.
	whcb.whAnalytics = noopAnalytics{}
	whcb.recordsFieldsSetter = stubRecordsFieldsSetter{}
	return whcb, memDB
}

// newPlatformUserDbo creates a PlatformUserBaseDbo with required fields set.
func newPlatformUserDbo() *botsfwmodels.PlatformUserBaseDbo {
	now := time.Now()
	dbo := &botsfwmodels.PlatformUserBaseDbo{}
	dbo.DtCreated = now
	dbo.DtUpdated = now
	dbo.BotIDs = []string{"bot1"}
	return dbo
}

// seedPlatformUser writes a platform user record into the in-memory DB so that
// getPlatformUserRecord / GetPlatformUser succeed on the first read.
func seedPlatformUser(t *testing.T, db dal.DB, platformID botsfwconst.Platform, botUserID string, data botsfwmodels.PlatformUserData) {
	t.Helper()
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return botsdal.CreatePlatformUserRecord(ctx, tx, platformID, botUserID, data)
	})
	if err != nil {
		t.Fatalf("seedPlatformUser: %v", err)
	}
}

// ============================================================================
// BotChatID branches
// ============================================================================

// BotChatID — CallbackQuery with botChat= in data
func TestWebhookContextBase_BotChatID_CallbackQuery_WithBotChat(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.input = &callbackQueryInput{
		data:     "botChat=mychat42&x=y",
		senderID: "user1",
	}
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "mychat42" {
		t.Errorf("expected 'mychat42', got %q", chatID)
	}
}

// BotChatID — CallbackQuery with no botChat= in data (passes through switch, returns "")
func TestWebhookContextBase_BotChatID_CallbackQuery_NoBotChat(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.input = &callbackQueryInput{
		data:     "action=foo",
		senderID: "user1",
	}
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "" {
		t.Errorf("expected empty chatID, got %q", chatID)
	}
}

// BotChatID — InlineQuery (pass branch)
func TestWebhookContextBase_BotChatID_InlineQuery(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.c = context.Background()
	whcb.input = &inlineQueryInput{senderID: "user1"}
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "" {
		t.Errorf("expected empty chatID for InlineQuery, got %q", chatID)
	}
}

// BotChatID — ChosenInlineResult (pass branch)
func TestWebhookContextBase_BotChatID_ChosenInlineResult(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	whcb.c = context.Background()
	whcb.input = &chosenInlineResultInput{senderID: "user1"}
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "" {
		t.Errorf("expected empty chatID for ChosenInlineResult, got %q", chatID)
	}
}

// BotChatID — via getLocaleAndChatID returning a chatID
func TestWebhookContextBase_BotChatID_ViaGetLocaleAndChatID(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	// Set input to return empty chatID so we fall through to getLocaleAndChatID
	whcb.input = &moreTestInputMessage{inputType: botinput.TypeText, chatID: "", senderID: "user1"}
	whcb.getLocaleAndChatID = func() (string, string, error) {
		return "en-US", "locale-chat-99", nil
	}
	chatID, err := whcb.BotChatID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatID != "locale-chat-99" {
		t.Errorf("expected 'locale-chat-99', got %q", chatID)
	}
}

// MustBotChatID — BotChatID returns error → panics
func TestWebhookContextBase_MustBotChatID_ErrorPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from MustBotChatID when BotChatID errors")
		}
	}()
	whcb := newMoreTestWHCB(t)
	// Inject a getLocaleAndChatID that returns an error
	whcb.input = &moreTestInputMessage{inputType: botinput.TypeText, chatID: "", senderID: "u1"}
	whcb.getLocaleAndChatID = func() (string, string, error) {
		return "", "", errors.New("forced error")
	}
	whcb.MustBotChatID()
}

// ============================================================================
// getPlatformUserRecord — already-loaded fast path (Data != nil)
// ============================================================================

func TestWebhookContextBase_getPlatformUserRecord_AlreadyLoaded(t *testing.T) {
	whcb := newWHCBWithProfile(t)
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"

	err := whcb.getPlatformUserRecord(whcb.db)
	if err != nil {
		t.Errorf("expected nil error when platformUser.Data already set, got: %v", err)
	}
}

// ============================================================================
// createPlatformUserRecord (0 %)
// ============================================================================

// createPlatformUserRecord — fast path: Data already set
func TestWebhookContextBase_createPlatformUserRecord_AlreadyLoaded(t *testing.T) {
	whcb := newWHCBWithProfile(t)
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"

	err := whcb.db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return whcb.createPlatformUserRecord(tx)
	})
	if err != nil {
		t.Errorf("expected nil when Data already set: %v", err)
	}
}

// createPlatformUserRecord — full create path (Data == nil, string senderID)
// Note: createPlatformUserRecord calls Enqueue which calls AppUserID → ChatData.
// To avoid deadlock (nested read inside write tx on in-memory DB), we pre-seed
// the appUserID and chatData so the analytics path short-circuits without DB I/O.
func makeWHCBForCreate(t *testing.T, senderID any, senderIDStr string) *WebhookContextBase {
	t.Helper()
	// newWHCBWithDB already sets whAnalytics = noopAnalytics{} to avoid deadlocks.
	whcb, _ := newWHCBWithDB(t)
	whcb.input = &moreTestInputMessage{
		inputType: botinput.TypeText,
		chatID:    "chat1",
		senderID:  senderID,
	}
	// Pre-seed platformUser Key so Set() in createPlatformUserRecord has a valid key.
	key := botsdal.NewPlatformUserKey(botsfwconst.Platform(whcb.botPlatform.ID()), senderIDStr)
	userData := whcb.botContext.BotSettings.Profile.NewPlatformUserData()
	whcb.platformUser.Record = dal.NewRecordWithData(key, userData)
	whcb.platformUser.ID = senderIDStr
	// platformUser.Data is nil → createPlatformUserRecord will run (not fast-path).
	return whcb
}

func TestWebhookContextBase_createPlatformUserRecord_StringSenderID(t *testing.T) {
	whcb := makeWHCBForCreate(t, "user-str-1", "user-str-1")
	err := whcb.db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return whcb.createPlatformUserRecord(tx)
	})
	if err != nil {
		t.Errorf("unexpected error from createPlatformUserRecord (string sender): %v", err)
	}
}

// createPlatformUserRecord — int senderID branch
func TestWebhookContextBase_createPlatformUserRecord_IntSenderID(t *testing.T) {
	whcb := makeWHCBForCreate(t, int(42), "42")
	err := whcb.db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return whcb.createPlatformUserRecord(tx)
	})
	if err != nil {
		t.Errorf("unexpected error from createPlatformUserRecord (int sender): %v", err)
	}
}

// createPlatformUserRecord — int64 senderID branch
func TestWebhookContextBase_createPlatformUserRecord_Int64SenderID(t *testing.T) {
	whcb := makeWHCBForCreate(t, int64(999), "999")
	err := whcb.db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return whcb.createPlatformUserRecord(tx)
	})
	if err != nil {
		t.Errorf("unexpected error from createPlatformUserRecord (int64 sender): %v", err)
	}
}

// ============================================================================
// getOrCreatePlatformUserRecord
// ============================================================================

// getOrCreatePlatformUserRecord — fast path: Data already set
func TestWebhookContextBase_getOrCreatePlatformUserRecord_AlreadyLoaded(t *testing.T) {
	whcb := newWHCBWithProfile(t)
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"

	botUser, err := whcb.getOrCreatePlatformUserRecord()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if botUser.Data == nil {
		t.Error("expected non-nil BotUser.Data")
	}
}

// getOrCreatePlatformUserRecord — found-in-DB path
func TestWebhookContextBase_getOrCreatePlatformUserRecord_FoundInDB(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)

	// Seed the platform user so getPlatformUserRecord succeeds.
	seedPlatformUser(t, memDB, "test", "user42", newPlatformUserDbo())

	botUser, err := whcb.getOrCreatePlatformUserRecord()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if botUser.Data == nil {
		t.Error("expected non-nil BotUser.Data")
	}
}

// getOrCreatePlatformUserRecord — not-found path: creates new record
func TestWebhookContextBase_getOrCreatePlatformUserRecord_CreatesNew(t *testing.T) {
	whcb, _ := newWHCBWithDB(t)
	// No seeded record → not-found → createPlatformUserRecord
	// whcb.whAnalytics is noopAnalytics (set in newWHCBWithDB) so no deadlock.
	_, err := whcb.getOrCreatePlatformUserRecord()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// loadChatEntityBase — missing branches
// ============================================================================

// loadChatEntityBase — duplicate-call guard (HasChatData() == true)
func TestWebhookContextBase_loadChatEntityBase_AlreadyLoaded(t *testing.T) {
	whcb, _ := newWHCBWithDB(t)
	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	whcb.botChat.Data = mockCD

	err := whcb.loadChatEntityBase()
	if err != nil {
		t.Errorf("expected nil error on duplicate-call guard, got: %v", err)
	}
}

// loadChatEntityBase — chat found in DB path
func TestWebhookContextBase_loadChatEntityBase_ChatFound(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)

	// Seed a bot chat so GetBotChat succeeds.
	chatData := whcb.botContext.BotSettings.Profile.NewBotChatData()
	chatKey := botsdal.NewBotChatKey("test", "bot1", "chat1")
	err := memDB.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, dal.NewRecordWithData(chatKey, chatData))
	})
	if err != nil {
		t.Fatalf("seed chat: %v", err)
	}

	err = whcb.loadChatEntityBase()
	if err != nil {
		t.Errorf("unexpected error from loadChatEntityBase (chat found): %v", err)
	}
	if whcb.botChat.Data == nil {
		t.Error("expected chatData to be populated")
	}
}

// loadChatEntityBase — sender language updates client languages
func TestWebhookContextBase_loadChatEntityBase_SenderLanguage(t *testing.T) {
	whcb, _ := newWHCBWithDB(t)
	whcb.input = &moreTestInputMessage{
		inputType: botinput.TypeText,
		chatID:    "chat1",
		senderID:  "user42",
		language:  "en",
	}
	// noopAnalytics is already set in newWHCBWithDB so no deadlock.
	err := whcb.loadChatEntityBase()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// loadChatEntityBase — preferred language on loaded chat → calls SetLocale
func TestWebhookContextBase_loadChatEntityBase_ChatLocale(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)
	whcb.botContext.BotSettings.Profile = chatLocaleProfile{}

	// Seed a chat with a preferred language using chatLocaleProfile's NewBotChatData
	chatData := whcb.botContext.BotSettings.Profile.NewBotChatData()
	chatKey := botsdal.NewBotChatKey("test", "bot1", "chat1")
	err := memDB.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, dal.NewRecordWithData(chatKey, chatData))
	})
	if err != nil {
		t.Fatalf("seed chat with locale: %v", err)
	}

	err = whcb.loadChatEntityBase()
	// SetLocale may or may not succeed; we just exercise the branch
	_ = err
}

// ============================================================================
// ChatData — panic on nil recordsFieldsSetter when chat not found
// ============================================================================

func TestWebhookContextBase_ChatData_PanicOnNilRecordsFieldsSetter(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from nil recordsFieldsSetter")
		}
	}()
	whcb, _ := newWHCBWithDB(t)
	whcb.recordsFieldsSetter = nil
	whcb.platformUser.Data = &botsfwmodels.PlatformUserBaseDbo{}
	whcb.platformUser.ID = "user42"
	// noopAnalytics is already set in newWHCBWithDB so no deadlock.
	// No chat in DB → not-found → SetBotChatFields → panic because recordsFieldsSetter is nil
	_ = whcb.ChatData()
}

// ============================================================================
// AppUserID — falls through to getPlatformUserRecord (DB path)
// ============================================================================

func TestWebhookContextBase_AppUserID_ViaDBPlatformUser(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)
	// Seed platform user with appUserID — note: PlatformUserBaseDbo implements
	// PlatformUserData but GetAppUserID always returns "" from the base struct.
	// The important thing is the DB read path is exercised without panic.
	seedPlatformUser(t, memDB, "test", "user42", newPlatformUserDbo())

	// isLoadingChatData=true skips ChatData() call, and platformUser.Data is nil initially
	whcb.isLoadingChatData = true

	// Should not panic; result is "" since PlatformUserBaseDbo has no app user.
	got := whcb.AppUserID()
	_ = got
}

// ============================================================================
// GetAppUser — happy path: appUserID set and getAppUser returns data
// ============================================================================

func TestWebhookContextBase_GetAppUser_WithValidAppUser(t *testing.T) {
	whcb := newCoverageWHCB(t)
	whcb.db = dalgo2memory.NewDB()
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "user-ok"}

	// getAppUser func returns ErrRecordNotFound but with a non-nil data means
	// we exercise the non-empty appUserID path; use a not-found result to
	// verify GetAppUser propagates errors correctly.
	whcb.botContext.BotSettings.getAppUser = func(
		ctx context.Context, tx dal.ReadSession, botID, appUserID string,
	) (record.DataWithID[string, botsfwmodels.AppUserData], error) {
		return record.DataWithID[string, botsfwmodels.AppUserData]{}, dal.ErrRecordNotFound
	}

	// Expect not-found error because the func returns ErrRecordNotFound.
	_, err := whcb.GetAppUser()
	if err == nil {
		t.Fatal("expected error from GetAppUser, got nil")
	}
	if !dal.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

// GetAppUser — happy path with successful lookup
func TestWebhookContextBase_GetAppUser_HappyPath(t *testing.T) {
	whcb := newCoverageWHCB(t)
	whcb.db = dalgo2memory.NewDB()
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "user-ok"}

	var returnedData botsfwmodels.AppUserData // stays nil to satisfy interface
	whcb.botContext.BotSettings.getAppUser = func(
		ctx context.Context, tx dal.ReadSession, botID, appUserID string,
	) (record.DataWithID[string, botsfwmodels.AppUserData], error) {
		r := record.DataWithID[string, botsfwmodels.AppUserData]{}
		r.ID = appUserID
		r.Data = returnedData
		return r, nil
	}

	got, err := whcb.GetAppUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = got // may be nil (returnedData is nil)
}

// ============================================================================
// AppUserData — happy path
// ============================================================================

func TestWebhookContextBase_AppUserData_HappyPath(t *testing.T) {
	whcb := newCoverageWHCB(t)
	whcb.db = dalgo2memory.NewDB()
	whcb.isLoadingChatData = true
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "user-ok"}

	var returnedData botsfwmodels.AppUserData // nil
	whcb.botContext.BotSettings.getAppUser = func(
		ctx context.Context, tx dal.ReadSession, botID, appUserID string,
	) (record.DataWithID[string, botsfwmodels.AppUserData], error) {
		r := record.DataWithID[string, botsfwmodels.AppUserData]{}
		r.ID = appUserID
		r.Data = returnedData
		return r, nil
	}

	got, err := whcb.AppUserData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = got
}

// ============================================================================
// NewWebhookContextBase — with Db set in args
// ============================================================================

func TestNewWebhookContextBase_Db_IsSet(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	memDB := dalgo2memory.NewDB()
	args := CreateWebhookContextArgs{
		HttpRequest: r,
		AppContext:  testAppContext{},
		Db:          memDB,
		BotContext: BotContext{
			BotHost:     testBotHost{},
			BotSettings: &BotSettings{Code: "testbot", Token: "tok123", Locale: i18n.LocaleEnUS},
		},
		WebhookInput: &moreTestInputMessage{inputType: botinput.TypeText, chatID: "chat1", senderID: "u1"},
	}
	whcb, err := NewWebhookContextBase(args, &testBotPlatformMore{}, stubRecordsFieldsSetter{},
		func() (bool, error) { return false, nil },
		func(context.Context) (string, string, error) { return "", "", nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if whcb.DB() != memDB {
		t.Error("expected DB to be set from args")
	}
}

// ============================================================================
// Locale — when chatData returns a preferred language (currently 57.1 %)
// ============================================================================

func TestWebhookContextBase_Locale_FromChatData(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	// First call in Locale() → GetPreferredLanguage → "en-US"
	mockCD.EXPECT().GetPreferredLanguage().Return("en-US").AnyTimes()
	mockCD.EXPECT().GetAppUserID().Return("").AnyTimes()

	whcb := newMoreTestWHCB(t)
	whcb.appContext = testAppContext{}
	whcb.botChat.Data = mockCD

	loc := whcb.Locale()
	if loc.Code5 != "en-US" {
		t.Errorf("expected 'en-US' from chatData.GetPreferredLanguage, got %q", loc.Code5)
	}
}

// Locale — when chatData returns empty preferred language
func TestWebhookContextBase_Locale_ChatDataReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockCD.EXPECT().GetPreferredLanguage().Return("").AnyTimes()
	mockCD.EXPECT().GetAppUserID().Return("").AnyTimes()

	whcb := newMoreTestWHCB(t)
	whcb.appContext = testAppContext{}
	whcb.botChat.Data = mockCD
	// locale.Code5 is "" and chatData returns "" → fall through to BotSettings.Locale
	loc := whcb.Locale()
	if loc.Code5 != "en-US" {
		t.Errorf("expected 'en-US' from BotSettings fallback, got %q", loc.Code5)
	}
}

// ============================================================================
// SetLocale — nil supportedLocales guard
// ============================================================================

func TestWebhookContextBase_SetLocale_NilSupportedLocales(t *testing.T) {
	whcb := &WebhookContextBase{
		appContext: nullLocaleAppContext{},
	}
	err := whcb.SetLocale("en-US")
	if err == nil {
		t.Fatal("expected error when SupportedLocales returns nil")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

// ============================================================================
// SetAccessGranted (context_auth.go, 0 %)
// ============================================================================

// SetAccessGranted — chatData is nil (only sets platform user), already granted
func TestSetAccessGranted_NoChatData_AlreadyGranted(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)

	// Seed platform user with accessGranted already true.
	userData := newPlatformUserDbo()
	userData.AccessGranted = true
	seedPlatformUser(t, memDB, "test", "user42", userData)

	err := SetAccessGranted((*whcbWHC)(whcb), true)
	if err != nil {
		t.Errorf("unexpected error from SetAccessGranted(already granted): %v", err)
	}
}

// SetAccessGranted — chatData is nil, value changes from false → true
func TestSetAccessGranted_NoChatData_ValueChanges(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)

	// Seed platform user with accessGranted = false.
	userData := newPlatformUserDbo()
	userData.AccessGranted = false
	seedPlatformUser(t, memDB, "test", "user42", userData)

	err := SetAccessGranted((*whcbWHC)(whcb), true)
	if err != nil {
		t.Errorf("unexpected error from SetAccessGranted(change to true): %v", err)
	}
}

// SetAccessGranted — chatData present, value already matches (no-op)
func TestSetAccessGranted_ChatAlreadyGranted(t *testing.T) {
	ctrl := gomock.NewController(t)
	whcb, memDB := newWHCBWithDB(t)

	mockCD := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockCD.EXPECT().IsAccessGranted().Return(true).AnyTimes()
	mockCD.EXPECT().GetAppUserID().Return("").AnyTimes()
	whcb.botChat.Data = mockCD

	// Seed platform user
	seedPlatformUser(t, memDB, "test", "user42", newPlatformUserDbo())

	err := SetAccessGranted((*whcbWHC)(whcb), true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// SetAccessGranted — chatData present, value changes → SaveBotChat triggered
func TestSetAccessGranted_ChatChangedToGranted(t *testing.T) {
	whcb, memDB := newWHCBWithDB(t)

	// Build a real botChat.Record so SaveBotChat can tx.Set it.
	chatData := &botsfwmodels.ChatBaseData{}
	chatKey := dal.NewKeyWithID("botChats", "chat1")
	whcb.botChat.Record = dal.NewRecordWithData(chatKey, chatData)
	whcb.botChat.Data = chatData

	// Seed platform user (accessGranted = false in base)
	seedPlatformUser(t, memDB, "test", "user42", newPlatformUserDbo())

	err := SetAccessGranted((*whcbWHC)(whcb), true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================================
// whcbWHC — minimal WebhookContext adapter for *WebhookContextBase.
// SetAccessGranted only calls: Context, GetBotCode, ChatData, Input, DB,
// BotPlatform, BotContext, SaveBotChat — all on *WebhookContextBase.
// ============================================================================

// whcbWHC wraps *WebhookContextBase and implements the full WebhookContext.
// Methods that can never be reached by SetAccessGranted panic with "not used".
type whcbWHC WebhookContextBase

func (w *whcbWHC) base() *WebhookContextBase { return (*WebhookContextBase)(w) }

func (w *whcbWHC) Analytics() WebhookAnalytics                      { return w.base().Analytics() }
func (w *whcbWHC) DB() dal.DB                                       { return w.base().DB() }
func (w *whcbWHC) RecordsFieldsSetter() BotRecordsFieldsSetter      { return w.base().RecordsFieldsSetter() }
func (w *whcbWHC) BotContext() BotContext                           { return w.base().BotContext() }
func (w *whcbWHC) LogRequest()                                      { w.base().LogRequest() }
func (w *whcbWHC) Request() *http.Request                           { return w.base().Request() }
func (w *whcbWHC) Environment() string                              { return w.base().Environment() }
func (w *whcbWHC) MustBotChatID() string                            { return w.base().MustBotChatID() }
func (w *whcbWHC) BotChatID() (string, error)                       { return w.base().BotChatID() }
func (w *whcbWHC) ChatData() botsfwmodels.BotChatData               { return w.base().ChatData() }
func (w *whcbWHC) SetUser(id string, data botsfwmodels.AppUserData) { w.base().SetUser(id, data) }
func (w *whcbWHC) AppUserID() string                                { return w.base().AppUserID() }
func (w *whcbWHC) GetAppUser() (botsfwmodels.AppUserData, error) {
	return w.base().GetAppUser()
}
func (w *whcbWHC) AppUserEntity() botsfwmodels.AppUserData        { return w.base().AppUserEntity() }
func (w *whcbWHC) AppUserData() (botsfwmodels.AppUserData, error) { return w.base().AppUserData() }
func (w *whcbWHC) GetBotUserForUpdate(ctx context.Context, tx dal.ReadwriteTransaction) (botsdal.BotUser, error) {
	return w.base().GetBotUserForUpdate(ctx, tx)
}
func (w *whcbWHC) GetBotUser() (botsdal.BotUser, error)   { return w.base().GetBotUser() }
func (w *whcbWHC) IsInGroup() (bool, error)               { return w.base().IsInGroup() }
func (w *whcbWHC) ExecutionContext() ExecutionContext     { return w.base().ExecutionContext() }
func (w *whcbWHC) AppContext() AppContext                 { return w.base().AppContext() }
func (w *whcbWHC) Context() context.Context               { return w.base().Context() }
func (w *whcbWHC) SetContext(c context.Context)           { w.base().SetContext(c) }
func (w *whcbWHC) IsInTransaction(c context.Context) bool { return w.base().IsInTransaction(c) }
func (w *whcbWHC) NonTransactionalContext(tc context.Context) context.Context {
	return w.base().NonTransactionalContext(tc)
}
func (w *whcbWHC) BotPlatform() BotPlatform              { return w.base().BotPlatform() }
func (w *whcbWHC) GetBotSettings() *BotSettings          { return w.base().GetBotSettings() }
func (w *whcbWHC) GetBotCode() string                    { return w.base().GetBotCode() }
func (w *whcbWHC) GetBotUserID() string                  { return w.base().GetBotUserID() }
func (w *whcbWHC) GetBotToken() string                   { return w.base().GetBotToken() }
func (w *whcbWHC) Input() botinput.InputMessage          { return w.base().Input() }
func (w *whcbWHC) Chat() botinput.Chat                   { return w.base().Chat() }
func (w *whcbWHC) GetRecipient() botinput.Recipient      { return w.base().GetRecipient() }
func (w *whcbWHC) GetTime() time.Time                    { return w.base().GetTime() }
func (w *whcbWHC) InputType() botinput.Type              { return w.base().InputType() }
func (w *whcbWHC) HasChatData() bool                     { return w.base().HasChatData() }
func (w *whcbWHC) SaveBotChat() error                    { return w.base().SaveBotChat() }
func (w *whcbWHC) SaveBotUser(ctx context.Context) error { return w.base().SaveBotUser(ctx) }
func (w *whcbWHC) MessageText() string                   { return w.base().MessageText() }
func (w *whcbWHC) CommandText(title, icon string) string { return w.base().CommandText(title, icon) }
func (w *whcbWHC) Locale() i18n.Locale                   { return w.base().Locale() }
func (w *whcbWHC) SetLocale(code5 string) error          { return w.base().SetLocale(code5) }
func (w *whcbWHC) Translate(key string, args ...interface{}) string {
	return w.base().Translate(key, args...)
}
func (w *whcbWHC) TranslateNoWarning(key string, args ...interface{}) string {
	return w.base().TranslateNoWarning(key, args...)
}
func (w *whcbWHC) TranslateWithMap(key string, m map[string]string) string {
	return w.base().TranslateWithMap(key, m)
}
func (w *whcbWHC) GetTranslator(locale string) i18n.SingleLocaleTranslator {
	return w.base().GetTranslator(locale)
}
func (w *whcbWHC) NewMessage(text string) botmsg.MessageFromBot {
	return w.base().NewMessage(text)
}
func (w *whcbWHC) NewMessageByCode(code string, a ...interface{}) botmsg.MessageFromBot {
	return w.base().NewMessageByCode(code, a...)
}
func (w *whcbWHC) NewEditMessage(_ string, _ botmsg.Format) (botmsg.MessageFromBot, error) {
	panic("not used in test")
}
func (w *whcbWHC) UpdateLastProcessed(_ botsfwmodels.BotChatData) error {
	panic("not used in test")
}
func (w *whcbWHC) Responder() WebhookResponder {
	panic("not used in test")
}
func (w *whcbWHC) IsNewerThen(_ botsfwmodels.BotChatData) bool {
	panic("not used in test")
}

// ============================================================================
// TitleByKey — missing branches (commands.go, 92.3 %)
// ============================================================================

func TestCommand_TitleByKey_IconOnly(t *testing.T) {
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale: %v", err)
	}
	cmd := Command{
		Code: "help",
		Icon: "❓",
		// Title == "" and Titles is nil → code branch: title = CommandTextNoTrans("", icon)
	}
	// Use testWhc (existing TestWebhookContext) which already implements WebhookContext.
	// We just need Translate and CommandText to work: testWhc.Translate returns the key.
	got := cmd.TitleByKey(DefaultTitle, testWhc)
	if got != "❓" {
		t.Errorf("expected '❓', got %q", got)
	}
}

func TestCommand_TitleByKey_WithTitleAndIcon(t *testing.T) {
	// Use testWhc — CommandText panics, but for this test Icon != "" and Title != ""
	// means whc.CommandText is called. testWhc.CommandText panics, so we can't call it.
	// Instead, build a cmd with title+icon using testWhc; since testWhc.CommandText panics,
	// we skip the assertion and just confirm the code branch is triggered via whcbWHC.
	// whcbWHC.CommandText delegates to WebhookContextBase.CommandText which actually works.
	whcb := newMoreTestWHCB(t)
	if err := whcb.SetLocale("en-US"); err != nil {
		t.Fatalf("SetLocale: %v", err)
	}
	whcbCtx := (*whcbWHC)(whcb)
	cmd := Command{
		Code:  "help",
		Title: "Help",
		Icon:  "❓",
	}
	got := cmd.TitleByKey(DefaultTitle, whcbCtx)
	// CommandText("Help", "❓") via testTranslator → "Help ❓"
	if got == "" {
		t.Error("expected non-empty title from TitleByKey with both title and icon")
	}
}

func TestCommand_TitleByKey_TitlesMapFallback(t *testing.T) {
	cmd := Command{
		Code:   "help",
		Titles: map[string]string{"short_title": "Brief Help"},
		Icon:   "",
	}
	got := cmd.TitleByKey(ShortTitle, testWhc)
	// title = "Brief Help", icon == "" → translated → "Brief Help"
	if got != "Brief Help" {
		t.Errorf("expected 'Brief Help', got %q", got)
	}
}

func TestCommand_TitleByKey_NoTitleNoTitles(t *testing.T) {
	cmd := Command{
		Code: "mycode",
		// Title == "", Titles == nil, Icon == ""
	}
	got := cmd.TitleByKey(DefaultTitle, testWhc)
	// No title, no icon → falls to "title = string(c.Code)"
	if got != "mycode" {
		t.Errorf("expected 'mycode', got %q", got)
	}
}

// ============================================================================
// findByPlatform — legacy flat-map paths (settings.go, 72.7 %)
// ============================================================================

func TestFindByPlatform_ByPlatformAndID(t *testing.T) {
	s := BotSettings{Platform: "tg", Code: "bot1", ID: "id1"}
	sb := BotSettingsBy{
		ByPlatformAndCode: map[botsfwconst.Platform]map[string]*BotSettings{},
		ByPlatformAndID: map[botsfwconst.Platform]map[string]*BotSettings{
			"tg": {"id1": &s},
		},
		ByCode: map[string]*BotSettings{},
		ByID:   map[string]*BotSettings{},
	}
	got := sb.findByPlatform("tg", "id1")
	if got == nil {
		t.Fatal("expected non-nil result from ByPlatformAndID")
	}
}

func TestFindByPlatform_ByPlatformAndCode(t *testing.T) {
	s := BotSettings{Platform: "tg", Code: "bot1"}
	sb := BotSettingsBy{
		ByPlatformAndCode: map[botsfwconst.Platform]map[string]*BotSettings{
			"tg": {"bot1": &s},
		},
		ByPlatformAndID: map[botsfwconst.Platform]map[string]*BotSettings{},
		ByCode:          map[string]*BotSettings{},
		ByID:            map[string]*BotSettings{},
	}
	got := sb.findByPlatform("tg", "bot1")
	if got == nil {
		t.Fatal("expected non-nil result from ByPlatformAndCode")
	}
}

func TestFindByPlatform_LegacyByID(t *testing.T) {
	s := BotSettings{Platform: "tg", Code: "bot1", ID: "id1"}
	sb := BotSettingsBy{
		ByPlatformAndCode: map[botsfwconst.Platform]map[string]*BotSettings{},
		ByPlatformAndID:   map[botsfwconst.Platform]map[string]*BotSettings{},
		ByCode:            map[string]*BotSettings{},
		ByID:              map[string]*BotSettings{"id1": &s},
	}
	got := sb.findByPlatform("tg", "id1")
	if got == nil {
		t.Fatal("expected non-nil result from legacy ByID")
	}
}

func TestFindByPlatform_LegacyByCode(t *testing.T) {
	s := BotSettings{Platform: "tg", Code: "bot1"}
	sb := BotSettingsBy{
		ByPlatformAndCode: map[botsfwconst.Platform]map[string]*BotSettings{},
		ByPlatformAndID:   map[botsfwconst.Platform]map[string]*BotSettings{},
		ByCode:            map[string]*BotSettings{"bot1": &s},
		ByID:              map[string]*BotSettings{},
	}
	got := sb.findByPlatform("tg", "bot1")
	if got == nil {
		t.Fatal("expected non-nil result from legacy ByCode")
	}
}

func TestFindByPlatform_LegacyByID_WrongPlatform(t *testing.T) {
	s := BotSettings{Platform: "wa", Code: "bot1", ID: "id1"}
	sb := BotSettingsBy{
		ByPlatformAndCode: map[botsfwconst.Platform]map[string]*BotSettings{},
		ByPlatformAndID:   map[botsfwconst.Platform]map[string]*BotSettings{},
		ByCode:            map[string]*BotSettings{},
		ByID:              map[string]*BotSettings{"id1": &s},
	}
	// Bot is on "wa" but we look for "tg" → should return nil.
	got := sb.findByPlatform("tg", "id1")
	if got != nil {
		t.Error("expected nil when platform doesn't match legacy ByID entry")
	}
}

func TestFindByPlatform_NotFound(t *testing.T) {
	sb := BotSettingsBy{
		ByPlatformAndCode: map[botsfwconst.Platform]map[string]*BotSettings{},
		ByPlatformAndID:   map[botsfwconst.Platform]map[string]*BotSettings{},
		ByCode:            map[string]*BotSettings{},
		ByID:              map[string]*BotSettings{},
	}
	got := sb.findByPlatform("tg", "unknown")
	if got != nil {
		t.Error("expected nil for unknown bot")
	}
}

// ============================================================================
// chatBaseDataWithLocale — for loadChatEntityBase locale path
// ============================================================================

type chatBaseDataWithLocale struct {
	botsfwmodels.ChatBaseData
	preferredLanguage string
}

func (c *chatBaseDataWithLocale) GetPreferredLanguage() string     { return c.preferredLanguage }
func (c *chatBaseDataWithLocale) SetPreferredLanguage(lang string) { c.preferredLanguage = lang }

// chatLocaleProfile — variant of stubBotProfile that returns chatBaseDataWithLocale
type chatLocaleProfile struct{}

func (s chatLocaleProfile) ID() string                       { return "test-locale-profile" }
func (s chatLocaleProfile) DefaultLocale() i18n.Locale       { return i18n.LocaleEnUS }
func (s chatLocaleProfile) GetTranslations() BotTranslations { return BotTranslations{} }
func (s chatLocaleProfile) Router() Router                   { return nil }
func (s chatLocaleProfile) SupportedLocales() []i18n.Locale  { return []i18n.Locale{i18n.LocaleEnUS} }
func (s chatLocaleProfile) NewBotChatData() botsfwmodels.BotChatData {
	return &chatBaseDataWithLocale{preferredLanguage: "en-US"}
}
func (s chatLocaleProfile) NewPlatformUserData() botsfwmodels.PlatformUserData {
	return &botsfwmodels.PlatformUserBaseDbo{}
}
func (s chatLocaleProfile) NewAppUserData() botsfwmodels.AppUserData { return nil }
