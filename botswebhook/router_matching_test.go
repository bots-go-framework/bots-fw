package botswebhook

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

// --- helpers ---

func dummyAction(_ botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyCallbackAction(_ botsfw.WebhookContext, _ *url.URL) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyTextAction(_ botsfw.WebhookContext, _ string) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyLocationAction(_ botsfw.WebhookContext, _, _ float64) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyInlineQueryAction(_ botsfw.WebhookContext, _ botinput.InlineQuery, _ *url.URL) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyChosenInlineResultAction(_ botsfw.WebhookContext, _ botinput.ChosenInlineResult, _ *url.URL) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyPreCheckoutQueryAction(_ botsfw.WebhookContext, _ botinput.PreCheckoutQuery) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummySuccessfulPaymentAction(_ botsfw.WebhookContext, _ botinput.SuccessfulPayment) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

func dummyRefundedPaymentAction(_ botsfw.WebhookContext, _ botinput.RefundedPayment) (botmsg.MessageFromBot, error) {
	return botmsg.MessageFromBot{}, nil
}

// setupBasicWHC sets up a MockWebhookContext with Context() and Input().LogRequest() for matchCallbackCommands tests
func setupBasicWHC(ctrl *gomock.Controller) *mock_botsfw.MockWebhookContext {
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()
	mockInput := mock_botinput.NewMockInputMessage(ctrl)
	mockInput.EXPECT().LogRequest().AnyTimes()
	mockWHC.EXPECT().Input().Return(mockInput).AnyTimes()
	return mockWHC
}

// setupMessageWHC sets up a MockWebhookContext suitable for matchMessageCommands tests
func setupMessageWHC(ctrl *gomock.Controller, awaitingReplyTo string) *mock_botsfw.MockWebhookContext {
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockChatData.EXPECT().GetAwaitingReplyTo().Return(awaitingReplyTo).AnyTimes()
	mockWHC.EXPECT().ChatData().Return(mockChatData).AnyTimes()

	mockWHC.EXPECT().TranslateNoWarning(gomock.Any(), gomock.Any()).Return("").AnyTimes()
	mockWHC.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, args ...interface{}) string { return key },
	).AnyTimes()
	mockWHC.EXPECT().CommandText(gomock.Any(), gomock.Any()).DoAndReturn(
		func(title, icon string) string { return title + " " + icon },
	).AnyTimes()
	return mockWHC
}

// --- matchCallbackCommands tests ---

func TestMatchCallbackCommands_ByURLPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := setupBasicWHC(ctrl)

	// matchCallbackCommands looks up commands[CommandCode(dataURL.Path)]
	// URL path for "settings" is "settings" (no leading slash in a relative URL)
	commands := map[botsfw.CommandCode]botsfw.Command{
		"settings": {
			Code:           "settings",
			CallbackAction: dummyCallbackAction,
		},
	}
	dataURL, _ := url.Parse("settings")
	matched, err := matchCallbackCommands(mockWHC, "settings", dataURL, commands)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matched == nil {
		t.Fatal("expected a matched command, got nil")
	} else if matched.Code != "settings" {
		t.Errorf("expected code 'settings', got %q", matched.Code)
	}
}

func TestMatchCallbackCommands_NoMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := setupBasicWHC(ctrl)

	commands := map[botsfw.CommandCode]botsfw.Command{
		"settings": {
			Code:           "settings",
			CallbackAction: dummyCallbackAction,
		},
	}
	dataURL, _ := url.Parse("/unknown")
	matched, err := matchCallbackCommands(mockWHC, "/unknown", dataURL, commands)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matched != nil {
		t.Errorf("expected nil, got command %q", matched.Code)
	}
}

func TestMatchCallbackCommands_ByMatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := setupBasicWHC(ctrl)

	commands := map[botsfw.CommandCode]botsfw.Command{
		"custom": {
			Code: "custom",
			Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
				return true
			},
			CallbackAction: dummyCallbackAction,
		},
	}
	dataURL, _ := url.Parse("/nomatch")
	matched, err := matchCallbackCommands(mockWHC, "/nomatch", dataURL, commands)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matched == nil {
		t.Fatal("expected a matched command via Matcher, got nil")
	} else if matched.Code != "custom" {
		t.Errorf("expected code 'custom', got %q", matched.Code)
	}
}

// --- matchNonTextCommands tests ---

func TestMatchNonTextCommands_EmptyAwaitingReply_SingleEmptyCode(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)

	commands := []botsfw.Command{{Code: "", Action: dummyAction}}
	matched := router.matchNonTextCommands(mockWHC, "", commands)
	if matched == nil {
		t.Fatal("expected match for single empty-code command with empty awaitingReplyTo")
	}
}

func TestMatchNonTextCommands_EmptyAwaitingReply_MultipleCommands(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)

	commands := []botsfw.Command{
		{Code: "a", Action: dummyAction},
		{Code: "b", Action: dummyAction},
	}
	matched := router.matchNonTextCommands(mockWHC, "", commands)
	if matched != nil {
		t.Errorf("expected nil for multiple commands with empty awaitingReplyTo, got %q", matched.Code)
	}
}

func TestMatchNonTextCommands_AwaitingReplyMatchesCode(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)

	commands := []botsfw.Command{
		{Code: "loc_cmd", Action: dummyAction},
	}
	matched := router.matchNonTextCommands(mockWHC, "loc_cmd", commands)
	if matched == nil {
		t.Fatal("expected match by code")
	} else if matched.Code != "loc_cmd" {
		t.Errorf("expected code 'loc_cmd', got %q", matched.Code)
	}
}

func TestMatchNonTextCommands_AwaitingReplyWithQueryParams(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)

	commands := []botsfw.Command{
		{Code: "loc_cmd", Action: dummyAction},
	}
	matched := router.matchNonTextCommands(mockWHC, "loc_cmd?param=1", commands)
	if matched == nil {
		t.Fatal("expected match after trimming query params")
	} else if matched.Code != "loc_cmd" {
		t.Errorf("expected code 'loc_cmd', got %q", matched.Code)
	}
}

func TestMatchNonTextCommands_AwaitingReplyViaMatcher(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)

	commands := []botsfw.Command{
		{
			Code: "matchable",
			Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
				return true
			},
			Action: dummyAction,
		},
	}
	matched := router.matchNonTextCommands(mockWHC, "something_else", commands)
	if matched == nil {
		t.Fatal("expected match via Matcher")
	} else if matched.Code != "matchable" {
		t.Errorf("expected code 'matchable', got %q", matched.Code)
	}
}

func TestMatchNonTextCommands_AwaitingReplyNoMatch(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)

	commands := []botsfw.Command{
		{Code: "other", Action: dummyAction},
	}
	matched := router.matchNonTextCommands(mockWHC, "nomatch", commands)
	if matched != nil {
		t.Errorf("expected nil, got %q", matched.Code)
	}
}

// --- matchMessageCommands tests ---

func TestMatchMessageCommands_ByCode(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "help", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, true, "/help", "", commands)
	if matched == nil {
		t.Fatal("expected match by code /help")
	} else if matched.Code != "help" {
		t.Errorf("expected code 'help', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_ByCodeWithBotName(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "help", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, true, "/help@mybot", "", commands)
	if matched == nil {
		t.Fatal("expected match by code /help@mybot")
	} else if matched.Code != "help" {
		t.Errorf("expected code 'help', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_ByExactMatch(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "greet", ExactMatch: "Hello!", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "Hello!", "", commands)
	if matched == nil {
		t.Fatal("expected match by ExactMatch")
	} else if matched.Code != "greet" {
		t.Errorf("expected code 'greet', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_ByDefaultTitle(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	// Command with no Title and no Icon: DefaultTitle returns string(c.Code)
	commands := []botsfw.Command{
		{Code: "mycommand", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "mycommand", "", commands)
	if matched == nil {
		t.Fatal("expected match by DefaultTitle (code as title)")
	} else if matched.Code != "mycommand" {
		t.Errorf("expected code 'mycommand', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_ByDefaultTitle_WithTitleTranslation(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	// Command with Title and no Icon: DefaultTitle calls whc.Translate(title) → returns title unchanged (our mock)
	commands := []botsfw.Command{
		{Code: "info", Title: "Information", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "Information", "", commands)
	if matched == nil {
		t.Fatal("expected match by DefaultTitle with Title translation")
	} else if matched.Code != "info" {
		t.Errorf("expected code 'info', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_ByMatcher(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{
			Code: "custom",
			Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
				return true
			},
			TextAction: dummyTextAction,
		},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "anythingunknown", "", commands)
	if matched == nil {
		t.Fatal("expected match by Matcher")
	} else if matched.Code != "custom" {
		t.Errorf("expected code 'custom', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_AwaitingReplyTo(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "reply_cmd")

	commands := []botsfw.Command{
		{Code: "reply_cmd", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "some user text", "", commands)
	if matched == nil {
		t.Fatal("expected match via awaitingReplyTo")
	} else if matched.Code != "reply_cmd" {
		t.Errorf("expected code 'reply_cmd', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_NoMatch(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "help", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "random text", "", commands)
	if matched != nil {
		t.Errorf("expected nil, got %q", matched.Code)
	}
}

func TestMatchMessageCommands_StartAction(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "start", TextAction: dummyTextAction},
		{
			Code:        "ref123",
			StartAction: dummyTextAction,
		},
	}
	// "/start ref123" should match the command with code "ref123" and StartAction
	matched := router.matchMessageCommands(mockWHC, nil, true, "/start ref123", "", commands)
	if matched == nil {
		t.Fatal("expected match for /start <code> with StartAction")
	} else if matched.Code != "ref123" {
		t.Errorf("expected code 'ref123', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_StartFallsBackToStartCommand(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "start", TextAction: dummyTextAction},
	}
	// "/start unknown" – the "start" command matches by code prefix "/start ...",
	// but "unknown" doesn't match any StartAction code; it falls back to the start command.
	matched := router.matchMessageCommands(mockWHC, nil, true, "/start unknown", "", commands)
	if matched == nil {
		t.Fatal("expected fallback match to start command")
	} else if matched.Code != "start" {
		t.Errorf("expected code 'start', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_CommandWithSpace(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "help", TextAction: dummyTextAction},
	}
	// "/help argument" should match command with Code "help" since it starts with "/help "
	matched := router.matchMessageCommands(mockWHC, nil, true, "/help argument", "", commands)
	if matched == nil {
		t.Fatal("expected match for /help with argument")
	} else if matched.Code != "help" {
		t.Errorf("expected code 'help', got %q", matched.Code)
	}
}

// --- changeLocaleIfLangPassed tests ---

func TestChangeLocaleIfLangPassed_NoLangParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	u, _ := url.Parse("/settings")
	m, err := changeLocaleIfLangPassed(mockWHC, u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty text, got %q", m.Text)
	}
}

func TestChangeLocaleIfLangPassed_EnExpandedToEnUS(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockWHC.EXPECT().ChatData().Return(mockChatData).AnyTimes()

	mockWHC.EXPECT().Locale().Return(i18n.Locale{Code5: "ru-RU"}).AnyTimes()
	mockWHC.EXPECT().SetLocale("en-US").Return(nil)
	mockChatData.EXPECT().SetPreferredLanguage("en-US")

	u, _ := url.Parse("/settings?l=en")
	m, err := changeLocaleIfLangPassed(mockWHC, u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty text, got %q", m.Text)
	}
}

func TestChangeLocaleIfLangPassed_FaExpandedToFaIR(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockWHC.EXPECT().ChatData().Return(mockChatData).AnyTimes()

	mockWHC.EXPECT().Locale().Return(i18n.Locale{Code5: "en-US"}).AnyTimes()
	mockWHC.EXPECT().SetLocale("fa-IR").Return(nil)
	mockChatData.EXPECT().SetPreferredLanguage("fa-IR")

	u, _ := url.Parse("/settings?l=fa")
	m, err := changeLocaleIfLangPassed(mockWHC, u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty text, got %q", m.Text)
	}
}

func TestChangeLocaleIfLangPassed_SameLocaleNoChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockWHC.EXPECT().ChatData().Return(mockChatData).AnyTimes()

	mockWHC.EXPECT().Locale().Return(i18n.Locale{Code5: "en-US"}).AnyTimes()
	// SetLocale should NOT be called since lang == currentLocale

	u, _ := url.Parse("/settings?l=en-US")
	m, err := changeLocaleIfLangPassed(mockWHC, u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty text, got %q", m.Text)
	}
}

// --- RegisterCommands tests ---

func TestRegisterCommands_CallbackActionNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "cb_only",
		CallbackAction: dummyCallbackAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeCallbackQuery]["cb_only"]; !ok {
		t.Error("expected command registered for TypeCallbackQuery")
	}
}

func TestRegisterCommands_LocationActionNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "loc_only",
		LocationAction: dummyLocationAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeLocation]["loc_only"]; !ok {
		t.Error("expected command registered for TypeLocation")
	}
}

func TestRegisterCommands_ChosenInlineResultActionNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                     "cir_only",
		ChosenInlineResultAction: dummyChosenInlineResultAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeChosenInlineResult]["cir_only"]; !ok {
		t.Error("expected command registered for TypeChosenInlineResult")
	}
}

func TestRegisterCommands_PreCheckoutQueryActionNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                   "precheckout_only",
		PreCheckoutQueryAction: dummyPreCheckoutQueryAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypePreCheckoutQuery]["precheckout_only"]; !ok {
		t.Error("expected command registered for TypePreCheckoutQuery")
	}
}

func TestRegisterCommands_SuccessfulPaymentActionNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                    "payment_ok",
		SuccessfulPaymentAction: dummySuccessfulPaymentAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeSuccessfulPayment]["payment_ok"]; !ok {
		t.Error("expected command registered for TypeSuccessfulPayment")
	}
}

func TestRegisterCommands_RefundedPaymentActionNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                  "refund",
		RefundedPaymentAction: dummyRefundedPaymentAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeRefundedPayment]["refund"]; !ok {
		t.Error("expected command registered for TypeRefundedPayment")
	}
}

func TestRegisterCommands_TextAndCallbackNoInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "dual",
		TextAction:     dummyTextAction,
		CallbackAction: dummyCallbackAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeText]["dual"]; !ok {
		t.Error("expected command registered for TypeText")
	}
	if _, ok := cmds[botinput.TypeCallbackQuery]["dual"]; !ok {
		t.Error("expected command registered for TypeCallbackQuery")
	}
}

func TestRegisterCommands_InlineQueryWithInputType(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:              "iq",
		InputTypes:        []botinput.Type{botinput.TypeInlineQuery},
		InlineQueryAction: dummyInlineQueryAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeInlineQuery]["iq"]; !ok {
		t.Error("expected command registered for TypeInlineQuery")
	}
}

func TestRegisterCommands_ActionWithNoInputTypes_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for Action with no InputTypes")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:   "bad",
		Action: dummyAction,
	})
}

func TestRegisterCommands_StartActionNoTextAction(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:        "starter",
		StartAction: dummyTextAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeText]["starter"]; !ok {
		t.Error("expected command registered for TypeText when StartAction set and no TextAction")
	}
}

func TestRegisterCommands_TextActionWithInputTypeCallbackAlsoAddsText(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "both",
		InputTypes:     []botinput.Type{botinput.TypeCallbackQuery},
		CallbackAction: dummyCallbackAction,
		TextAction:     dummyTextAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeCallbackQuery]["both"]; !ok {
		t.Error("expected command registered for TypeCallbackQuery")
	}
	if _, ok := cmds[botinput.TypeText]["both"]; !ok {
		t.Error("expected TextAction auto-registered for TypeText")
	}
}

func TestRegisterCommands_CallbackActionWithInputTypeTextAlsoAddsCallback(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "both2",
		InputTypes:     []botinput.Type{botinput.TypeText},
		TextAction:     dummyTextAction,
		CallbackAction: dummyCallbackAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeText]["both2"]; !ok {
		t.Error("expected command registered for TypeText")
	}
	if _, ok := cmds[botinput.TypeCallbackQuery]["both2"]; !ok {
		t.Error("expected CallbackAction auto-registered for TypeCallbackQuery")
	}
}

func TestRegisterCommands_ChosenInlineResultWithInputTypeAutoAdds(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                     "cirtest",
		InputTypes:               []botinput.Type{botinput.TypeText},
		TextAction:               dummyTextAction,
		ChosenInlineResultAction: dummyChosenInlineResultAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeChosenInlineResult]["cirtest"]; !ok {
		t.Error("expected ChosenInlineResultAction auto-registered for TypeChosenInlineResult")
	}
}

func TestRegisterCommands_InlineQueryWithInputTypeAutoAdds(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:              "iqauto",
		InputTypes:        []botinput.Type{botinput.TypeText},
		TextAction:        dummyTextAction,
		InlineQueryAction: dummyInlineQueryAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeInlineQuery]["iqauto"]; !ok {
		t.Error("expected InlineQueryAction auto-registered for TypeInlineQuery")
	}
}

// --- addCommand duplicate panic test ---

func TestAddCommand_DuplicatePanics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{
		Code:       "dup",
		TextAction: dummyTextAction,
		InputTypes: []botinput.Type{botinput.TypeText},
	}
	router.RegisterCommands(cmd)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate command code")
		}
	}()
	router.RegisterCommands(cmd)
}

// --- RegisterCommands panics for missing actions with InputTypes ---

func TestRegisterCommands_TextInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypeText with no TextAction and no Action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypeText},
	})
}

func TestRegisterCommands_CallbackInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypeCallbackQuery with no CallbackAction and no Action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypeCallbackQuery},
	})
}

func TestRegisterCommands_InlineQueryInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypeInlineQuery with no InlineQueryAction and no Action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypeInlineQuery},
	})
}

func TestRegisterCommands_ChosenInlineResultInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypeChosenInlineResult with no action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypeChosenInlineResult},
	})
}

func TestRegisterCommands_PreCheckoutInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypePreCheckoutQuery with no action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypePreCheckoutQuery},
	})
}

func TestRegisterCommands_SuccessfulPaymentInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypeSuccessfulPayment with no action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypeSuccessfulPayment},
	})
}

func TestRegisterCommands_LocationInputTypeNoAction_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for TypeLocation with no action")
		}
	}()
	router.RegisterCommands(botsfw.Command{
		Code:       "noact",
		InputTypes: []botinput.Type{botinput.TypeLocation},
	})
}

func TestRegisterCommands_WithActionAndInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "generic",
		InputTypes: []botinput.Type{botinput.TypeText},
		Action:     dummyAction,
	})
	cmds := router.RegisteredCommands()
	if _, ok := cmds[botinput.TypeText]["generic"]; !ok {
		t.Error("expected command with Action registered for TypeText")
	}
}

// --- matchMessageCommands: Commands field matching ---

func TestMatchMessageCommands_ByCommandsField(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "alias", Commands: []string{"/alias", "/a"}, TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, true, "/a", "", commands)
	if matched == nil {
		t.Fatal("expected match by Commands field")
	} else if matched.Code != "alias" {
		t.Errorf("expected code 'alias', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_ByCommandsFieldWithArgs(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	commands := []botsfw.Command{
		{Code: "alias", Commands: []string{"/alias"}, TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "/alias extra", "", commands)
	if matched == nil {
		t.Fatal("expected match by Commands field prefix")
	} else if matched.Code != "alias" {
		t.Errorf("expected code 'alias', got %q", matched.Code)
	}
}

// --- matchMessageCommands: DefaultTitle with Icon ---

func TestMatchMessageCommands_ByDefaultTitle_WithIcon(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "")

	// Command with Title and Icon: DefaultTitle calls whc.CommandText(title, icon) → "Help 🆘"
	commands := []botsfw.Command{
		{Code: "helpicon", Title: "Help", Icon: "🆘", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "Help 🆘", "", commands)
	if matched == nil {
		t.Fatal("expected match by DefaultTitle with icon")
	} else if matched.Code != "helpicon" {
		t.Errorf("expected code 'helpicon', got %q", matched.Code)
	}
}

// --- TypeCommands addCommand empty code panics ---

// --- matchMessageCommands: awaiting reply with suffix path ---

func TestMatchMessageCommands_AwaitingReplyToSuffix(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "parent/child_cmd")

	commands := []botsfw.Command{
		{Code: "child_cmd", TextAction: dummyTextAction},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "some user input", "", commands)
	if matched == nil {
		t.Fatal("expected match via awaitingReplyTo suffix path")
	} else if matched.Code != "child_cmd" {
		t.Errorf("expected code 'child_cmd', got %q", matched.Code)
	}
}

func TestMatchMessageCommands_AwaitingReplyToWithReplies(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	mockWHC := setupMessageWHC(ctrl, "parent/sub")

	commands := []botsfw.Command{
		{
			Code:       "parent",
			TextAction: dummyTextAction,
			Replies: []botsfw.Command{
				{Code: "sub", TextAction: dummyTextAction},
			},
		},
	}
	matched := router.matchMessageCommands(mockWHC, nil, false, "any text", "", commands)
	if matched == nil {
		t.Fatal("expected match via command.Replies")
	} else if matched.Code != "sub" {
		t.Errorf("expected code 'sub', got %q", matched.Code)
	}
}

// --- changeLocaleIfLangPassed: SetLocale error path ---

func TestChangeLocaleIfLangPassed_SetLocaleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockWHC.EXPECT().ChatData().Return(mockChatData).AnyTimes()

	mockWHC.EXPECT().Locale().Return(i18n.Locale{Code5: "ru-RU"}).AnyTimes()
	mockWHC.EXPECT().SetLocale("de-DE").Return(fmt.Errorf("locale not found"))
	// SetPreferredLanguage should NOT be called

	u, _ := url.Parse("/settings?l=de-DE")
	m, err := changeLocaleIfLangPassed(mockWHC, u)
	if err != nil {
		t.Fatalf("error should be swallowed, got: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty text, got %q", m.Text)
	}
}

// --- changeLocaleIfLangPassed: 5-char lang code ---

func TestChangeLocaleIfLangPassed_FullCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
	mockWHC.EXPECT().Context().Return(context.Background()).AnyTimes()

	mockChatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	mockWHC.EXPECT().ChatData().Return(mockChatData).AnyTimes()

	mockWHC.EXPECT().Locale().Return(i18n.Locale{Code5: "en-US"}).AnyTimes()
	mockWHC.EXPECT().SetLocale("ru-RU").Return(nil)
	mockChatData.EXPECT().SetPreferredLanguage("ru-RU")

	u, _ := url.Parse("/settings?l=ru-RU")
	m, err := changeLocaleIfLangPassed(mockWHC, u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty text, got %q", m.Text)
	}
}

func TestTypeCommands_AddCommand_EmptyCode_Panics(t *testing.T) {
	tc := newTypeCommands(1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty code")
		}
	}()
	tc.addCommand(botsfw.Command{}, botinput.TypeText)
}
