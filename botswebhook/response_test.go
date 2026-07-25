package botswebhook

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwstore/botsfwstoretest"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"go.uber.org/mock/gomock"
)

func TestHandleWebhook_PreCommitFailuresAreGeneric(t *testing.T) {
	const secret = "PRIVATE-PROCESS-DETAIL"

	tests := []struct {
		name string
		get  func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error)
	}{
		{
			name: "error",
			get: func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
				return nil, nil, errors.New(secret)
			},
		},
		{
			name: "panic",
			get: func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
				panic(secret)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			driver := webhookDriver{botHost: testBotHostForDriver{}}
			driver.HandleWebhook(
				recorder,
				httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil),
				envelopeWebhookHandler{get: test.get},
			)

			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusInternalServerError)
			}
			if body := recorder.Body.String(); body != http.StatusText(http.StatusInternalServerError)+"\n" {
				t.Fatalf("HTTP body = %q, want generic error", body)
			}
			if strings.Contains(recorder.Body.String(), secret) {
				t.Fatalf("HTTP body exposes private error: %q", recorder.Body.String())
			}
		})
	}
}

func TestHandleWebhook_PreCommitProcessErrorIsSanitized(t *testing.T) {
	const secret = "PRIVATE-DISPATCH-DETAIL"

	ctrl := gomock.NewController(t)
	input := newDriverPanicTextInput(ctrl)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().ChatData().Return(nil)

	router := envelopeRouter{
		dispatch: func(botsfw.WebhookHandler, botsfw.WebhookResponder, botsfw.WebhookContext) error {
			return errors.New(secret)
		},
	}
	botContext := &botsfw.BotContext{
		BotSettings: &botsfw.BotSettings{
			Code:    "response-test",
			Env:     botsfw.EnvLocal,
			Store:   &botsfwstoretest.FakeStateStore{},
			Profile: envelopeBotProfile{router: router},
		},
	}
	handler := envelopeWebhookHandler{
		get: func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
			return botContext, []botinput.EntryInputs{{Inputs: []botinput.InputMessage{input}}}, nil
		},
		create: func(botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
			return whc, nil
		},
		responder: func(http.ResponseWriter, botsfw.WebhookContext) botsfw.WebhookResponder {
			return nil
		},
	}
	recorder := httptest.NewRecorder()

	webhookDriver{botHost: testBotHostForDriver{}}.HandleWebhook(
		recorder,
		httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil),
		handler,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if body := recorder.Body.String(); body != http.StatusText(http.StatusInternalServerError)+"\n" {
		t.Fatalf("HTTP body = %q, want generic error", body)
	}
	if strings.Contains(recorder.Body.String(), secret) {
		t.Fatalf("HTTP body exposes private process error: %q", recorder.Body.String())
	}
}

func TestHandleWebhook_PostCommitFailuresDoNotRewriteResponse(t *testing.T) {
	const (
		committedBody = "platform response"
		secret        = "PRIVATE-POST-COMMIT-DETAIL"
	)

	tests := []struct {
		name     string
		dispatch func() error
	}{
		{name: "error", dispatch: func() error { return errors.New(secret) }},
		{name: "panic", dispatch: func() error { panic(secret) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			input := newDriverPanicTextInput(ctrl)
			whc := mock_botsfw.NewMockWebhookContext(ctrl)
			whc.EXPECT().ChatData().Return(nil)
			if test.name == "panic" {
				whc.EXPECT().Input().Return(input)
				whc.EXPECT().Responder().Return(nil)
			}

			router := envelopeRouter{
				dispatch: func(botsfw.WebhookHandler, botsfw.WebhookResponder, botsfw.WebhookContext) error {
					return test.dispatch()
				},
			}
			botContext := &botsfw.BotContext{
				BotSettings: &botsfw.BotSettings{
					Code:    "response-test",
					Env:     botsfw.EnvLocal,
					Store:   &botsfwstoretest.FakeStateStore{},
					Profile: envelopeBotProfile{router: router},
				},
			}
			handler := envelopeWebhookHandler{
				get: func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
					return botContext, []botinput.EntryInputs{{Inputs: []botinput.InputMessage{input}}}, nil
				},
				create: func(botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
					return whc, nil
				},
				responder: func(w http.ResponseWriter, _ botsfw.WebhookContext) botsfw.WebhookResponder {
					w.WriteHeader(http.StatusAccepted)
					_, _ = io.WriteString(w, committedBody)
					return nil
				},
			}

			recorder := httptest.NewRecorder()
			driver := webhookDriver{botHost: testBotHostForDriver{}}
			driver.HandleWebhook(
				recorder,
				httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil),
				handler,
			)

			if recorder.Code != http.StatusAccepted {
				t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusAccepted)
			}
			if body := recorder.Body.String(); body != committedBody {
				t.Fatalf("HTTP body = %q, want unchanged %q", body, committedBody)
			}
			if strings.Contains(recorder.Body.String(), secret) {
				t.Fatalf("HTTP body exposes private error: %q", recorder.Body.String())
			}
		})
	}
}

func TestHandleWebhook_RejectsOversizeBody(t *testing.T) {
	handlerCalled := false
	handler := envelopeWebhookHandler{
		get: func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
			handlerCalled = true
			return nil, nil, nil
		},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"http://localhost/webhook",
		io.LimitReader(strings.NewReader(strings.Repeat("x", 1024)), MaxWebhookRequestBodyBytes+1),
	)
	request.ContentLength = MaxWebhookRequestBodyBytes + 1
	recorder := httptest.NewRecorder()

	webhookDriver{botHost: testBotHostForDriver{}}.HandleWebhook(recorder, request, handler)

	if handlerCalled {
		t.Fatal("platform handler was called for an oversized request")
	}
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if body := recorder.Body.String(); body != http.StatusText(http.StatusRequestEntityTooLarge)+"\n" {
		t.Fatalf("HTTP body = %q, want generic request-too-large error", body)
	}
}

func TestHandleWebhook_BoundsBodyWithUnknownLength(t *testing.T) {
	handler := envelopeWebhookHandler{
		get: func(_ context.Context, request *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
			_, err := io.ReadAll(request.Body)
			return nil, nil, err
		},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"http://localhost/webhook",
		strings.NewReader(strings.Repeat("x", int(MaxWebhookRequestBodyBytes)+1)),
	)
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	webhookDriver{botHost: testBotHostForDriver{}}.HandleWebhook(recorder, request, handler)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestWebhookResponse_PreservesOptionalInterfaces(t *testing.T) {
	minimal := newWebhookResponse(&minimalResponseWriter{header: make(http.Header)})
	if _, ok := minimal.writer.(http.Flusher); ok {
		t.Fatal("minimal writer unexpectedly implements http.Flusher")
	}
	if _, ok := minimal.writer.(http.Hijacker); ok {
		t.Fatal("minimal writer unexpectedly implements http.Hijacker")
	}
	if _, ok := minimal.writer.(http.Pusher); ok {
		t.Fatal("minimal writer unexpectedly implements http.Pusher")
	}

	richRaw := &optionalResponseWriter{minimalResponseWriter: minimalResponseWriter{header: make(http.Header)}}
	rich := newWebhookResponse(richRaw)
	if _, ok := rich.writer.(http.Flusher); !ok {
		t.Fatal("wrapped writer does not preserve http.Flusher")
	}
	if _, ok := rich.writer.(http.Hijacker); !ok {
		t.Fatal("wrapped writer does not preserve http.Hijacker")
	}
	if _, ok := rich.writer.(http.Pusher); !ok {
		t.Fatal("wrapped writer does not preserve http.Pusher")
	}
	if _, ok := rich.writer.(io.ReaderFrom); !ok {
		t.Fatal("wrapped writer does not preserve io.ReaderFrom")
	}
	if _, ok := rich.writer.(io.StringWriter); !ok {
		t.Fatal("wrapped writer does not preserve io.StringWriter")
	}

	rich.writer.(http.Flusher).Flush()
	if !rich.isCommitted() || !richRaw.flushed {
		t.Fatal("Flush did not commit and delegate to the underlying writer")
	}
}

type envelopeWebhookHandler struct {
	botsfw.WebhookHandler
	get       func(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error)
	create    func(botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error)
	responder func(http.ResponseWriter, botsfw.WebhookContext) botsfw.WebhookResponder
}

func (h envelopeWebhookHandler) GetBotContextAndInputs(
	ctx context.Context,
	request *http.Request,
) (*botsfw.BotContext, []botinput.EntryInputs, error) {
	return h.get(ctx, request)
}

func (h envelopeWebhookHandler) CreateWebhookContext(args botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
	return h.create(args)
}

func (h envelopeWebhookHandler) GetResponder(w http.ResponseWriter, whc botsfw.WebhookContext) botsfw.WebhookResponder {
	return h.responder(w, whc)
}

type envelopeRouter struct {
	botsfw.Router
	dispatch func(botsfw.WebhookHandler, botsfw.WebhookResponder, botsfw.WebhookContext) error
}

func (r envelopeRouter) Dispatch(
	handler botsfw.WebhookHandler,
	responder botsfw.WebhookResponder,
	whc botsfw.WebhookContext,
) error {
	return r.dispatch(handler, responder, whc)
}

type envelopeBotProfile struct {
	botsfw.BotProfile
	router botsfw.Router
}

func (p envelopeBotProfile) Router() botsfw.Router {
	return p.router
}

type minimalResponseWriter struct {
	header http.Header
	status int
	body   strings.Builder
}

func (w *minimalResponseWriter) Header() http.Header { return w.header }
func (w *minimalResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}
func (w *minimalResponseWriter) Write(body []byte) (int, error) {
	return w.body.Write(body)
}

type optionalResponseWriter struct {
	minimalResponseWriter
	flushed bool
}

func (w *optionalResponseWriter) Flush() {
	w.flushed = true
}
func (*optionalResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("not connected")
}
func (*optionalResponseWriter) Push(string, *http.PushOptions) error {
	return nil
}
func (w *optionalResponseWriter) ReadFrom(source io.Reader) (int64, error) {
	return io.Copy(&w.body, source)
}
func (w *optionalResponseWriter) WriteString(body string) (int, error) {
	return w.body.WriteString(body)
}
