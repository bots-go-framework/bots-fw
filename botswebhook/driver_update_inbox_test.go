package botswebhook

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore/botsfwstoretest"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/bots-go-framework/bots-fw/botsfwtest"
	"github.com/strongo/i18n"
)

type updateInboxTestHandler struct {
	createWebhookContext func(botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error)
}

func (updateInboxTestHandler) RegisterHttpHandlers(botsfw.WebhookDriver, botsfw.BotHost, botsfw.HttpRouter, string) {
}

func (updateInboxTestHandler) HandleWebhookRequest(http.ResponseWriter, *http.Request) {}

func (updateInboxTestHandler) GetBotContextAndInputs(context.Context, *http.Request) (*botsfw.BotContext, []botinput.EntryInputs, error) {
	return nil, nil, errors.New("not used by processWebhookEntry test")
}

func (h updateInboxTestHandler) CreateWebhookContext(args botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
	return h.createWebhookContext(args)
}

func (updateInboxTestHandler) GetResponder(_ http.ResponseWriter, whc botsfw.WebhookContext) botsfw.WebhookResponder {
	return whc.Responder()
}

func (updateInboxTestHandler) HandleUnmatched(botsfw.WebhookContext) botmsg.MessageFromBot {
	return botmsg.MessageFromBot{}
}

func TestProcessWebhookEntryClaimsWholeProviderEntry(t *testing.T) {
	tests := []struct {
		name           string
		failInput      int
		wantErr        bool
		wantCompleted  int
		wantFailed     int
		wantDispatched int
	}{
		{
			name:           "all sibling inputs complete under one claim",
			failInput:      -1,
			wantCompleted:  1,
			wantDispatched: 2,
		},
		{
			name:           "partial failure fails the entry once",
			failInput:      1,
			wantErr:        true,
			wantFailed:     1,
			wantDispatched: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var claimed, completed, failed, dispatched int
			const leaseID = "lease-1"
			store := &botsfwstoretest.FakeStateStore{
				ClaimWebhookUpdateFunc: func(_ context.Context, key botsfwstore.WebhookUpdateKey, _ time.Time) (botsfwstore.WebhookUpdateClaim, error) {
					claimed++
					if key.UpdateID != "provider-entry-42" {
						t.Fatalf("claimed update ID = %q", key.UpdateID)
					}
					return botsfwstore.WebhookUpdateClaim{
						Status:  botsfwstore.WebhookUpdateClaimAcquired,
						LeaseID: leaseID,
					}, nil
				},
				CompleteWebhookUpdateFunc: func(_ context.Context, _ botsfwstore.WebhookUpdateKey, gotLeaseID string) error {
					completed++
					if gotLeaseID != leaseID {
						t.Fatalf("completed lease ID = %q", gotLeaseID)
					}
					return nil
				},
				FailWebhookUpdateFunc: func(_ context.Context, _ botsfwstore.WebhookUpdateKey, gotLeaseID string, code botsfwstore.WebhookUpdateFailureCode) error {
					failed++
					if gotLeaseID != leaseID {
						t.Fatalf("failed lease ID = %q", gotLeaseID)
					}
					if code != botsfwstore.WebhookUpdateFailureProcessing {
						t.Fatalf("failure code = %q", code)
					}
					return nil
				},
			}

			router := NewWebhookRouter(nil)
			router.SetFallbackHandler(botinput.TypeText, func(botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
				dispatched++
				return botmsg.MessageFromBot{}, nil
			})
			profile := botsfw.NewBotProfile(
				"update-inbox-test",
				router,
				func() botsfwmodels.BotChatData { return nil },
				func() botsfwmodels.PlatformUserData { return nil },
				i18n.LocaleEnUS,
				nil,
				botsfw.BotTranslations{},
			)
			botContext := &botsfw.BotContext{BotSettings: &botsfw.BotSettings{
				Platform: botsfwconst.PlatformTelegram,
				Code:     "test-bot",
				Profile:  profile,
				Store:    store,
			}}

			contexts := []*botsfwtest.FakeWebhookContext{
				botsfwtest.NewFakeWebhookContext(botsfwtest.WithTextMessage("first")),
				botsfwtest.NewFakeWebhookContext(botsfwtest.WithTextMessage("second")),
			}
			inputIndex := map[botinput.InputMessage]int{
				contexts[0].Input(): 0,
				contexts[1].Input(): 1,
			}
			handler := updateInboxTestHandler{
				createWebhookContext: func(args botsfw.CreateWebhookContextArgs) (botsfw.WebhookContext, error) {
					index := inputIndex[args.WebhookInput]
					if index == tt.failInput {
						return nil, errors.New("injected context failure")
					}
					return contexts[index], nil
				},
			}
			entry := botinput.EntryInputs{
				Entry:  testDurableWebhookEntry{updateID: "provider-entry-42", ok: true},
				Inputs: []botinput.InputMessage{contexts[0].Input(), contexts[1].Input()},
			}

			err := (webhookDriver{}).processWebhookEntry(
				context.Background(),
				newWebhookResponse(httptest.NewRecorder()),
				httptest.NewRequest(http.MethodPost, "/webhook", nil),
				handler,
				botContext,
				entry,
				func(error, string) {},
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("processWebhookEntry() error = %v, wantErr %t", err, tt.wantErr)
			}
			if claimed != 1 {
				t.Fatalf("ClaimWebhookUpdate calls = %d, want 1", claimed)
			}
			if completed != tt.wantCompleted {
				t.Fatalf("CompleteWebhookUpdate calls = %d, want %d", completed, tt.wantCompleted)
			}
			if failed != tt.wantFailed {
				t.Fatalf("FailWebhookUpdate calls = %d, want %d", failed, tt.wantFailed)
			}
			if dispatched != tt.wantDispatched {
				t.Fatalf("router dispatches = %d, want %d", dispatched, tt.wantDispatched)
			}
		})
	}
}

func TestProcessWebhookEntryLeasedResponseIsCommitAware(t *testing.T) {
	tests := []struct {
		name       string
		commit     bool
		wantStatus int
		wantBody   string
	}{
		{
			name:       "pre-commit returns generic retry status",
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   http.StatusText(http.StatusServiceUnavailable) + "\n",
		},
		{
			name:       "post-commit leaves platform response unchanged",
			commit:     true,
			wantStatus: http.StatusAccepted,
			wantBody:   "platform response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &botsfwstoretest.FakeStateStore{
				ClaimWebhookUpdateFunc: func(context.Context, botsfwstore.WebhookUpdateKey, time.Time) (botsfwstore.WebhookUpdateClaim, error) {
					return botsfwstore.WebhookUpdateClaim{Status: botsfwstore.WebhookUpdateClaimLeased}, nil
				},
			}
			botContext := &botsfw.BotContext{BotSettings: &botsfw.BotSettings{
				Platform: botsfwconst.PlatformTelegram,
				Code:     "test-bot",
				Store:    store,
			}}
			recorder := httptest.NewRecorder()
			response := newWebhookResponse(recorder)
			if tt.commit {
				response.writer.WriteHeader(http.StatusAccepted)
				_, _ = io.WriteString(response.writer, tt.wantBody)
			}

			err := (webhookDriver{}).processWebhookEntry(
				context.Background(),
				response,
				httptest.NewRequest(http.MethodPost, "/webhook", nil),
				updateInboxTestHandler{},
				botContext,
				botinput.EntryInputs{Entry: testDurableWebhookEntry{updateID: "provider-entry-42", ok: true}},
				func(error, string) {},
			)
			if err != nil {
				t.Fatalf("processWebhookEntry() error = %v", err)
			}
			if recorder.Code != tt.wantStatus {
				t.Fatalf("HTTP status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if body := recorder.Body.String(); body != tt.wantBody {
				t.Fatalf("HTTP body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
