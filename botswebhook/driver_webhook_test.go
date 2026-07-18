package botswebhook

// driver_webhook_test.go: tests for HandleWebhook, processWebhookInput, and
// RegisterWebhookHandlers — the three functions in driver.go that were at 0%
// before this file was added.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	mock_dal "github.com/dal-go/dalgo/mocks/mock_dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

// ─── simple BotHost for tests ─────────────────────────────────────────────────

type webhookTestBotHost struct{}

func (webhookTestBotHost) Context(r *http.Request) context.Context { return r.Context() }
func (webhookTestBotHost) GetHTTPClient(_ context.Context) *http.Client {
	return http.DefaultClient
}

// newMockIMWithSender creates a MockInputMessage that returns a non-nil sender.
// logInput() in driver.go calls input.GetSender() and then accesses sender.GetID()
// etc. — so we need a non-nil sender to avoid a nil-pointer panic inside logInput.
// logInputSender is defined in coverage_extra_test.go (same package).
func newMockIMWithSender(ctrl *gomock.Controller, inputType botinput.Type) *mock_botinput.MockInputMessage {
	im := mock_botinput.NewMockInputMessage(ctrl)
	im.EXPECT().InputType().Return(inputType).AnyTimes()
	im.EXPECT().LogRequest().AnyTimes()
	im.EXPECT().Chat().Return(nil).AnyTimes()
	sender := logInputSender{id: "u1", firstName: "Test", lastName: "User"}
	im.EXPECT().GetSender().Return(sender).AnyTimes()
	im.EXPECT().GetRecipient().Return(nil).AnyTimes()
	im.EXPECT().GetTime().AnyTimes()
	im.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	im.EXPECT().MessageIntID().Return(0).AnyTimes()
	im.EXPECT().MessageStringID().Return("").AnyTimes()
	return im
}

// ─── helper: build a BotContext with a mock profile + mock DB ────────────────

func makeTestBotContext(profile botsfw.BotProfile, db dal.DB, env string) *botsfw.BotContext {
	return &botsfw.BotContext{
		BotSettings: &botsfw.BotSettings{
			Env:  env,
			Code: "testbot",
			GetDatabase: func(_ context.Context) (dal.DB, error) {
				return db, nil
			},
			Profile: profile,
		},
	}
}

func makeTestBotContextWithDBErr(profile botsfw.BotProfile, dbErr error, env string) *botsfw.BotContext {
	return &botsfw.BotContext{
		BotSettings: &botsfw.BotSettings{
			Env:  env,
			Code: "testbot",
			GetDatabase: func(_ context.Context) (dal.DB, error) {
				return nil, dbErr
			},
			Profile: profile,
		},
	}
}

// ─── RegisterWebhookHandlers ──────────────────────────────────────────────────

// TestRegisterWebhookHandlers_CallsRegisterHttpHandlers verifies that
// RegisterWebhookHandlers calls RegisterHttpHandlers on each handler.
func TestRegisterWebhookHandlers_CallsRegisterHttpHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	d := webhookDriver{botHost: mockHost}

	handler1 := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler2 := mock_botsfw.NewMockWebhookHandler(ctrl)

	// Both handlers should get RegisterHttpHandlers called exactly once.
	handler1.EXPECT().RegisterHttpHandlers(d, mockHost, nil, "/bot").Times(1)
	handler2.EXPECT().RegisterHttpHandlers(d, mockHost, nil, "/bot").Times(1)

	d.RegisterWebhookHandlers(nil, "/bot", handler1, handler2)
}

// TestRegisterWebhookHandlers_NoHandlers verifies that calling with zero handlers
// is a no-op (doesn't panic).
func TestRegisterWebhookHandlers_NoHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	d := webhookDriver{botHost: mockHost}

	// Should not panic.
	d.RegisterWebhookHandlers(nil, "/bot")
}

// ─── HandleWebhook — nil-parameter panics ────────────────────────────────────

func TestHandleWebhook_NilW_Panics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	d := webhookDriver{botHost: mockHost}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when w == nil")
		}
	}()
	r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	d.HandleWebhook(nil, r, handler)
}

func TestHandleWebhook_NilR_Panics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	d := webhookDriver{botHost: mockHost}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when r == nil")
		}
	}()
	w := httptest.NewRecorder()
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	d.HandleWebhook(w, nil, handler)
}

func TestHandleWebhook_NilHandler_Panics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	d := webhookDriver{botHost: mockHost}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when webhookHandler == nil")
		}
	}()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	d.HandleWebhook(w, r, nil)
}

// ─── HandleWebhook — invalidContextOrInputs paths ─────────────────────────────

// TestHandleWebhook_GetBotContextAndInputsError covers the branch where
// GetBotContextAndInputs returns an error → invalidContextOrInputs returns true
// → HandleWebhook returns early.
func TestHandleWebhook_GetBotContextAndInputsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).
		Return(nil, nil, errors.New("parsing failed"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)
}

// TestHandleWebhook_NilBotContext covers nil botContext returned alongside nil entries.
func TestHandleWebhook_NilBotContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).
		Return(nil, nil, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)
}

// TestHandleWebhook_EmptyEntries covers the path where GetBotContextAndInputs returns
// a valid botContext but zero entries → the loop body never executes.
func TestHandleWebhook_EmptyEntries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	bc := &botsfw.BotContext{
		BotSettings: &botsfw.BotSettings{
			Env:     botsfw.EnvProduction,
			Code:    "testbot",
			Profile: profile,
		},
	}

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).
		Return(bc, []botinput.EntryInputs{}, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)
}

// ─── processWebhookInput — CreateWebhookContext error ────────────────────────

// TestProcessWebhookInput_CreateWebhookContextError exercises the branch where
// CreateWebhookContext returns an error → handleError is called → function returns.
func TestProcessWebhookInput_CreateWebhookContextError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	router := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(router).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	// newMockIMWithSender ensures logInput won't panic on nil sender.
	im := newMockIMWithSender(ctrl, botinput.TypeText)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	createErr := errors.New("cannot create context")
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(nil, createErr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	var handleErrorCalled bool
	handleError := func(err error, msg string) {
		handleErrorCalled = true
		http.Error(w, msg, http.StatusInternalServerError)
	}

	d := webhookDriver{botHost: webhookTestBotHost{}}
	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err == nil {
		t.Fatal("expected error from processWebhookInput, got nil")
	}
	if !handleErrorCalled {
		t.Fatal("expected handleError to be called on CreateWebhookContext failure")
	}
}

// ─── processWebhookInput — happy path (chatData non-nil, AppUserID already set) ──

// TestProcessWebhookInput_HappyPath drives a successful round-trip through
// processWebhookInput.  chatData.GetAppUserID() returns a non-empty string so
// the transaction branch is skipped.
func TestProcessWebhookInput_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	whc, responder, wAnalytics, chatData := setupDispatchWHC(ctrl, im)
	chatData.EXPECT().GetAppUserID().Return("user1").AnyTimes()
	wAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	cmdRouter := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(cmdRouter).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	handleError := func(err error, msg string) {
		t.Errorf("unexpected handleError call: %v - %s", err, msg)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── processWebhookInput — chatData nil path ─────────────────────────────────

// TestProcessWebhookInput_NilChatData exercises the branch where ChatData() returns
// nil (e.g. inline queries) — the transaction block is skipped entirely.
// We build the whc manually (not via setupDispatchWHC) so ChatData returns nil
// from the start.  We use TypeVoice (an unknown/unhandled type) so that
// logInputDetails in router.go doesn't perform a direct type assertion that would
// panic on our plain MockInputMessage.
func TestProcessWebhookInput_NilChatData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeVoice)

	// Build a minimal whc with ChatData() == nil.
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(nil).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	wAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(wAnalytics).AnyTimes()
	wAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	whc.EXPECT().SaveBotChat().Return(nil).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(text string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
	}).AnyTimes()

	platform := mock_botsfw.NewMockBotPlatform(ctrl)
	platform.EXPECT().ID().Return("test").AnyTimes()
	whc.EXPECT().BotPlatform().Return(platform).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	cmdRouter := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(cmdRouter).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	handleError := func(err error, msg string) {
		t.Errorf("unexpected handleError call: %v - %s", err, msg)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── HandleWebhook — CallbackQuery error → HTTP 200 ──────────────────────────

// TestHandleWebhook_CallbackQueryError_Returns200 verifies that when a processing
// error occurs for a TypeCallbackQuery input, the HTTP response is 200 (not 500).
func TestHandleWebhook_CallbackQueryError_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	// Use newMockIMWithSender so logInput() doesn't panic on nil sender.
	im := newMockIMWithSender(ctrl, botinput.TypeCallbackQuery)

	mockDB := mock_dal.NewMockDB(ctrl)
	profile := mock_botsfw.NewMockBotProfile(ctrl)
	router := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(router).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	createErr := errors.New("callback context creation failed")
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).
		Return(bc, []botinput.EntryInputs{{Inputs: []botinput.InputMessage{im}}}, nil)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(nil, createErr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)

	// For TypeCallbackQuery, the handleError func calls w.WriteHeader(200).
	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for callback-query error, got %d", w.Code)
	}
}

// TestHandleWebhook_NonCallbackQueryError_Returns500 verifies that when a processing
// error occurs for a non-callback-query input, the HTTP response is 500.
func TestHandleWebhook_NonCallbackQueryError_Returns500(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	// Use newMockIMWithSender so logInput() doesn't panic on nil sender.
	im := newMockIMWithSender(ctrl, botinput.TypeText)

	mockDB := mock_dal.NewMockDB(ctrl)
	profile := mock_botsfw.NewMockBotProfile(ctrl)
	router := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(router).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	createErr := errors.New("text context creation failed")
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).
		Return(bc, []botinput.EntryInputs{{Inputs: []botinput.InputMessage{im}}}, nil)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(nil, createErr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)

	// For non-callback-query, the handleError func calls http.Error(w, ..., 500).
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected HTTP 500 for non-callback-query error, got %d", w.Code)
	}
}

// ─── processWebhookInput — panic recovery path ───────────────────────────────

// TestProcessWebhookInput_PanicRecovery verifies that a panic inside
// processWebhookInput is caught by the deferred recover and does not propagate.
// We use EnvLocal so the analytics path is not triggered.
func TestProcessWebhookInput_PanicRecovery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	router := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(router).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	// newMockIMWithSender so logInput doesn't panic on nil sender before we even
	// get to CreateWebhookContext.
	im := newMockIMWithSender(ctrl, botinput.TypeText)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).DoAndReturn(
		func(_ botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
			panic("simulated panic inside CreateWebhookContext")
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}, panicTextFooter: "test-footer"}
	handleError := func(err error, msg string) {}

	// Must not panic — the deferred recover inside processWebhookInput catches it.
	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	_ = err
}

// ─── processWebhookInput — panic recovery WITH analytics.Enabled ─────────────

// TestProcessWebhookInput_PanicRecovery_AnalyticsEnabled covers the branch where
// Analytics.Enabled returns true → the analytics-enabled code path is exercised.
// whc is still nil at panic time so the "if whc != nil" branch is NOT taken.
func TestProcessWebhookInput_PanicRecovery_AnalyticsEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	router := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(router).AnyTimes()

	// Use EnvLocal so Env != EnvProduction; we override with Analytics.Enabled.
	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).DoAndReturn(
		func(_ botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
			panic("simulated panic for analytics test")
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	analyticsEnabledCalled := false
	d := webhookDriver{
		botHost:         webhookTestBotHost{},
		panicTextFooter: "analytics-footer",
		Analytics: AnalyticsSettings{
			// Return false so that reportPanicToAnalytics is not called
			// (whc == nil so it would panic inside reportPanicToAnalytics too).
			Enabled: func(_ *http.Request) bool {
				analyticsEnabledCalled = true
				return false
			},
		},
	}
	handleError := func(err error, msg string) {}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	_ = err

	if !analyticsEnabledCalled {
		t.Error("expected Analytics.Enabled to be called during panic recovery")
	}
}

// ─── HandleWebhook — happy path end-to-end ───────────────────────────────────

// TestHandleWebhook_HappyPath drives HandleWebhook through a complete single-entry
// single-input dispatch cycle with no errors.  ChatData.GetAppUserID() returns
// "user1" so no transaction is needed.
func TestHandleWebhook_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	whc, responder, wAnalytics, chatData := setupDispatchWHC(ctrl, im)
	chatData.EXPECT().GetAppUserID().Return("user1").AnyTimes()
	wAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	mockDB := mock_dal.NewMockDB(ctrl)
	profile := mock_botsfw.NewMockBotProfile(ctrl)
	cmdRouter := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(cmdRouter).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).
		Return(bc, []botinput.EntryInputs{{Inputs: []botinput.InputMessage{im}}}, nil)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)
}

// ─── HandleWebhook — multiple entries log path (len > 1) ─────────────────────

// TestHandleWebhook_MultipleEntries drives the len(entriesWithInputs) > 1 log
// branch in HandleWebhook.  We short-circuit processWebhookInput via a
// CreateWebhookContext error so we don't need a full dispatch setup.
func TestHandleWebhook_MultipleEntries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHost := mock_botsfw.NewMockBotHost(ctrl)
	mockHost.EXPECT().Context(gomock.Any()).Return(context.Background()).AnyTimes()

	d := webhookDriver{botHost: mockHost}

	// Use newMockIMWithSender so logInput doesn't panic on nil sender.
	im1 := newMockIMWithSender(ctrl, botinput.TypeText)
	im2 := newMockIMWithSender(ctrl, botinput.TypeText)

	mockDB := mock_dal.NewMockDB(ctrl)
	profile := mock_botsfw.NewMockBotProfile(ctrl)
	router := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(router).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	createErr := errors.New("short circuit")
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().GetBotContextAndInputs(gomock.Any(), gomock.Any()).Return(bc,
		[]botinput.EntryInputs{
			{Inputs: []botinput.InputMessage{im1}},
			{Inputs: []botinput.InputMessage{im2}},
		}, nil)
	// CreateWebhookContext called twice (once per input).
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(nil, createErr).Times(2)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)

	d.HandleWebhook(w, r, handler)
}

// ─── processWebhookInput — chatData non-nil with existing AppUserID (skip tx) ──

// TestProcessWebhookInput_ChatData_ExistingUser verifies that when ChatData is non-nil
// and GetAppUserID() returns a non-empty string, the transaction path is skipped.
func TestProcessWebhookInput_ChatData_ExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeCallbackQuery)

	whc, responder, wAnalytics, chatData := setupDispatchWHC(ctrl, im)
	chatData.EXPECT().GetAppUserID().Return("existing-user").AnyTimes()
	wAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	cmdRouter := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(cmdRouter).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	handleError := func(err error, msg string) {
		t.Errorf("unexpected handleError: %v - %s", err, msg)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── processWebhookInput — nil input panics (recovered) ─────────────────────

// TestProcessWebhookInput_NilInput covers line 179: `if input == nil { panic(...) }`
// The panic is caught by the deferred recover.  We use EnvLocal so no analytics
// call is made on the nil whc.
func TestProcessWebhookInput_NilInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)
	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	handleError := func(err error, msg string) {}

	// nil input triggers panic inside processWebhookInput; recover catches it.
	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, nil, handleError)
	_ = err
}

// ─── processWebhookInput — GetDatabase error ─────────────────────────────────

// TestProcessWebhookInput_GetDatabaseError covers lines 184-187: when GetDatabase
// returns an error the function returns early with the wrapped error.
func TestProcessWebhookInput_GetDatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbErr := errors.New("db unavailable")
	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()

	bc := makeTestBotContextWithDBErr(profile, dbErr, botsfw.EnvProduction)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	// CreateWebhookContext should NOT be called since GetDatabase fails first.

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}

	var handleErrorCalled bool
	handleError := func(err error, msg string) {
		handleErrorCalled = true
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	// GetDatabase errors are returned directly (not via handleError).
	if err == nil {
		t.Fatal("expected error from GetDatabase failure, got nil")
	}
	if handleErrorCalled {
		t.Error("handleError should not be called for GetDatabase failures")
	}
}

// ─── processWebhookInput — new user transaction path ─────────────────────────

// webhookTestAppContext is a minimal AppContext implementation for testing the
// transaction path in processWebhookInput.  It satisfies botsfw.AppContext.
type webhookTestAppContext struct{}

func (webhookTestAppContext) SupportedLocales() []i18n.Locale { return nil }
func (webhookTestAppContext) GetLocaleByCode5(_ string) (i18n.Locale, error) {
	return i18n.Locale{}, nil
}
func (webhookTestAppContext) GetTranslator(_ context.Context) i18n.Translator {
	return webhookTestTranslator{}
}
func (webhookTestAppContext) SetLocale(_ string) error { return nil }
func (webhookTestAppContext) CreateAppUserFromBotUser(
	_ context.Context, _ dal.ReadwriteTransaction, _ botsdal.Bot,
) (
	record.DataWithID[string, botsfwmodels.AppUserData],
	botsdal.BotUser,
	error,
) {
	// Return zero-value records (nil Record) so the tx.Insert loop is skipped.
	return record.DataWithID[string, botsfwmodels.AppUserData]{
			RecordWithID: dal.RecordWithID[string]{ID: "newuser"},
		},
		botsdal.BotUser{},
		nil
}

type webhookTestTranslator struct{}

func (webhookTestTranslator) Translate(key, _ string, _ ...any) string { return key }
func (webhookTestTranslator) TranslateWithMap(key, _ string, _ map[string]string) string {
	return key
}
func (webhookTestTranslator) TranslateNoWarning(key, _ string, _ ...any) string { return key }

// TestProcessWebhookInput_NewUserTransaction covers lines 196-239: when chatData is
// non-nil and GetAppUserID() == "", the code runs RunReadwriteTransaction to
// create a new app user.
func TestProcessWebhookInput_NewUserTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := mock_dal.NewMockReadwriteTransaction(ctrl)
	// The transaction body returns a DataWithID with nil Record, so Insert is
	// never called.  If it were, we'd need mockTx.EXPECT().Insert(...).

	mockDB := mock_dal.NewMockDB(ctrl)
	mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f dal.RWTxWorker, opts ...dal.TransactionOption) error {
			return f(ctx, mockTx)
		},
	)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("").AnyTimes()
	chatData.EXPECT().SetAppUserID(gomock.Any()).AnyTimes()
	chatData.EXPECT().GetAwaitingReplyTo().Return("").AnyTimes()
	chatData.EXPECT().IsChanged().Return(false).AnyTimes()
	chatData.EXPECT().HasChangedVars().Return(false).AnyTimes()
	chatData.EXPECT().SetDtLastInteraction(gomock.Any()).AnyTimes()
	chatData.EXPECT().SetUpdatedTime(gomock.Any()).AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	wAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(wAnalytics).AnyTimes()
	wAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	whc.EXPECT().SaveBotChat().Return(nil).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().SetUser(gomock.Any(), gomock.Any()).AnyTimes()

	platform := mock_botsfw.NewMockBotPlatform(ctrl)
	platform.EXPECT().ID().Return("tg").AnyTimes()
	whc.EXPECT().BotPlatform().Return(platform).AnyTimes()

	// Return a concrete AppContext (not a mock) so CreateAppUserFromBotUser works.
	whc.EXPECT().AppContext().Return(webhookTestAppContext{}).AnyTimes()

	// Locale/translate expectations for Dispatch.
	whc.EXPECT().Locale().AnyTimes()
	whc.EXPECT().SetLocale(gomock.Any()).Return(nil).AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().TranslateNoWarning(gomock.Any(), gomock.Any()).Return("").AnyTimes()
	whc.EXPECT().CommandText(gomock.Any(), gomock.Any()).Return("").AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(t string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: t}}
	}).AnyTimes()
	whc.EXPECT().NewMessageByCode(gomock.Any(), gomock.Any()).DoAndReturn(func(code string, _ ...interface{}) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: code}}
	}).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	cmdRouter := NewWebhookRouter(nil)
	profile.EXPECT().Router().Return(cmdRouter).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	handleError := func(err error, msg string) {
		t.Errorf("unexpected handleError: %v - %s", err, msg)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestProcessWebhookInput_NewUserTransaction_TxError covers the branch where
// RunReadwriteTransaction itself returns an error → handleError is called.
func TestProcessWebhookInput_NewUserTransaction_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txErr := errors.New("transaction failed")
	mockDB := mock_dal.NewMockDB(ctrl)
	mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).Return(txErr)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("").AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	whc.EXPECT().BotPlatform().Return(mock_botsfw.NewMockBotPlatform(ctrl)).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().AppContext().Return(webhookTestAppContext{}).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()
	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	var handleErrorCalled bool
	handleError := func(err error, msg string) {
		handleErrorCalled = true
		http.Error(w, msg, http.StatusInternalServerError)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err == nil {
		t.Fatal("expected error from transaction failure, got nil")
	}
	if !handleErrorCalled {
		t.Error("expected handleError to be called on transaction failure")
	}
}

// ─── processWebhookInput — panic recovery with whc != nil ────────────────────

// TestProcessWebhookInput_PanicWithWhcSet covers lines 157-175: whc is set before
// the panic (CreateWebhookContext succeeds), then GetResponder panics.
// The recover block: whc != nil → BotChatID() → SendMessageThroughGate.
func TestProcessWebhookInput_PanicWithWhcSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeText)
	im.EXPECT().BotChatID().Return("chat123", nil).AnyTimes()

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("existingUser").AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	// SendMessage is called during panic recovery.
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(t string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: t}}
	}).AnyTimes()
	whcAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whcAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	whc.EXPECT().Analytics().Return(whcAnalytics).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	// GetResponder panics after whc is set.
	handler.EXPECT().GetResponder(gomock.Any(), whc).DoAndReturn(
		func(_ http.ResponseWriter, _ botsfw.WebhookContext) botsfw.WebhookResponder {
			panic("simulated panic after whc is set")
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}, panicTextFooter: "test-footer"}
	handleError := func(err error, msg string) {}

	// Must not propagate the panic (recover catches it).
	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	_ = err
}

// TestProcessWebhookInput_PanicWithWhcSet_SendNotPermitted covers the
// IsSendNotPermitted branch in the whc != nil recovery path.
func TestProcessWebhookInput_PanicWithWhcSet_SendNotPermitted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeText)
	im.EXPECT().BotChatID().Return("chat123", nil).AnyTimes()

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("existingUser").AnyTimes()

	// Responder that returns a "send not permitted" error.
	responder := &gatedTestResponder{refuse: fmt.Errorf("gated: %w", botsfw.ErrSendNotPermitted)}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(t string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: t}}
	}).AnyTimes()
	whcAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whcAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	whc.EXPECT().Analytics().Return(whcAnalytics).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).DoAndReturn(
		func(_ http.ResponseWriter, _ botsfw.WebhookContext) botsfw.WebhookResponder {
			panic("simulated panic: send not permitted path")
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}, panicTextFooter: "test-footer"}
	handleError := func(err error, msg string) {}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	_ = err
}

// TestProcessWebhookInput_PanicWithWhcSet_SendError covers line 170:
// SendMessageThroughGate returns a non-IsSendNotPermitted error.
func TestProcessWebhookInput_PanicWithWhcSet_SendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeText)
	im.EXPECT().BotChatID().Return("chat123", nil).AnyTimes()

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("existingUser").AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	// SendMessage returns a generic (non-IsSendNotPermitted) error.
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, errors.New("network error")).AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(t string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: t}}
	}).AnyTimes()
	whcAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whcAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	whc.EXPECT().Analytics().Return(whcAnalytics).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).DoAndReturn(
		func(_ http.ResponseWriter, _ botsfw.WebhookContext) botsfw.WebhookResponder {
			panic("simulated panic: send error path")
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}, panicTextFooter: "test-footer"}
	handleError := func(err error, msg string) {}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	_ = err
}

// ─── processWebhookInput — panic-text truncation (>3KB) ──────────────────────

// TestProcessWebhookInput_PanicRecovery_TruncatesLongMessage covers lines 145-147:
// the panic-text truncation branch.  We trigger a panic whose *value* is a >3KB
// string, guaranteeing len(messageText) > 3*1024 (messageText prepends
// "Panic: " and appends a stack trace + footer, so a 4KB recovered value is more
// than enough).  The recover() is reached because CreateWebhookContext panics.
// The test passing (no re-panic) IS the assertion that the truncation slice is
// in-bounds.  Env is local and Analytics is disabled, so it takes the
// non-analytics branch (line 154) and whc is still nil (no send path).
func TestProcessWebhookInput_PanicRecovery_TruncatesLongMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_dal.NewMockDB(ctrl)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()

	// EnvLocal + no Analytics.Enabled → non-analytics branch at line 154.
	bc := makeTestBotContext(profile, mockDB, botsfw.EnvLocal)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	// Panic with a >3KB string value so messageText exceeds maxLen (3*1024).
	bigPanic := strings.Repeat("x", 4096)
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).DoAndReturn(
		func(_ botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
			panic(bigPanic)
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}, panicTextFooter: "trunc-footer"}
	handleError := func(err error, msg string) {}

	// If the truncation slice were out of bounds it would panic here (unrecovered).
	// Reaching this point without a second panic proves the branch executed safely.
	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	_ = err
}

// ─── processWebhookInput — configurable AppContext for tx-body branches ──────

// webhookConfigurableAppContext lets each test control what CreateAppUserFromBotUser
// returns, so we can cover: the error branch (214), the appUser.Record != nil append
// (217), the botUser.Record != nil append (220) and the tx.Insert loop (228-231).
type webhookConfigurableAppContext struct {
	appUser record.DataWithID[string, botsfwmodels.AppUserData]
	botUser botsdal.BotUser
	err     error
}

func (webhookConfigurableAppContext) SupportedLocales() []i18n.Locale { return nil }
func (webhookConfigurableAppContext) GetLocaleByCode5(_ string) (i18n.Locale, error) {
	return i18n.Locale{}, nil
}
func (webhookConfigurableAppContext) GetTranslator(_ context.Context) i18n.Translator {
	return webhookTestTranslator{}
}
func (webhookConfigurableAppContext) SetLocale(_ string) error { return nil }
func (c webhookConfigurableAppContext) CreateAppUserFromBotUser(
	_ context.Context, _ dal.ReadwriteTransaction, _ botsdal.Bot,
) (
	record.DataWithID[string, botsfwmodels.AppUserData],
	botsdal.BotUser,
	error,
) {
	return c.appUser, c.botUser, c.err
}

// buildTxWHC constructs a WebhookContext mock wired for the transaction path, with a
// dispatch-friendly router.  appCtx controls CreateAppUserFromBotUser behavior.
func buildTxWHC(
	ctrl *gomock.Controller,
	im *mock_botinput.MockInputMessage,
	appCtx botsfw.AppContext,
) (*mock_botsfw.MockWebhookContext, *mock_botsfw.MockWebhookResponder) {
	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("").AnyTimes()
	chatData.EXPECT().SetAppUserID(gomock.Any()).AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().AppContext().Return(appCtx).AnyTimes()
	whc.EXPECT().SetUser(gomock.Any(), gomock.Any()).AnyTimes()

	platform := mock_botsfw.NewMockBotPlatform(ctrl)
	platform.EXPECT().ID().Return("tg").AnyTimes()
	whc.EXPECT().BotPlatform().Return(platform).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	wAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(wAnalytics).AnyTimes()
	wAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(tx string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: tx}}
	}).AnyTimes()
	return whc, responder
}

// mkAppUserRecord builds a non-nil app-user DataWithID with a populated .Record.
func mkAppUserRecord(id string) record.DataWithID[string, botsfwmodels.AppUserData] {
	key := dal.NewKeyWithID("appUsers", id)
	return record.DataWithID[string, botsfwmodels.AppUserData]{
		RecordWithID: dal.RecordWithID[string]{
			ID:     id,
			Key:    key,
			Record: dal.NewRecordWithData(key, &struct{ Name string }{Name: id}),
		},
	}
}

// mkBotUserRecord builds a non-nil bot-user record.
func mkBotUserRecord(id string) botsdal.BotUser {
	key := dal.NewKeyWithID("botUsers", id)
	return botsdal.BotUser{
		RecordWithID: dal.RecordWithID[string]{
			ID:     id,
			Key:    key,
			Record: dal.NewRecordWithData(key, &struct{ Name string }{Name: id}),
		},
	}
}

// TestProcessWebhookInput_NewUserTx_CreateAppUserError covers line 214-216: when
// CreateAppUserFromBotUser returns an error, the transaction body returns that
// error, RunReadwriteTransaction propagates it, and handleError is called.
func TestProcessWebhookInput_NewUserTx_CreateAppUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createErr := errors.New("create app user failed")

	mockTx := mock_dal.NewMockReadwriteTransaction(ctrl)
	mockDB := mock_dal.NewMockDB(ctrl)
	mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f dal.RWTxWorker, opts ...dal.TransactionOption) error {
			return f(ctx, mockTx)
		},
	)

	im := newMockIMWithSender(ctrl, botinput.TypeText)
	appCtx := webhookConfigurableAppContext{err: createErr}
	whc, _ := buildTxWHC(ctrl, im, appCtx)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()
	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	var handleErrorCalled bool
	handleError := func(err error, msg string) {
		handleErrorCalled = true
		http.Error(w, msg, http.StatusInternalServerError)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err == nil {
		t.Fatal("expected error from CreateAppUserFromBotUser failure, got nil")
	}
	if !handleErrorCalled {
		t.Error("expected handleError to be called")
	}
}

// TestProcessWebhookInput_NewUserTx_InsertsRecords covers lines 217-218, 220-221 and
// 228-231: both appUser.Record and botUser.Record are non-nil so both are appended
// and inserted via tx.Insert (which succeeds).
func TestProcessWebhookInput_NewUserTx_InsertsRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := mock_dal.NewMockReadwriteTransaction(ctrl)
	// Two records → two successful inserts.
	mockTx.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	mockDB := mock_dal.NewMockDB(ctrl)
	mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f dal.RWTxWorker, opts ...dal.TransactionOption) error {
			return f(ctx, mockTx)
		},
	)

	im := newMockIMWithSender(ctrl, botinput.TypeText)
	appCtx := webhookConfigurableAppContext{
		appUser: mkAppUserRecord("newuser"),
		botUser: mkBotUserRecord("botuser"),
	}
	whc, responder := buildTxWHC(ctrl, im, appCtx)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()
	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	handleError := func(err error, msg string) {
		t.Errorf("unexpected handleError: %v - %s", err, msg)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestProcessWebhookInput_NewUserTx_InsertError covers line 229-231: tx.Insert
// returns an error → the transaction body returns it → handleError is called.
func TestProcessWebhookInput_NewUserTx_InsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	insertErr := errors.New("insert failed")
	mockTx := mock_dal.NewMockReadwriteTransaction(ctrl)
	// First Insert fails, loop returns immediately.
	mockTx.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(insertErr)

	mockDB := mock_dal.NewMockDB(ctrl)
	mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f dal.RWTxWorker, opts ...dal.TransactionOption) error {
			return f(ctx, mockTx)
		},
	)

	im := newMockIMWithSender(ctrl, botinput.TypeText)
	appCtx := webhookConfigurableAppContext{
		appUser: mkAppUserRecord("newuser"),
	}
	whc, _ := buildTxWHC(ctrl, im, appCtx)

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(NewWebhookRouter(nil)).AnyTimes()
	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	var handleErrorCalled bool
	handleError := func(err error, msg string) {
		handleErrorCalled = true
		http.Error(w, msg, http.StatusInternalServerError)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err == nil {
		t.Fatal("expected error from tx.Insert failure, got nil")
	}
	if !handleErrorCalled {
		t.Error("expected handleError to be called on insert failure")
	}
}

// ─── processWebhookInput — Dispatch error path ───────────────────────────────

// errRouter is a stub botsfw.Router whose Dispatch always returns an error, used to
// cover the Dispatch-error branch (lines 244-247) in processWebhookInput.
type errRouter struct{ err error }

func (errRouter) RegisterCommands(_ ...botsfw.Command)                              {}
func (errRouter) RegisterCommandsForInputType(_ botinput.Type, _ ...botsfw.Command) {}
func (r errRouter) Dispatch(_ botsfw.WebhookHandler, _ botsfw.WebhookResponder, _ botsfw.WebhookContext) error {
	return r.err
}
func (errRouter) RegisteredCommands() map[botinput.Type]map[botsfw.CommandCode]botsfw.Command {
	return nil
}
func (errRouter) SetFallbackHandler(_ botinput.Type, _ botsfw.CommandAction) {}

// TestProcessWebhookInput_DispatchError covers lines 244-247: router.Dispatch returns
// an error → handleError("Failed to dispatch") is called and the error is returned.
func TestProcessWebhookInput_DispatchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dispatchErr := errors.New("dispatch boom")

	mockDB := mock_dal.NewMockDB(ctrl)

	im := newMockIMWithSender(ctrl, botinput.TypeText)

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAppUserID().Return("existing-user").AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()

	profile := mock_botsfw.NewMockBotProfile(ctrl)
	profile.EXPECT().Router().Return(errRouter{err: dispatchErr}).AnyTimes()

	bc := makeTestBotContext(profile, mockDB, botsfw.EnvProduction)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().CreateWebhookContext(gomock.Any()).Return(whc, nil)
	handler.EXPECT().GetResponder(gomock.Any(), whc).Return(responder)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
	ctx := context.Background()

	d := webhookDriver{botHost: webhookTestBotHost{}}
	var handleErrorCalled bool
	var gotMsg string
	handleError := func(err error, msg string) {
		handleErrorCalled = true
		gotMsg = msg
		http.Error(w, msg, http.StatusInternalServerError)
	}

	err := d.processWebhookInput(ctx, w, r, handler, bc, 0, im, handleError)
	if err == nil {
		t.Fatal("expected error from Dispatch failure, got nil")
	}
	if !handleErrorCalled {
		t.Error("expected handleError to be called on Dispatch failure")
	}
	if gotMsg != "Failed to dispatch" {
		t.Errorf("expected handleError message 'Failed to dispatch', got %q", gotMsg)
	}
}
