package botswebhook

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

// --- Combined input types (embed MockInputMessage + specific interface) ---

// callbackInput satisfies both botinput.InputMessage and botinput.CallbackQuery
type callbackInput struct {
	*mock_botinput.MockInputMessage
	data string
}

func (c *callbackInput) GetData() string              { return c.data }
func (c *callbackInput) GetID() string                { return "cb1" }
func (c *callbackInput) GetFrom() botinput.Sender     { return nil }
func (c *callbackInput) GetMessage() botinput.Message { return nil }
func (c *callbackInput) Chat() botinput.Chat          { return nil }

// inlineQueryInput satisfies both botinput.InputMessage and botinput.InlineQuery
type inlineQueryInput struct {
	*mock_botinput.MockInputMessage
	query string
}

func (i *inlineQueryInput) GetID() any               { return "iq1" }
func (i *inlineQueryInput) GetInlineQueryID() string { return "iq1" }
func (i *inlineQueryInput) GetFrom() botinput.Sender { return nil }
func (i *inlineQueryInput) GetQuery() string         { return i.query }
func (i *inlineQueryInput) GetOffset() string        { return "" }

// chosenInlineResultInput satisfies botinput.InputMessage and botinput.ChosenInlineResult
type chosenInlineResultInput struct {
	*mock_botinput.MockInputMessage
	query string
}

func (c *chosenInlineResultInput) GetResultID() string        { return "res1" }
func (c *chosenInlineResultInput) GetInlineMessageID() string { return "im1" }
func (c *chosenInlineResultInput) GetFrom() botinput.Sender   { return nil }
func (c *chosenInlineResultInput) GetQuery() string           { return c.query }

// preCheckoutQueryInput satisfies botinput.InputMessage and botinput.PreCheckoutQuery
type preCheckoutQueryInput struct {
	*mock_botinput.MockInputMessage
	payload string
}

func (p *preCheckoutQueryInput) GetPreCheckoutQueryID() string    { return "pcq1" }
func (p *preCheckoutQueryInput) GetCurrency() string              { return "USD" }
func (p *preCheckoutQueryInput) GetTotalAmount() int              { return 1000 }
func (p *preCheckoutQueryInput) GetInvoicePayload() string        { return p.payload }
func (p *preCheckoutQueryInput) GetFrom() botinput.Sender         { return nil }
func (p *preCheckoutQueryInput) GetShippingOptionID() string      { return "" }
func (p *preCheckoutQueryInput) GetOrderInfo() botinput.OrderInfo { return nil }

// successfulPaymentInput satisfies botinput.InputMessage and botinput.SuccessfulPayment
type successfulPaymentInput struct {
	*mock_botinput.MockInputMessage
	payload string
}

func (s *successfulPaymentInput) GetCurrency() string                      { return "USD" }
func (s *successfulPaymentInput) GetTotalAmount() int                      { return 1000 }
func (s *successfulPaymentInput) GetInvoicePayload() string                { return s.payload }
func (s *successfulPaymentInput) GetMessengerChargeID() string             { return "mc1" }
func (s *successfulPaymentInput) GetPaymentProviderChargeID() string       { return "ppc1" }
func (s *successfulPaymentInput) GetSubscriptionExpirationDate() time.Time { return time.Time{} }
func (s *successfulPaymentInput) GetIsRecurring() bool                     { return false }
func (s *successfulPaymentInput) GetIsFirstRecurring() bool                { return false }
func (s *successfulPaymentInput) GetShippingOptionID() string              { return "" }
func (s *successfulPaymentInput) GetOrderInfo() botinput.OrderInfo         { return nil }

// contactInput satisfies botinput.InputMessage and botinput.ContactMessage
type contactInput struct {
	*mock_botinput.MockInputMessage
}

func (c *contactInput) GetPhoneNumber() string { return "+1234567890" }
func (c *contactInput) GetFirstName() string   { return "John" }
func (c *contactInput) GetLastName() string    { return "Doe" }
func (c *contactInput) GetBotUserID() string   { return "u123" }
func (c *contactInput) GetVCard() string       { return "" }

// referralInput satisfies botinput.InputMessage and botinput.ReferralMessage
type referralInput struct {
	*mock_botinput.MockInputMessage
}

func (r *referralInput) Type() string    { return "ref" }
func (r *referralInput) Source() string  { return "web" }
func (r *referralInput) RefData() string { return "data123" }

// --- Helpers ---

func newMockIM(ctrl *gomock.Controller, inputType botinput.Type) *mock_botinput.MockInputMessage {
	im := mock_botinput.NewMockInputMessage(ctrl)
	im.EXPECT().InputType().Return(inputType).AnyTimes()
	im.EXPECT().LogRequest().AnyTimes()
	im.EXPECT().Chat().Return(nil).AnyTimes()
	im.EXPECT().GetSender().Return(nil).AnyTimes()
	im.EXPECT().GetRecipient().Return(nil).AnyTimes()
	im.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	im.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	im.EXPECT().MessageIntID().Return(0).AnyTimes()
	im.EXPECT().MessageStringID().Return("").AnyTimes()
	return im
}

func setupDispatchWHC(ctrl *gomock.Controller, input botinput.InputMessage) (
	*mock_botsfw.MockWebhookContext,
	*mock_botsfw.MockWebhookResponder,
	*mock_botsfw.MockWebhookAnalytics,
	*mock_botsfwmodels.MockBotChatData,
) {
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(input).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	platform := mock_botsfw.NewMockBotPlatform(ctrl)
	platform.EXPECT().ID().Return("test").AnyTimes()
	whc.EXPECT().BotPlatform().Return(platform).AnyTimes()

	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().Locale().Return(i18n.Locale{Code5: "en-US"}).AnyTimes()
	whc.EXPECT().SetLocale(gomock.Any()).Return(nil).AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().TranslateNoWarning(gomock.Any(), gomock.Any()).Return("").AnyTimes()
	whc.EXPECT().CommandText(gomock.Any(), gomock.Any()).DoAndReturn(
		func(t, i string) string { return t },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	whc.EXPECT().NewMessageByCode(gomock.Any(), gomock.Any()).DoAndReturn(
		func(code string, _ ...interface{}) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: code}}
		},
	).AnyTimes()
	whc.EXPECT().SaveBotChat().Return(nil).AnyTimes()

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAwaitingReplyTo().Return("").AnyTimes()
	chatData.EXPECT().IsChanged().Return(false).AnyTimes()
	chatData.EXPECT().HasChangedVars().Return(false).AnyTimes()
	chatData.EXPECT().SetDtLastInteraction(gomock.Any()).AnyTimes()
	chatData.EXPECT().SetUpdatedTime(gomock.Any()).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()

	return whc, responder, analytics, chatData
}

// ======================== matchByQueryOrMatcher ========================

func TestMatchByQueryOrMatcher_QueryMatchesAndHasAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)

	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: "help"}

	commands := map[botsfw.CommandCode]botsfw.Command{
		"help": {Code: "help", InlineQueryAction: dummyInlineQueryAction},
	}
	matched, qURL := matchByQueryOrMatcher(whc, iqInput, commands, func(c botsfw.Command) bool {
		return c.InlineQueryAction != nil
	})
	if matched == nil {
		t.Fatal("expected match by query URL path")
	} else if matched.Code != "help" {
		t.Errorf("expected code 'help', got %q", matched.Code)
	}
	if qURL == nil {
		t.Fatal("expected queryURL to be non-nil")
	}
}

func TestMatchByQueryOrMatcher_QueryMatchesButNoAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)

	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: "help"}

	commands := map[botsfw.CommandCode]botsfw.Command{
		"help": {Code: "help"}, // no InlineQueryAction
	}
	matched, _ := matchByQueryOrMatcher(whc, iqInput, commands, func(c botsfw.Command) bool {
		return c.InlineQueryAction != nil
	})
	if matched != nil {
		t.Errorf("expected nil when hasAction returns false, got %q", matched.Code)
	}
}

func TestMatchByQueryOrMatcher_EmptyQuery_MatcherMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)

	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: ""}

	commands := map[botsfw.CommandCode]botsfw.Command{
		"custom": {
			Code: "custom",
			Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
				return true
			},
			InlineQueryAction: dummyInlineQueryAction,
		},
	}
	matched, _ := matchByQueryOrMatcher(whc, iqInput, commands, func(c botsfw.Command) bool {
		return c.InlineQueryAction != nil
	})
	if matched == nil {
		t.Fatal("expected match via Matcher")
	} else if matched.Code != "custom" {
		t.Errorf("expected code 'custom', got %q", matched.Code)
	}
}

func TestMatchByQueryOrMatcher_NoMatchAtAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)

	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: "unknown"}

	commands := map[botsfw.CommandCode]botsfw.Command{
		"help": {Code: "help", InlineQueryAction: dummyInlineQueryAction},
	}
	matched, _ := matchByQueryOrMatcher(whc, iqInput, commands, func(c botsfw.Command) bool {
		return c.InlineQueryAction != nil
	})
	if matched != nil {
		t.Errorf("expected nil, got %q", matched.Code)
	}
}

func TestMatchByQueryOrMatcher_InvalidURL_FallsToMatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)

	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	// A query with valid URL syntax but no matching path, and a matcher that matches
	iqInput := &inlineQueryInput{MockInputMessage: im, query: "/no-match"}

	commands := map[botsfw.CommandCode]botsfw.Command{
		"custom": {
			Code: "custom",
			Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
				return true
			},
			InlineQueryAction: dummyInlineQueryAction,
		},
	}
	matched, _ := matchByQueryOrMatcher(whc, iqInput, commands, func(c botsfw.Command) bool {
		return c.InlineQueryAction != nil
	})
	if matched == nil {
		t.Fatal("expected match via Matcher after URL path miss")
	} else if matched.Code != "custom" {
		t.Errorf("expected code 'custom', got %q", matched.Code)
	}
}

// ======================== DispatchInlineQuery ========================

func TestDispatchInlineQuery_Panics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from DispatchInlineQuery")
		}
	}()
	router.DispatchInlineQuery(nil)
}

// ======================== Dispatch ========================

func TestDispatch_CallbackQuery_MatchedByPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)
	cbInput := &callbackInput{MockInputMessage: im, data: "settings"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cbInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "settings",
		CallbackAction: dummyCallbackAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_CallbackQuery_MatchedByMatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)
	cbInput := &callbackInput{MockInputMessage: im, data: "/nomatch"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cbInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code: "custom",
		Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
			return true
		},
		CallbackAction: dummyCallbackAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_InlineQuery_MatchedByQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: "search"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, iqInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:              "search",
		InputTypes:        []botinput.Type{botinput.TypeInlineQuery},
		InlineQueryAction: dummyInlineQueryAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_InlineQuery_FallbackToSingleCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: ""}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, iqInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:              "default_iq",
		InputTypes:        []botinput.Type{botinput.TypeInlineQuery},
		InlineQueryAction: dummyInlineQueryAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_InlineQuery_WithAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: ""}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, iqInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "fallback",
		InputTypes: []botinput.Type{botinput.TypeInlineQuery},
		Action:     dummyAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_ChosenInlineResult_MatchedByQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeChosenInlineResult)
	cirInput := &chosenInlineResultInput{MockInputMessage: im, query: "pick"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cirInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                     "pick",
		ChosenInlineResultAction: dummyChosenInlineResultAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_ChosenInlineResult_NoMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeChosenInlineResult)
	cirInput := &chosenInlineResultInput{MockInputMessage: im, query: "unknown"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cirInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	// Two commands so len(typeCommands.all) != 1, preventing fallback to single command
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommandsForInputType(botinput.TypeChosenInlineResult,
		botsfw.Command{Code: "cmd1", ChosenInlineResultAction: dummyChosenInlineResultAction, InputTypes: []botinput.Type{botinput.TypeChosenInlineResult}},
		botsfw.Command{Code: "cmd2", ChosenInlineResultAction: dummyChosenInlineResultAction, InputTypes: []botinput.Type{botinput.TypeChosenInlineResult}},
	)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_ChosenInlineResult_WithAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeChosenInlineResult)
	cirInput := &chosenInlineResultInput{MockInputMessage: im, query: ""}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cirInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                     "single",
		ChosenInlineResultAction: dummyChosenInlineResultAction,
		InputTypes:               []botinput.Type{botinput.TypeChosenInlineResult},
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_TextMessage_ByCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("/help").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, textInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "help",
		TextAction: dummyTextAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_TextMessage_StartAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("/start ref123").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, textInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(
		botsfw.Command{Code: "start", TextAction: dummyTextAction},
		botsfw.Command{Code: "ref123", StartAction: dummyTextAction},
	)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_TextMessage_MatcherMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("random text").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, textInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code: "catch_all",
		Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
			return true
		},
		TextAction: dummyTextAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_TextMessage_WithGenericAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("/gen").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, textInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "gen",
		InputTypes: []botinput.Type{botinput.TypeText},
		Action:     dummyAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_LocationMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	locInput := mock_botinput.NewMockLocationMessage(ctrl)
	locInput.EXPECT().InputType().Return(botinput.TypeLocation).AnyTimes()
	locInput.EXPECT().LogRequest().AnyTimes()
	locInput.EXPECT().Chat().Return(nil).AnyTimes()
	locInput.EXPECT().GetSender().Return(nil).AnyTimes()
	locInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	locInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	locInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	locInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	locInput.EXPECT().MessageStringID().Return("").AnyTimes()
	locInput.EXPECT().GetLatitude().Return(51.5074).AnyTimes()
	locInput.EXPECT().GetLongitude().Return(-0.1278).AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(locInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	platform := mock_botsfw.NewMockBotPlatform(ctrl)
	platform.EXPECT().ID().Return("test").AnyTimes()
	whc.EXPECT().BotPlatform().Return(platform).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().SaveBotChat().Return(nil).AnyTimes()

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAwaitingReplyTo().Return("loc_cmd").AnyTimes()
	chatData.EXPECT().IsChanged().Return(false).AnyTimes()
	chatData.EXPECT().HasChangedVars().Return(false).AnyTimes()
	chatData.EXPECT().SetDtLastInteraction(gomock.Any()).AnyTimes()
	chatData.EXPECT().SetUpdatedTime(gomock.Any()).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "loc_cmd",
		LocationAction: dummyLocationAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_PreCheckoutQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypePreCheckoutQuery)
	pcqInput := &preCheckoutQueryInput{MockInputMessage: im, payload: "pay_cmd"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, pcqInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                   "pay_cmd",
		PreCheckoutQueryAction: dummyPreCheckoutQueryAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_SuccessfulPayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeSuccessfulPayment)
	spInput := &successfulPaymentInput{MockInputMessage: im, payload: "pay_cmd"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, spInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                    "pay_cmd",
		SuccessfulPaymentAction: dummySuccessfulPaymentAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_NoCommandsForType(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeContact)
	ciInput := &contactInput{MockInputMessage: im}

	whc, responder, _, _ := setupDispatchWHC(ctrl, ciInput)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	// No commands registered for TypeContact

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_UnmatchedText_HandleUnmatchedCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("random unmatched text").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, _, _ := setupDispatchWHC(ctrl, textInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "help",
		TextAction: dummyTextAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "unmatched response"}}).Times(1)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_UnmatchedText_NoHandleUnmatchedResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("random unmatched text").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, _, _ := setupDispatchWHC(ctrl, textInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "help",
		TextAction: dummyTextAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).Times(1)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_ChosenInlineResult_WithGenericAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeChosenInlineResult)
	cirInput := &chosenInlineResultInput{MockInputMessage: im, query: ""}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cirInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	// Single command with Action instead of ChosenInlineResultAction
	router.RegisterCommandsForInputType(botinput.TypeChosenInlineResult, botsfw.Command{
		Code:                     "fallback_cir",
		Action:                   dummyAction,
		ChosenInlineResultAction: nil,
		InputTypes:               []botinput.Type{botinput.TypeChosenInlineResult},
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_PreCheckoutQuery_WithGenericAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypePreCheckoutQuery)
	pcqInput := &preCheckoutQueryInput{MockInputMessage: im, payload: "pay_generic"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, pcqInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommandsForInputType(botinput.TypePreCheckoutQuery, botsfw.Command{
		Code:       "pay_generic",
		Action:     dummyAction,
		InputTypes: []botinput.Type{botinput.TypePreCheckoutQuery},
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_SuccessfulPayment_WithGenericAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeSuccessfulPayment)
	spInput := &successfulPaymentInput{MockInputMessage: im, payload: "pay_generic"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, spInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommandsForInputType(botinput.TypeSuccessfulPayment, botsfw.Command{
		Code:       "pay_generic",
		Action:     dummyAction,
		InputTypes: []botinput.Type{botinput.TypeSuccessfulPayment},
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ======================== logInputDetails ========================

func TestLogInputDetails_TypeText_NotEdited(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().Text().Return("hello").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(textInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, true)
}

func TestLogInputDetails_TypeText_Edited(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().Text().Return("edited text").AnyTimes()
	textInput.EXPECT().IsEdited().Return(true).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(textInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, false)
}

func TestLogInputDetails_TypeContact(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeContact)
	ci := &contactInput{MockInputMessage: im}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(ci).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, true)
}

func TestLogInputDetails_TypeInlineQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeInlineQuery)
	iqInput := &inlineQueryInput{MockInputMessage: im, query: "search term"}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(iqInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, true)
}

func TestLogInputDetails_TypeCallbackQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)
	cbInput := &callbackInput{MockInputMessage: im, data: "some/data"}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(cbInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, true)
}

func TestLogInputDetails_TypeChosenInlineResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeChosenInlineResult)
	cirInput := &chosenInlineResultInput{MockInputMessage: im, query: "q"}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(cirInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, true)
}

func TestLogInputDetails_TypeReferral(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeReferral)
	refInput := &referralInput{MockInputMessage: im}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(refInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, false)
}

func TestLogInputDetails_DefaultCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeLocation)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	logInputDetails(whc, true)
}

func TestLogInputDetails_SendMessageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeLocation)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, fmt.Errorf("send failed")).AnyTimes()

	logInputDetails(whc, false)
}

// ======================== processCommandResponse ========================

func TestProcessCommandResponse_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", TextAction: dummyTextAction}

	router.processCommandResponse(&cmd, responder, whc, botmsg.MessageFromBot{}, fmt.Errorf("test error"))
}

func TestProcessCommandResponse_SendsSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}}
	router.processCommandResponse(&cmd, responder, whc, m, nil)
}

func TestProcessCommandResponse_MessageNotModified(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		botsfw.OnMessageSentResponse{}, fmt.Errorf("Bad Request: message is not modified"),
	).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}}
	router.processCommandResponse(&cmd, responder, whc, m, nil)
}

func TestProcessCommandResponse_MessageToEditNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		botsfw.OnMessageSentResponse{}, fmt.Errorf("Bad Request: message to edit not found"),
	).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}}
	router.processCommandResponse(&cmd, responder, whc, m, nil)
}

func TestProcessCommandResponse_OtherSendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		botsfw.OnMessageSentResponse{}, fmt.Errorf("network error"),
	).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}}
	router.processCommandResponse(&cmd, responder, whc, m, nil)
}

func TestProcessCommandResponse_NilCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc, responder, _, _ := setupDispatchWHC(ctrl, im)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}}
	router.processCommandResponse(nil, responder, whc, m, nil)
}

func TestProcessCommandResponse_WithResponseChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), botsfw.BotAPISendMessageOverHTTPS).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}, ResponseChannel: botsfw.BotAPISendMessageOverHTTPS}
	router.processCommandResponse(&cmd, responder, whc, m, nil)
}

// ======================== processCommandResponseError ========================

func TestProcessCommandResponseError_ProductionEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: botsfw.EnvProduction}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	analytics.EXPECT().Enqueue(gomock.Any()).Times(1)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("some error"))
}

func TestProcessCommandResponseError_TextInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("some error"))
}

func TestProcessCommandResponseError_TextInput_WithErrorFooter(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(func(_ context.Context, _ ErrorFooterArgs) string {
		return "Contact support"
	}).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("some error"))
}

func TestProcessCommandResponseError_TextInput_SendFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, fmt.Errorf("send failed")).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("some error"))
}

func TestProcessCommandResponseError_CallbackQueryInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)
	cbInput := &callbackInput{MockInputMessage: im, data: "some/data"}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(cbInput).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("callback error"))
}

func TestProcessCommandResponseError_CallbackQueryInput_SendFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)
	cbInput := &callbackInput{MockInputMessage: im, data: "some/data"}

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(cbInput).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, fmt.Errorf("send failed")).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("callback error"))
}

func TestProcessCommandResponseError_DefaultInputType(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeLocation)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("some error"))
}

func TestProcessCommandResponseError_ContactInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeContact)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test"}

	router.processCommandResponseError(whc, &cmd, responder, fmt.Errorf("some error"))
}

// ======================== Dispatch with chat data changes ========================

func TestDispatch_TextMessage_ChatDataChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	textInput := mock_botinput.NewMockTextMessage(ctrl)
	textInput.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	textInput.EXPECT().LogRequest().AnyTimes()
	textInput.EXPECT().Chat().Return(nil).AnyTimes()
	textInput.EXPECT().GetSender().Return(nil).AnyTimes()
	textInput.EXPECT().GetRecipient().Return(nil).AnyTimes()
	textInput.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	textInput.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	textInput.EXPECT().MessageIntID().Return(0).AnyTimes()
	textInput.EXPECT().MessageStringID().Return("").AnyTimes()
	textInput.EXPECT().Text().Return("/help").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(textInput).AnyTimes()

	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	whc.EXPECT().Responder().Return(responder).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	whc.EXPECT().Analytics().Return(analytics).AnyTimes()

	platform := mock_botsfw.NewMockBotPlatform(ctrl)
	platform.EXPECT().ID().Return("test").AnyTimes()
	whc.EXPECT().BotPlatform().Return(platform).AnyTimes()
	whc.EXPECT().GetBotCode().Return("testbot").AnyTimes()
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{Env: "local"}).AnyTimes()
	whc.EXPECT().Locale().Return(i18n.Locale{Code5: "en-US"}).AnyTimes()
	whc.EXPECT().Translate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key string, _ ...interface{}) string { return key },
	).AnyTimes()
	whc.EXPECT().TranslateNoWarning(gomock.Any(), gomock.Any()).Return("").AnyTimes()
	whc.EXPECT().CommandText(gomock.Any(), gomock.Any()).DoAndReturn(
		func(t, i string) string { return t },
	).AnyTimes()
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(
		func(text string) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
		},
	).AnyTimes()
	whc.EXPECT().NewMessageByCode(gomock.Any(), gomock.Any()).DoAndReturn(
		func(code string, _ ...interface{}) botmsg.MessageFromBot {
			return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: code}}
		},
	).AnyTimes()

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAwaitingReplyTo().Return("").AnyTimes()
	chatData.EXPECT().IsChanged().Return(true).AnyTimes()
	chatData.EXPECT().HasChangedVars().Return(false).AnyTimes()
	chatData.EXPECT().SetDtLastInteraction(gomock.Any()).AnyTimes()
	chatData.EXPECT().SetUpdatedTime(gomock.Any()).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	whc.EXPECT().SaveBotChat().Return(nil).AnyTimes()

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "help",
		TextAction: dummyTextAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch_CallbackQuery_EmptyData(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeCallbackQuery)
	// Use non-empty data that matches by Matcher, not by URL path
	cbInput := &callbackInput{MockInputMessage: im, data: "some-unmatched-path"}

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, cbInput)

	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code: "custom",
		Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
			return true
		},
		CallbackAction: dummyCallbackAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
