package botswebhook

// coverage_extra_test.go: tests added to raise botswebhook coverage beyond 71.7%.
//
// Targets:
//   - SetFallbackHandler: nil-map re-init branch (zero-value webhooksRouter{})
//   - RegisterCommandsForInputType: duplicate TypeInlineQuery panic branch
//   - RegisterCommands: Action-with-no-InputTypes panic branch
//   - Dispatch: botinput.Message generic case, default case, no-commandAction error,
//     fallback when matchedCommand==nil for no-choice-inline-result
//   - processCommandResponse: SendGate gate-refused branch
//   - webhookDriver.logInput: all switch cases
//   - webhookDriver.reportPanicToAnalytics

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"github.com/strongo/analytics"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

// ─── gated responder ──────────────────────────────────────────────────────────

// gatedTestResponder implements both botsfw.WebhookResponder and botsfw.SendGate.
// When refuse is non-nil, CanSend returns that error so processCommandResponse's
// "gateErr != nil" branch is reached.
type gatedTestResponder struct {
	refuse error
}

func (r *gatedTestResponder) SendMessage(_ context.Context, _ botmsg.MessageFromBot, _ botmsg.BotAPISendMessageChannel) (botsfw.OnMessageSentResponse, error) {
	return botsfw.OnMessageSentResponse{}, nil
}
func (r *gatedTestResponder) DeleteMessage(_ context.Context, _ string) error { return nil }
func (r *gatedTestResponder) CanSend(_ context.Context, _ botmsg.MessageFromBot) error {
	return r.refuse
}

// ─── SetFallbackHandler: nil-map re-init branch ───────────────────────────────

// TestSetFallbackHandler_nilMapReinit exercises the `if fallbackHandlers == nil`
// branch by using a zero-value webhooksRouter (all fields are nil).
func TestSetFallbackHandler_nilMapReinit(t *testing.T) {
	router := &webhooksRouter{} // zero value — fallbackHandlers is nil

	action := func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		return botmsg.MessageFromBot{}, nil
	}
	// Must not panic; must initialise fallbackHandlers on the fly.
	router.SetFallbackHandler(botinput.TypeText, action)

	if router.fallbackHandlers == nil {
		t.Fatal("SetFallbackHandler did not initialise fallbackHandlers map")
	}
	if _, ok := router.fallbackHandlers[botinput.TypeText]; !ok {
		t.Fatal("TypeText fallback not stored")
	}
}

// ─── RegisterCommandsForInputType: duplicate TypeInlineQuery panic ────────────

func TestRegisterCommandsForInputType_DuplicateInlineQueryPanics(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{
		Code:              "iq",
		InputTypes:        []botinput.Type{botinput.TypeInlineQuery},
		InlineQueryAction: dummyInlineQueryAction,
	}
	router.RegisterCommandsForInputType(botinput.TypeInlineQuery, cmd)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate TypeInlineQuery registration")
		}
	}()
	// Second call for the same type must panic.
	router.RegisterCommandsForInputType(botinput.TypeInlineQuery, botsfw.Command{
		Code:              "iq2",
		InputTypes:        []botinput.Type{botinput.TypeInlineQuery},
		InlineQueryAction: dummyInlineQueryAction,
	})
}

// ─── RegisterCommands: Action-with-no-InputTypes panic ───────────────────────

func TestRegisterCommands_ActionWithNoInputTypesPanics(t *testing.T) {
	router := NewWebhookRouter(nil)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for command with Action but no InputTypes")
		}
	}()
	// Action set but InputTypes is empty — must panic.
	router.RegisterCommands(botsfw.Command{
		Code:   "bad",
		Action: dummyAction,
	})
}

// ─── processCommandResponse: m.Analytics non-nil path ────────────────────────

// TestProcessCommandResponse_WithAnalyticsMessage covers the `m.Analytics != nil`
// branch in processCommandResponse's analytics tracking block.
func TestProcessCommandResponse_WithAnalyticsMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	im := newMockIM(ctrl, botinput.TypeText)
	whc, responder, whcAnalytics, _ := setupDispatchWHC(ctrl, im)
	whcAnalytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).Times(1)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	// Set m.Analytics to a real analytics.Message to hit the non-nil branch.
	analyticsMsg := analytics.NewErrorMessage(fmt.Errorf("test event"))
	m := botmsg.MessageFromBot{
		TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"},
		Analytics:          analyticsMsg,
	}
	router.processCommandResponse(&cmd, responder, whc, m, nil)
}

// ─── logInputDetails: edited text with send error ────────────────────────────

// TestLogInputDetails_EditedText_SendError covers the branch at line ~799 of router.go
// where IsEdited()==true and the "editing not supported" message fails to send.
func TestLogInputDetails_EditedText_SendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	// Simulate a send failure for the "editing not supported" notification.
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(botsfw.OnMessageSentResponse{}, errors.New("send failed")).AnyTimes()

	logInputDetails(whc, false) // Must not panic; error is just logged.
}

// ─── RegisterCommands: LocationAction with InputTypes including TypeLocation ──

// TestRegisterCommands_LocationActionWithInputType covers the
// `if command.LocationAction != nil && locationAdded` branch (line ~231 router.go)
// by registering a command that has both LocationAction and TypeLocation in InputTypes.
func TestRegisterCommands_LocationActionWithInputType(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "loc",
		InputTypes:     []botinput.Type{botinput.TypeLocation},
		LocationAction: dummyLocationAction,
	})
	cmds := router.RegisteredCommands()
	// The command should be in TypeLocation commands.
	if len(cmds[botinput.TypeLocation]) == 0 {
		t.Fatal("expected command registered for TypeLocation")
	}
}

// ─── processCommandResponse: SendGate gate-refused branch ────────────────────

func TestProcessCommandResponse_GateRefused(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	im := newMockIM(ctrl, botinput.TypeText)
	whc, _, analytics, _ := setupDispatchWHC(ctrl, im)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	// Use a gated responder that refuses the send.
	gated := &gatedTestResponder{refuse: fmt.Errorf("outside window: %w", botsfw.ErrSendNotPermitted)}

	router := NewWebhookRouter(nil).(*webhooksRouter)
	cmd := botsfw.Command{Code: "test", Title: "Test"}

	m := botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}}
	// Should NOT call SendMessage (gate refused); must not panic.
	router.processCommandResponse(&cmd, gated, whc, m, nil)
}

// ─── Dispatch: botinput.Message generic case (single command, action) ─────────

// messageOnlyInput implements botinput.InputMessage but NOT TextMessage,
// LocationMessage, CallbackQuery, etc. It matches the generic `case botinput.Message:`
// branch in Dispatch.
type messageOnlyInput struct {
	*mock_botinput.MockMessage
}

func newMessageOnlyInput(ctrl *gomock.Controller, inputType botinput.Type) *messageOnlyInput {
	m := mock_botinput.NewMockMessage(ctrl)
	m.EXPECT().InputType().Return(inputType).AnyTimes()
	m.EXPECT().LogRequest().AnyTimes()
	m.EXPECT().Chat().Return(nil).AnyTimes()
	m.EXPECT().GetSender().Return(nil).AnyTimes()
	m.EXPECT().GetRecipient().Return(nil).AnyTimes()
	m.EXPECT().GetTime().Return(time.Now()).AnyTimes()
	m.EXPECT().BotChatID().Return("chat1", nil).AnyTimes()
	m.EXPECT().MessageIntID().Return(0).AnyTimes()
	m.EXPECT().MessageStringID().Return("").AnyTimes()
	return &messageOnlyInput{MockMessage: m}
}

// TestDispatch_GenericMessage_SingleCommand tests the `case botinput.Message:` branch
// when there is exactly one command registered (matchedCommand = &typeCommands.all[0]).
func TestDispatch_GenericMessage_SingleCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// TypeNewChatMembers does not match TextMessage/CallbackQuery/etc, so it falls
	// through to the generic `case botinput.Message:` branch.
	input := newMessageOnlyInput(ctrl, botinput.TypeNewChatMembers)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, input)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommandsForInputType(botinput.TypeNewChatMembers, botsfw.Command{
		Code:   "ncm",
		Action: dummyAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDispatch_GenericMessage_MatcherMatches tests the `case botinput.Message:` branch
// with multiple commands where a Matcher succeeds.
func TestDispatch_GenericMessage_MatcherMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	input := newMessageOnlyInput(ctrl, botinput.TypeNewChatMembers)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, input)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommandsForInputType(botinput.TypeNewChatMembers,
		botsfw.Command{Code: "ncm1", Action: dummyAction},
		botsfw.Command{
			Code:   "ncm2",
			Action: dummyAction,
			Matcher: func(_ botsfw.Command, _ botsfw.WebhookContext) bool {
				return true
			},
		},
	)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDispatch_GenericMessage_NoMatch tests the `case botinput.Message:` branch when
// neither single-command nor Matcher matches — matchedCommand stays nil.
func TestDispatch_GenericMessage_NoMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	input := newMessageOnlyInput(ctrl, botinput.TypeNewChatMembers)

	whc, responder, _, chatData := setupDispatchWHC(ctrl, input)
	// matchedCommand == nil && inputType != TypeChosenInlineResult => tries fallback then HandleUnmatched
	chatData.EXPECT().GetAwaitingReplyTo().Return("").AnyTimes()
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	// Two commands, neither has a Matcher — no match.
	router.RegisterCommandsForInputType(botinput.TypeNewChatMembers,
		botsfw.Command{Code: "ncm1", Action: dummyAction},
		botsfw.Command{Code: "ncm2", Action: dummyAction},
	)

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	handler.EXPECT().HandleUnmatched(gomock.Any()).Return(botmsg.MessageFromBot{}).AnyTimes()

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Dispatch: default case (non-standard InputType, not the above types) ─────

// TestDispatch_DefaultCase exercises the `default:` switch arm in Dispatch where
// the input type is not Unknown but also doesn't match any specific case.
func TestDispatch_DefaultCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// TypeVoice (4) does not implement CallbackQuery/InlineQuery/TextMessage/etc,
	// and MockInputMessage is not a botinput.Message subtype (no mock hierarchy).
	// We use MockInputMessage directly; its type will NOT match any case except
	// the final `default:`.
	im := newMockIM(ctrl, botinput.TypeVoice)

	whc, responder, analytics, _ := setupDispatchWHC(ctrl, im)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	// Register at least one command for TypeVoice so typeCommands.all is non-empty.
	router.RegisterCommandsForInputType(botinput.TypeVoice, botsfw.Command{
		Code:   "voice",
		Action: dummyAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Dispatch: no commandAction (matchedCommand set but action nil) ───────────

// TestDispatch_NoCommandAction covers the path where matchedCommand != nil but
// commandAction == nil (e.g. Message case with no Matcher, single command, Action==nil).
func TestDispatch_NoCommandAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	input := newMessageOnlyInput(ctrl, botinput.TypeNewChatMembers)

	whc, responder, _, _ := setupDispatchWHC(ctrl, input)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	// Single command with no Action — commandAction will remain nil after the case,
	// triggering the error path "no action for matched command".
	router.RegisterCommandsForInputType(botinput.TypeNewChatMembers, botsfw.Command{
		Code: "ncm_no_action",
		// Action is deliberately nil
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	// Expect an error return (no action for matched command)
	err := router.Dispatch(handler, responder, whc)
	if err == nil {
		t.Fatal("expected error when commandAction is nil, got nil")
	}
}

// ─── Dispatch: no commands for type + fallback handler fires ─────────────────

// TestDispatch_NoCommandsForType_FallbackFires tests the branch in Dispatch where no
// commands are registered for the input type but a fallback handler is set.
func TestDispatch_NoCommandsForType_FallbackFires(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	im := newMockIM(ctrl, botinput.TypeContact)
	ciInput := &contactInput{MockInputMessage: im}

	whc, responder, _, _ := setupDispatchWHC(ctrl, ciInput)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	fallbackFired := false
	router := NewWebhookRouter(nil)
	router.SetFallbackHandler(botinput.TypeContact, func(_ botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		fallbackFired = true
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "fb"}}, nil
	})
	// No commands registered for TypeContact.

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fallbackFired {
		t.Fatal("expected fallback to fire when no commands found for input type")
	}
}

// ─── composite input types for logInput tests ────────────────────────────────

// logInputSender is a trivial botinput.User implementation for logInput tests.
type logInputSender struct {
	id        string
	firstName string
	lastName  string
}

func (s logInputSender) GetID() any           { return s.id }
func (s logInputSender) GetFirstName() string { return s.firstName }
func (s logInputSender) GetLastName() string  { return s.lastName }
func (s logInputSender) GetUserName() string  { return "" }
func (s logInputSender) GetAvatar() string    { return "" }
func (s logInputSender) GetLanguage() string  { return "" }
func (s logInputSender) GetCountry() string   { return "" }
func (s logInputSender) IsBotUser() bool      { return false }
func (s logInputSender) Platform() string     { return "test" }

// logInputBase provides the botinput.InputMessage methods used in logInput calls.
// It embeds *mock_botinput.MockInputMessage and overrides GetSender.
type logInputBase struct {
	*mock_botinput.MockInputMessage
	sender botinput.User
}

func (b *logInputBase) GetSender() botinput.User { return b.sender }

// newLogInputBase creates a MockInputMessage with the minimum expectations for logInput.
func newLogInputBase(ctrl *gomock.Controller, inputType botinput.Type) *logInputBase {
	im := mock_botinput.NewMockInputMessage(ctrl)
	im.EXPECT().InputType().Return(inputType).AnyTimes()
	// logInput doesn't call these, but we set them to avoid unexpected-call panics.
	return &logInputBase{
		MockInputMessage: im,
		sender:           logInputSender{id: "u1", firstName: "Test", lastName: "User"},
	}
}

// ─── text message composite for logInput ─────────────────────────────────────

type logInputText struct {
	*logInputBase
	text string
}

func (l *logInputText) Text() string   { return l.text }
func (l *logInputText) IsEdited() bool { return false }

// ─── location message composite for logInput ─────────────────────────────────

type logInputLocation struct {
	*logInputBase
	lat, lon float64
}

func (l *logInputLocation) GetLatitude() float64  { return l.lat }
func (l *logInputLocation) GetLongitude() float64 { return l.lon }

// ─── new chat members composite for logInput ─────────────────────────────────

type logInputNewChatMembers struct {
	*logInputBase
	members []botinput.Actor
}

func (l *logInputNewChatMembers) NewChatMembers() []botinput.Actor { return l.members }

// logInputActor is a trivial botinput.Actor for testing.
type logInputActor struct{}

func (logInputActor) GetUserName() string  { return "alice" }
func (logInputActor) GetFirstName() string { return "Alice" }
func (logInputActor) GetLastName() string  { return "" }
func (logInputActor) GetID() any           { return "a1" }
func (logInputActor) GetAvatar() string    { return "" }
func (logInputActor) GetLanguage() string  { return "" }
func (logInputActor) IsBotUser() bool      { return false }
func (logInputActor) Platform() string     { return "test" }

// ─── contact message composite for logInput ──────────────────────────────────

type logInputContact struct {
	*logInputBase
}

func (l *logInputContact) GetPhoneNumber() string { return "+1" }
func (l *logInputContact) GetFirstName() string   { return "Bob" }
func (l *logInputContact) GetLastName() string    { return "" }
func (l *logInputContact) GetBotUserID() string   { return "b1" }
func (l *logInputContact) GetVCard() string       { return "" }

// ─── callback query composite for logInput ───────────────────────────────────

type logInputCallback struct {
	*logInputBase
	data string
}

func (l *logInputCallback) GetData() string              { return l.data }
func (l *logInputCallback) GetID() string                { return "cb1" }
func (l *logInputCallback) GetFrom() botinput.Sender     { return nil }
func (l *logInputCallback) GetMessage() botinput.Message { return nil }
func (l *logInputCallback) Chat() botinput.Chat          { return nil }

// ─── inline query composite for logInput ─────────────────────────────────────

type logInputInlineQuery struct {
	*logInputBase
	query string
}

func (l *logInputInlineQuery) GetID() any               { return "iq1" }
func (l *logInputInlineQuery) GetInlineQueryID() string { return "iq1" }
func (l *logInputInlineQuery) GetFrom() botinput.Sender { return nil }
func (l *logInputInlineQuery) GetQuery() string         { return l.query }
func (l *logInputInlineQuery) GetOffset() string        { return "" }

// ─── chosen inline result composite for logInput ─────────────────────────────

type logInputChosenInlineResult struct {
	*logInputBase
}

func (l *logInputChosenInlineResult) GetResultID() string        { return "r1" }
func (l *logInputChosenInlineResult) GetInlineMessageID() string { return "im1" }
func (l *logInputChosenInlineResult) GetFrom() botinput.Sender   { return nil }
func (l *logInputChosenInlineResult) GetQuery() string           { return "" }

// ─── referral message composite for logInput ─────────────────────────────────

// logInputReferral implements both botinput.ReferralMessage and botinput.TextMessage.
// NOTE: In logInput's type switch, TextMessage appears before ReferralMessage, so
// this type matches the TextMessage case. The ReferralMessage case in logInput is
// effectively unreachable without a type that implements ReferralMessage but not
// TextMessage AND has a TextMessage cast that doesn't panic — which is impossible
// without production types. We accept this one uncovered branch.
type logInputReferral struct {
	*logInputBase
}

func (l *logInputReferral) Type() string    { return "ref" }
func (l *logInputReferral) Source() string  { return "web" }
func (l *logInputReferral) RefData() string { return "d1" }
func (l *logInputReferral) Text() string    { return "ref-text" }
func (l *logInputReferral) IsEdited() bool  { return false }

// ─── shared users composite for logInput ─────────────────────────────────────

type logInputSharedUsers struct {
	*logInputBase
}

func (l *logInputSharedUsers) GetRequestID() int                                { return 1 }
func (l *logInputSharedUsers) GetSharedUsers() []botinput.SharedUserMessageItem { return nil }

// ─── webhookDriver.logInput tests ────────────────────────────────────────────

func TestLogInput_TextMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeText)
	input := &logInputText{logInputBase: base, text: "hello world"}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_LocationMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeLocation)
	input := &logInputLocation{logInputBase: base, lat: 51.5, lon: -0.1}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_NewChatMembersMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeNewChatMembers)
	input := &logInputNewChatMembers{
		logInputBase: base,
		members:      []botinput.Actor{logInputActor{}},
	}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_ContactMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeContact)
	input := &logInputContact{logInputBase: base}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_CallbackQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeCallbackQuery)
	input := &logInputCallback{logInputBase: base, data: "settings"}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_InlineQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeInlineQuery)
	input := &logInputInlineQuery{logInputBase: base, query: "search term"}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_ChosenInlineResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeChosenInlineResult)
	input := &logInputChosenInlineResult{logInputBase: base}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_ReferralMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeReferral)
	input := &logInputReferral{logInputBase: base}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_SharedUsersMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := newLogInputBase(ctrl, botinput.TypeSharedUsers)
	input := &logInputSharedUsers{logInputBase: base}

	d := webhookDriver{}
	d.logInput(context.Background(), 0, input)
}

func TestLogInput_DefaultCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Use a plain MockInputMessage (not a TextMessage/LocationMessage/etc)
	// This hits the `default:` case in logInput's switch.
	base := newLogInputBase(ctrl, botinput.TypeVoice)

	d := webhookDriver{}
	d.logInput(context.Background(), 1, base)
}

// ─── webhookDriver.reportPanicToAnalytics ─────────────────────────────────────

func TestReportPanicToAnalytics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	analytics.EXPECT().Enqueue(gomock.Any()).Times(1)

	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().Analytics().Return(analytics).Times(1)

	d := webhookDriver{}
	d.reportPanicToAnalytics(context.Background(), whc, "panic message: something blew up")
}

// ─── Dispatch: fallback for matchedCommand==nil and non-ChosenInlineResult ────

// TestDispatch_UnmatchedText_FallbackForInputType tests the "matchedCommand==nil"
// path that tries the fallback before calling HandleUnmatched, but here the fallback
// is registered for the input type (not the no-commands path).
func TestDispatch_UnmatchedText_FallbackForInputType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	textInput.EXPECT().Text().Return("no match text").AnyTimes()
	textInput.EXPECT().IsEdited().Return(false).AnyTimes()

	whc, responder, _, _ := setupDispatchWHC(ctrl, textInput)
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	fallbackFired := false
	router := NewWebhookRouter(nil)
	router.RegisterCommands(botsfw.Command{
		Code:       "cmd1",
		TextAction: dummyTextAction,
	})
	// Fallback fires when matchedCommand==nil (no command matched the text).
	router.SetFallbackHandler(botinput.TypeText, func(_ botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		fallbackFired = true
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "fallback"}}, nil
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	err := router.Dispatch(handler, responder, whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fallbackFired {
		t.Fatal("expected fallback to fire when no command matched")
	}
}

// ─── Dispatch: SaveBotChat error path ────────────────────────────────────────

// TestDispatch_SaveBotChat_Error covers the branch where SaveBotChat returns
// an error after a command succeeds (chatData.IsChanged() == true).
func TestDispatch_SaveBotChat_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	// SendMessage called twice: once for the save-error notice, once for the command response.
	responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(botsfw.OnMessageSentResponse{}, nil).AnyTimes()

	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	analytics.EXPECT().Enqueue(gomock.Any()).AnyTimes()
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

	chatData := mock_botsfwmodels.NewMockBotChatData(ctrl)
	chatData.EXPECT().GetAwaitingReplyTo().Return("").AnyTimes()
	chatData.EXPECT().IsChanged().Return(true).AnyTimes()
	chatData.EXPECT().HasChangedVars().Return(false).AnyTimes()
	chatData.EXPECT().SetDtLastInteraction(gomock.Any()).AnyTimes()
	chatData.EXPECT().SetUpdatedTime(gomock.Any()).AnyTimes()
	whc.EXPECT().ChatData().Return(chatData).AnyTimes()
	// SaveBotChat returns an error to trigger the error-handling branch.
	whc.EXPECT().SaveBotChat().Return(errors.New("db unavailable")).AnyTimes()

	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:       "help",
		TextAction: dummyTextAction,
	})

	handler := mock_botsfw.NewMockWebhookHandler(ctrl)
	err := router.Dispatch(handler, responder, whc)
	// Dispatch propagates the SaveBotChat error (after logging and notifying the user),
	// so we expect a non-nil error here.
	if err == nil {
		t.Fatal("expected Dispatch to surface the SaveBotChat error, got nil")
	}
}

// ─── Dispatch: RegisterCommands with RefundedPaymentAction only ──────────────

// TestRegisterCommands_RefundedPayment ensures that a command with only
// RefundedPaymentAction gets added for TypeRefundedPayment (covers the addCommand
// path at line ~172 of router.go).
func TestRegisterCommands_RefundedPayment(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:                  "refund",
		RefundedPaymentAction: dummyRefundedPaymentAction,
	})
	cmds := router.RegisteredCommands()
	if len(cmds[botinput.TypeRefundedPayment]) != 1 {
		t.Fatalf("expected 1 command for TypeRefundedPayment, got %d", len(cmds[botinput.TypeRefundedPayment]))
	}
}
