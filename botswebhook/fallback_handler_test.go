package botswebhook

import (
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"go.uber.org/mock/gomock"
)

// newMockTextInput creates a MockTextMessage that looks like the given text.
func newMockTextInput(ctrl *gomock.Controller, text string) *mock_botinput.MockTextMessage {
	m := mock_botinput.NewMockTextMessage(ctrl)
	m.EXPECT().InputType().Return(botinput.TypeText).AnyTimes()
	m.EXPECT().LogRequest().AnyTimes()
	m.EXPECT().Chat().Return(nil).AnyTimes()
	m.EXPECT().GetSender().Return(nil).AnyTimes()
	m.EXPECT().GetRecipient().Return(nil).AnyTimes()
	m.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	m.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	m.EXPECT().MessageIntID().Return(0).AnyTimes()
	m.EXPECT().MessageStringID().Return("").AnyTimes()
	m.EXPECT().Text().Return(text).AnyTimes()
	m.EXPECT().IsEdited().Return(false).AnyTimes()
	return m
}

// TestSetFallbackHandler_storesAndReplacesAction verifies that SetFallbackHandler
// stores the action and that a second call replaces the first.
func TestSetFallbackHandler_storesAndReplacesAction(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)

	called := 0
	first := func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		called = 1
		return botmsg.MessageFromBot{}, nil
	}
	second := func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		called = 2
		return botmsg.MessageFromBot{}, nil
	}

	router.SetFallbackHandler(botinput.TypeText, first)
	if _, ok := router.fallbackHandlers[botinput.TypeText]; !ok {
		t.Fatal("fallback not stored after first SetFallbackHandler")
	}

	router.SetFallbackHandler(botinput.TypeText, second)
	// Invoke the stored handler to confirm replacement.
	//nolint:errcheck
	router.fallbackHandlers[botinput.TypeText](nil)
	if called != 2 {
		t.Fatalf("expected called=2 (second handler), got %d", called)
	}
}

// TestSetFallbackHandler_firesWhenNoCommandMatches verifies that the fallback fires
// when a text message arrives and no command matches.
func TestSetFallbackHandler_firesWhenNoCommandMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router := NewWebhookRouter(nil)

	fallbackFired := false
	router.SetFallbackHandler(botinput.TypeText, func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		fallbackFired = true
		m := botmsg.MessageFromBot{}
		m.Text = "fallback"
		return m, nil
	})

	textInput := newMockTextInput(ctrl, "this text matches nothing")
	whc, responder, analytics, _ := setupDispatchWHC(ctrl, textInput)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	_ = handler

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("Dispatch returned unexpected error: %v", err)
	}
	if !fallbackFired {
		t.Fatal("expected fallback handler to fire, but it did not")
	}
}

// TestSetFallbackHandler_notFiredWhenCommandMatches checks that when a command
// matches the input the fallback is NOT invoked.
func TestSetFallbackHandler_notFiredWhenCommandMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router := NewWebhookRouter(nil)

	fallbackFired := false
	router.SetFallbackHandler(botinput.TypeText, func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		fallbackFired = true
		return botmsg.MessageFromBot{}, nil
	})

	commandFired := false
	router.RegisterCommands(botsfw.Command{
		Code:     "hello",
		Commands: []string{"/hello"},
		TextAction: func(whc botsfw.WebhookContext, text string) (botmsg.MessageFromBot, error) {
			commandFired = true
			m := botmsg.MessageFromBot{}
			m.Text = "hi!"
			return m, nil
		},
	})

	textInput := newMockTextInput(ctrl, "/hello")
	whc, responder, analytics, _ := setupDispatchWHC(ctrl, textInput)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("Dispatch returned unexpected error: %v", err)
	}
	if !commandFired {
		t.Fatal("expected registered command to fire")
	}
	if fallbackFired {
		t.Fatal("fallback should NOT fire when a command matched")
	}
}

// TestSetFallbackHandler_differentInputTypes verifies that fallbacks are scoped
// to their input type and do not interfere with each other.
func TestSetFallbackHandler_differentInputTypes(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)

	router.SetFallbackHandler(botinput.TypeText, func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		return botmsg.MessageFromBot{}, nil
	})
	router.SetFallbackHandler(botinput.TypeCallbackQuery, func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		return botmsg.MessageFromBot{}, nil
	})

	if len(router.fallbackHandlers) != 2 {
		t.Fatalf("expected 2 fallback handlers, got %d", len(router.fallbackHandlers))
	}
	if _, ok := router.fallbackHandlers[botinput.TypeText]; !ok {
		t.Fatal("TypeText fallback missing")
	}
	if _, ok := router.fallbackHandlers[botinput.TypeCallbackQuery]; !ok {
		t.Fatal("TypeCallbackQuery fallback missing")
	}
}
