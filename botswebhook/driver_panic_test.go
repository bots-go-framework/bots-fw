package botswebhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwstore/botsfwstoretest"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"go.uber.org/mock/gomock"
)

func TestProcessWebhookInput_PanicResponseIsSanitized(t *testing.T) {
	const (
		panicSecret  = "PANIC-DETAIL-MUST-STAY-INTERNAL"
		footerSecret = "PRIVATE-OPERATOR-FOOTER"
	)

	ctrl := gomock.NewController(t)
	input := newDriverPanicTextInput(ctrl)
	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	handler.EXPECT().
		CreateWebhookContext(gomock.Any()).
		Return(whc, nil)
	whc.EXPECT().
		ChatData().
		DoAndReturn(func() any {
			panic(panicSecret)
		})
	whc.EXPECT().Analytics().Return(analytics)
	analytics.EXPECT().
		Enqueue(gomock.Any()).
		Do(func(any) {
			panic("analytics re-entered the failing context")
		})
	whc.EXPECT().Input().Return(input)
	whc.EXPECT().Responder().Return(responder)

	wantText := ErrorIcon + " " + panicUserMessage
	whc.EXPECT().
		NewMessage(wantText).
		Return(botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: wantText}})

	var sent botmsg.MessageFromBot
	responder.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), botsfw.BotAPISendMessageOverResponse).
		DoAndReturn(func(_ context.Context, m botmsg.MessageFromBot, _ botmsg.BotAPISendMessageChannel) (botsfw.OnMessageSentResponse, error) {
			sent = m
			return botsfw.OnMessageSentResponse{}, nil
		})

	driver := webhookDriver{
		Analytics: AnalyticsSettings{
			Enabled: func(*http.Request) bool { return true },
		},
		panicTextFooter: footerSecret,
	}
	botContext := panicTestBotContext()
	request := httptest.NewRequest(http.MethodPost, "/webhook", nil)

	if err := driver.processWebhookInput(
		context.Background(),
		httptest.NewRecorder(),
		request,
		handler,
		botContext,
		0,
		input,
		func(error, string) {},
	); err != nil {
		t.Fatalf("processWebhookInput() error = %v", err)
	}

	if sent.Text != wantText {
		t.Fatalf("sent text = %q, want %q", sent.Text, wantText)
	}
	for _, privateValue := range []string{panicSecret, footerSecret, "Stack trace", "driver_panic_test.go"} {
		if strings.Contains(sent.Text, privateValue) {
			t.Fatalf("panic response exposes private diagnostic %q: %q", privateValue, sent.Text)
		}
	}
}

func TestProcessWebhookInput_PanicInUserNotificationDoesNotRepanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	input := newDriverPanicTextInput(ctrl)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	handler := mock_botsfw.NewMockWebhookHandler(ctrl)

	handler.EXPECT().
		CreateWebhookContext(gomock.Any()).
		Return(whc, nil)
	whc.EXPECT().
		ChatData().
		DoAndReturn(func() any {
			panic("initial panic")
		})
	whc.EXPECT().
		Input().
		DoAndReturn(func() botinput.InputMessage {
			panic("notification hook panic")
		})

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("processWebhookInput() recovery panicked again: %v", recovered)
		}
	}()

	driver := webhookDriver{}
	_ = driver.processWebhookInput(
		context.Background(),
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/webhook", nil),
		handler,
		panicTestBotContext(),
		0,
		input,
		func(error, string) {},
	)
}

func TestProcessWebhookInput_PanicBeforeWebhookContextDoesNotRepanic(t *testing.T) {
	driver := webhookDriver{
		Analytics: AnalyticsSettings{
			Enabled: func(*http.Request) bool { return true },
		},
		panicTextFooter: "private footer",
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("processWebhookInput() recovery panicked again: %v", recovered)
		}
	}()

	_ = driver.processWebhookInput(
		context.Background(),
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/webhook", nil),
		nil,
		panicTestBotContext(),
		0,
		nil,
		func(error, string) {},
	)
}

func newDriverPanicTextInput(ctrl *gomock.Controller) *mock_botinput.MockTextMessage {
	user := mock_botinput.NewMockUser(ctrl)
	user.EXPECT().GetID().Return("user-id")
	user.EXPECT().GetFirstName().Return("First")
	user.EXPECT().GetLastName().Return("Last")

	input := mock_botinput.NewMockTextMessage(ctrl)
	input.EXPECT().GetSender().Return(user)
	input.EXPECT().Text().Return("request text")
	input.EXPECT().BotChatID().Return("chat-id", nil).AnyTimes()
	return input
}

func panicTestBotContext() *botsfw.BotContext {
	return &botsfw.BotContext{
		BotSettings: &botsfw.BotSettings{
			Code:  "panic-test",
			Env:   botsfw.EnvLocal,
			Store: &botsfwstoretest.FakeStateStore{},
		},
	}
}
