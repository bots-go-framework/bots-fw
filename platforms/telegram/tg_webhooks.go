package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"io"
	"net/http"
	//"github.com/kylelemons/go-gypsy/yaml"
	//"bytes"
	"bytes"
	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"
	"strings"
	"time"
)

// NewTelegramWebhookHandler creates new Telegram webhooks handler
//
//goland:noinspection GoUnusedExportedFunction
func NewTelegramWebhookHandler(botsBy bots.SettingsProvider, translatorProvider bots.TranslatorProvider) bots.WebhookHandler {
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return tgWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:        Platform{},
			TranslatorProvider: translatorProvider,
		},
	}
}

type tgWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.SettingsProvider
}

var _ bots.WebhookHandler = (*tgWebhookHandler)(nil)

func (h tgWebhookHandler) HandleUnmatched(whc bots.WebhookContext) (m bots.MessageFromBot) {
	switch whc.InputType() {
	case bots.WebhookInputCallbackQuery:
		m.BotMessage = CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
			Text:      "⚠️ Error: Not matched to any command",
			ShowAlert: true,
		})
	}
	return
}

func (h tgWebhookHandler) RegisterHttpHandlers(driver bots.WebhookDriver, host bots.BotHost, router *httprouter.Router, pathPrefix string) {
	if router == nil {
		panic("router == nil")
	}
	h.BaseHandler.Register(driver, host)

	pathPrefix = strings.TrimSuffix(pathPrefix, "/")
	//router.POST(pathPrefix+"/telegram/webhook", h.HandleWebhookRequest) // TODO: Remove obsolete
	router.POST(pathPrefix+"/tg/hook", h.HandleWebhookRequest)
	router.GET(pathPrefix+"/tg/set-webhook", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.SetWebhook(h.Context(r), w, r)
	})
	router.GET(pathPrefix+"/tg/test", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Debugf(h.Context(r), "Test request")
		if _, err := w.Write([]byte("Test response")); err != nil {
			log.Errorf(r.Context(), "Failed to write test response: %v", err)
		}
	})
}

func (h tgWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Criticalf(h.Context(r), "Unhandled panic in Telegram handler: %v", err)
		}
	}()

	h.HandleWebhook(w, r, h)
}

func (h tgWebhookHandler) SetWebhook(c context.Context, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "tgWebhookHandler.SetWebhook()")
	ctxWithDeadline, cancel := context.WithTimeout(c, 30*time.Second)
	defer cancel()
	client := h.GetHTTPClient(ctxWithDeadline)
	botCode := r.URL.Query().Get("code")
	if botCode == "" {
		http.Error(w, "tgWebhookHandler: Missing required parameter: code", http.StatusBadRequest)
		return
	}
	botSettings, ok := h.botsBy(c).ByCode[botCode]
	if !ok {
		m := fmt.Sprintf("Bot not found by code: %v", botCode)
		http.Error(w, m, http.StatusBadRequest)
		log.Errorf(c, fmt.Sprintf("%v. All bots: %v", m, h.botsBy(c).ByCode))
		return
	}
	bot := tgbotapi.NewBotAPIWithClient(botSettings.Token, client)
	bot.EnableDebug(c)
	//bot.Debug = true

	webhookURL := fmt.Sprintf("https://%v/bot/tg/hook?id=%v&token=%v", r.Host, botCode, bot.Token)

	webhookConfig := tgbotapi.NewWebhook(webhookURL)
	webhookConfig.AllowedUpdates = []string{
		"message",
		"edited_message",
		"inline_query",
		"chosen_inline_result",
		"callback_query",
	}
	if response, err := bot.SetWebhook(webhookConfig); err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Errorf(c, "Failed to write error to response: %v", err)
		}
	} else {
		if _, err := w.Write([]byte(fmt.Sprintf("Webhook set\nErrorCode: %d\nDescription: %v\nContent: %v", response.ErrorCode, response.Description, string(response.Result)))); err != nil {
			log.Errorf(c, "Failed to write error to response: %v", err)
		}
	}
}

func (h tgWebhookHandler) GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	//log.Debugf(c, "tgWebhookHandler.GetBotContextAndInputs()")
	token := r.URL.Query().Get("token")
	botSettings, ok := h.botsBy(c).ByAPIToken[token]
	if !ok {
		errMess := fmt.Sprintf("Unknown token: [%v]", token)
		err = bots.ErrAuthFailed(errMess)
		return
	}
	botContext = bots.NewBotContext(h.BotHost, botSettings)
	var bodyBytes []byte
	defer func() {
		if r.Body != nil {
			if err := r.Body.Close(); err != nil {
				log.Errorf(c, "Failed to close request body: %v", err)
			}
		}
	}()
	if bodyBytes, err = io.ReadAll(r.Body); err != nil {
		err = errors.Wrap(err, "Failed to read request body")
		return
	}

	var requestLogged bool
	logRequestBody := func() {
		if !requestLogged {
			requestLogged = true
			if len(bodyBytes) < 1024*10 {
				var bodyToLog bytes.Buffer
				var bodyStr string
				if indentErr := json.Indent(&bodyToLog, bodyBytes, "", "\t"); indentErr == nil {
					bodyStr = bodyToLog.String()
				} else {
					bodyStr = string(bodyBytes)
				}
				log.Debugf(c, "Request body: %v", bodyStr)
			} else {
				log.Debugf(c, "Request len(body): %v", len(bodyBytes))
			}
		}
	}

	var update *tgbotapi.Update
	if update, err = h.unmarshalUpdate(c, bodyBytes); err != nil {
		logRequestBody()
		return
	}

	var input bots.WebhookInput
	if input, err = NewTelegramWebhookInput(update, logRequestBody); err != nil {
		logRequestBody()
		return
	}
	logRequestBody()

	entriesWithInputs = []bots.EntryInputs{
		{
			Entry:  tgWebhookEntry{update: update},
			Inputs: []bots.WebhookInput{input},
		},
	}

	if input == nil {
		logRequestBody()
		err = errors.WithMessage(bots.ErrNotImplemented, "Telegram input is <nil>")
		return
	}
	log.Debugf(c, "Telegram input type: %T", input)
	return
}

func (h tgWebhookHandler) unmarshalUpdate(_ context.Context, content []byte) (update *tgbotapi.Update, err error) {
	update = new(tgbotapi.Update)
	if err = ffjson.UnmarshalFast(content, update); err != nil {
		return
	}
	return
}

func (h tgWebhookHandler) CreateWebhookContext(
	appContext bots.BotAppContext,
	r *http.Request, botContext bots.BotContext,
	webhookInput bots.WebhookInput,
	botCoreStores bots.BotCoreStores,
	gaMeasurement bots.GaQueuer,
) bots.WebhookContext {
	return newTelegramWebhookContext(
		appContext, r, botContext, webhookInput.(TgWebhookInput), botCoreStores, gaMeasurement)
}

func (h tgWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if twhc, ok := whc.(*tgWebhookContext); ok {
		return newTgWebhookResponder(w, twhc)
	}
	panic(fmt.Sprintf("Expected tgWebhookContext, got: %T", whc))
}

func (h tgWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(PlatformID, appContext, r)
}
