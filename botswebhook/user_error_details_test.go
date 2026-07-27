package botswebhook

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

func TestExpandableUserErrorMessage_ClassifiesTelegramRejectionInEnglish(t *testing.T) {
	err := newTelegramProviderError(t, "123456:BOT-TOKEN", http.StatusBadRequest, "Bad Request: RICH_MESSAGE_DATE_TOO_LONG")
	whc := newUserErrorDetailsWebhookContext(
		t,
		i18n.LocaleEnUK,
		"123456:BOT-TOKEN",
		"request-en-123",
	)

	message, ok := expandableUserErrorMessage(whc, err, "")
	if !ok {
		t.Fatal("expandableUserErrorMessage() = false, want true")
	}
	const wantVisible = "⚠️ Telegram rejected this message, so it wasn’t delivered.\n\n" +
		"Reason: RICH_MESSAGE_DATE_TOO_LONG\n\n"
	if !strings.HasPrefix(message.Text, wantVisible) {
		t.Fatalf("message prefix = %q, want %q", message.Text, wantVisible)
	}
	assertExpandableProviderDetails(t, message,
		"<blockquote expandable><b>🔎 Technical details</b>",
		"Telegram method: sendRichMessage",
		"Telegram error code: 400",
		"Telegram description: Bad Request: RICH_MESSAGE_DATE_TOO_LONG",
		"Request ID: request-en-123",
		"Error chain:",
	)
}

func TestExpandableUserErrorMessage_ClassifiesTelegramRejectionInRussian(t *testing.T) {
	err := newTelegramProviderError(t, "123456:BOT-TOKEN", http.StatusBadRequest, "Bad Request: RICH_MESSAGE_DATE_TOO_LONG")
	whc := newUserErrorDetailsWebhookContext(
		t,
		i18n.LocaleRuRU,
		"123456:BOT-TOKEN",
		"request-ru-123",
	)

	message, ok := expandableUserErrorMessage(whc, err, "")
	if !ok {
		t.Fatal("expandableUserErrorMessage() = false, want true")
	}
	const wantVisible = "⚠️ Telegram отклонил это сообщение, поэтому оно не было доставлено.\n\n" +
		"Причина: RICH_MESSAGE_DATE_TOO_LONG\n\n"
	if !strings.HasPrefix(message.Text, wantVisible) {
		t.Fatalf("message prefix = %q, want %q", message.Text, wantVisible)
	}
	assertExpandableProviderDetails(t, message,
		"<blockquote expandable><b>🔎 Технические подробности</b>",
		"Метод Telegram: sendRichMessage",
		"Код ошибки Telegram: 400",
		"Описание Telegram: Bad Request: RICH_MESSAGE_DATE_TOO_LONG",
		"ID запроса: request-ru-123",
		"Цепочка ошибок:",
	)
}

func TestExpandableUserErrorMessage_GenericErrorKeepsFriendlyCopy(t *testing.T) {
	err := fmt.Errorf("loading preferences: %w", fmt.Errorf("storage unavailable"))
	whc := newUserErrorDetailsWebhookContext(t, i18n.LocaleEnUK, "", "", "Friendly oops")

	message, ok := expandableUserErrorMessage(whc, err, "Contact support")
	if !ok {
		t.Fatal("expandableUserErrorMessage() = false, want true")
	}
	const wantVisible = "Friendly oops\n\n" +
		"💢 Server error — failed to process message.\n\n" +
		"Contact support\n\n"
	if !strings.HasPrefix(message.Text, wantVisible) {
		t.Fatalf("message prefix = %q, want %q", message.Text, wantVisible)
	}
	for _, want := range []string{
		"🔎 Technical details",
		"1. loading preferences: storage unavailable",
		"2. storage unavailable",
	} {
		if !strings.Contains(message.Text, want) {
			t.Errorf("message does not contain %q: %s", want, message.Text)
		}
	}
}

func TestExpandableUserErrorMessage_RedactsSecretsRequestBodiesAndPrivateState(t *testing.T) {
	const (
		botToken          = "123456:PRIVATE-BOT-TOKEN"
		paymentToken      = "PRIVATE-PAYMENT-TOKEN"
		webhookSecret     = "PRIVATE-WEBHOOK-SECRET"
		authorization     = "PRIVATE-BEARER-CREDENTIAL"
		requestBody       = `{"private":"RAW-REQUEST-BODY"}`
		requestURL        = "https://api.telegram.org/bot123456:PRIVATE-BOT-TOKEN/sendMessage?body=RAW-REQUEST-BODY"
		walletPrivateData = "WALLET-BALANCE-AND-NONCE"
		gamePrivateData   = "PRIVATE-CARDS"
	)
	err := fmt.Errorf(
		"delivery failed token=%s\npayment_token=%s\nwebhook_secret=%s\nAuthorization: Bearer %s\nrequest URL=%s\nrequest body=%s\nwallet_private_state=%s\ngame_private_state=%s",
		botToken,
		paymentToken,
		webhookSecret,
		authorization,
		requestURL,
		requestBody,
		walletPrivateData,
		gamePrivateData,
	)
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	settings := &botsfw.BotSettings{
		Platform:           botsfwconst.PlatformTelegram,
		Token:              botToken,
		PaymentToken:       paymentToken,
		WebhookSecretToken: webhookSecret,
		UserErrorDetails: botsfw.UserErrorDetailsPolicy{
			Disclosure: botsfw.UserErrorDetailsDisclosureExpandable,
		},
	}
	whc.EXPECT().GetBotSettings().Return(settings).AnyTimes()
	whc.EXPECT().Locale().Return(i18n.LocaleEnUK).AnyTimes()
	whc.EXPECT().Request().Return(&http.Request{Header: make(http.Header)}).AnyTimes()
	whc.EXPECT().Translate(botsfw.MessageTextOopsSomethingWentWrong).Return("Friendly oops")
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(text string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
	})

	message, ok := expandableUserErrorMessage(whc, err, "")
	if !ok {
		t.Fatal("expandableUserErrorMessage() = false, want true")
	}
	for _, forbidden := range []string{
		botToken,
		paymentToken,
		webhookSecret,
		authorization,
		requestURL,
		"api.telegram.org",
		requestBody,
		"RAW-REQUEST-BODY",
		walletPrivateData,
		gamePrivateData,
	} {
		if strings.Contains(message.Text, forbidden) {
			t.Errorf("message leaks %q: %s", forbidden, message.Text)
		}
	}
	if !strings.Contains(message.Text, "[REDACTED]") {
		t.Errorf("message does not preserve redaction markers: %s", message.Text)
	}
}

func TestExpandableUserErrorMessage_RequiresExplicitPerBotOptIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	whc.EXPECT().GetBotSettings().Return(&botsfw.BotSettings{
		Platform: botsfwconst.PlatformTelegram,
	})

	if _, ok := expandableUserErrorMessage(whc, fmt.Errorf("private error"), ""); ok {
		t.Fatal("expandableUserErrorMessage() = true without opt-in")
	}
}

func TestProcessCommandResponseError_RetriesPlainTextAfterTelegramRejectsExpandableHTML(t *testing.T) {
	ctrl := gomock.NewController(t)
	im := newMockIM(ctrl, botinput.TypeText)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	settings := &botsfw.BotSettings{
		Env:      botsfw.EnvProduction,
		Platform: botsfwconst.PlatformTelegram,
		Token:    "123456:BOT-TOKEN",
		UserErrorDetails: botsfw.UserErrorDetailsPolicy{
			Disclosure: botsfw.UserErrorDetailsDisclosureExpandable,
		},
	}
	whc.EXPECT().Context().Return(context.Background()).AnyTimes()
	whc.EXPECT().Input().Return(im).AnyTimes()
	whc.EXPECT().GetBotSettings().Return(settings).AnyTimes()
	whc.EXPECT().GetBotCode().Return("SneatBot").AnyTimes()
	whc.EXPECT().Locale().Return(i18n.LocaleEnUK).AnyTimes()
	whc.EXPECT().Request().Return(&http.Request{Header: make(http.Header)}).AnyTimes()
	whc.EXPECT().Translate(botsfw.MessageTextOopsSomethingWentWrong).Return("Friendly oops")
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(text string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
	})
	analytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
	analytics.EXPECT().Enqueue(gomock.Any())
	whc.EXPECT().Analytics().Return(analytics)

	htmlRejection := newTelegramProviderError(t, settings.Token, http.StatusBadRequest, "Bad Request: can't parse entities")
	responder := mock_botsfw.NewMockWebhookResponder(ctrl)
	gomock.InOrder(
		responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), botsfw.BotAPISendMessageOverResponse).
			DoAndReturn(func(_ context.Context, message botmsg.MessageFromBot, _ botmsg.BotAPISendMessageChannel) (botsfw.OnMessageSentResponse, error) {
				if message.Format != botmsg.FormatHTML {
					t.Errorf("first message format = %v, want HTML", message.Format)
				}
				return botsfw.OnMessageSentResponse{}, htmlRejection
			}),
		responder.EXPECT().SendMessage(gomock.Any(), gomock.Any(), botsfw.BotAPISendMessageOverResponse).
			DoAndReturn(func(_ context.Context, message botmsg.MessageFromBot, _ botmsg.BotAPISendMessageChannel) (botsfw.OnMessageSentResponse, error) {
				if message.Format != botmsg.FormatText {
					t.Errorf("fallback message format = %v, want plain text", message.Format)
				}
				if strings.Contains(message.Text, "<blockquote") || strings.Contains(message.Text, "<code>") {
					t.Errorf("plain fallback still contains HTML: %s", message.Text)
				}
				if !strings.Contains(message.Text, "🔎 Technical details") ||
					!strings.Contains(message.Text, "storage unavailable") {
					t.Errorf("plain fallback lost safe technical details: %s", message.Text)
				}
				return botsfw.OnMessageSentResponse{}, nil
			}),
	)

	router := NewWebhookRouter(nil).(*webhooksRouter)
	if handled := router.processCommandResponseError(
		whc,
		&botsfw.Command{Code: "test"},
		responder,
		fmt.Errorf("storage unavailable"),
	); !handled {
		t.Fatal("processCommandResponseError() = false after successful plain fallback")
	}
}

func assertExpandableProviderDetails(t *testing.T, message botmsg.MessageFromBot, wants ...string) {
	t.Helper()
	if message.Format != botmsg.FormatHTML {
		t.Errorf("message.Format = %v, want FormatHTML", message.Format)
	}
	for _, want := range wants {
		if !strings.Contains(message.Text, want) {
			t.Errorf("message does not contain %q: %s", want, message.Text)
		}
	}
}

func newUserErrorDetailsWebhookContext(
	t *testing.T,
	locale i18n.Locale,
	token, requestID string,
	translatedOops ...string,
) botsfw.WebhookContext {
	t.Helper()
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	settings := &botsfw.BotSettings{
		Platform: botsfwconst.PlatformTelegram,
		Token:    token,
		UserErrorDetails: botsfw.UserErrorDetailsPolicy{
			Disclosure: botsfw.UserErrorDetailsDisclosureExpandable,
		},
	}
	whc.EXPECT().GetBotSettings().Return(settings).AnyTimes()
	whc.EXPECT().Locale().Return(locale).AnyTimes()
	request := &http.Request{Header: make(http.Header)}
	if requestID != "" {
		request.Header.Set("X-Request-ID", requestID)
	}
	whc.EXPECT().Request().Return(request).AnyTimes()
	if len(translatedOops) > 0 {
		whc.EXPECT().Translate(botsfw.MessageTextOopsSomethingWentWrong).Return(translatedOops[0])
	}
	whc.EXPECT().NewMessage(gomock.Any()).DoAndReturn(func(text string) botmsg.MessageFromBot {
		return botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: text}}
	})
	return whc
}

func newTelegramProviderError(t *testing.T, token string, statusCode int, description string) error {
	t.Helper()
	responseBody := fmt.Sprintf(
		`{"ok":false,"error_code":%d,"description":%q}`,
		statusCode,
		description,
	)
	bot := tgbotapi.NewBotAPIWithClient(token, &http.Client{
		Transport: telegramProviderRoundTripper(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    statusCode,
				ContentLength: int64(len(responseBody)),
				Body:          io.NopCloser(strings.NewReader(responseBody)),
			}, nil
		}),
	})
	_, err := bot.MakeRequest("sendRichMessage", nil)
	if err == nil {
		t.Fatal("MakeRequest() error = nil, want Telegram provider error")
	}
	return fmt.Errorf("sending preferences: %w", err)
}
