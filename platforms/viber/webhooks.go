package viber_bot

import (
	"regexp"
	"fmt"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"google.golang.org/appengine"
	"net/http"
	"net/url"
	"io/ioutil"
	"strings"
	"crypto/hmac"
	"crypto/sha256"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-viber/viberinterface"
	"google.golang.org/appengine/log"
	"encoding/hex"
)

func NewViberWebhookHandler(botsBy bots.BotSettingsProvider, webhookDriver bots.WebhookDriver, botHost bots.BotHost, translatorProvider bots.TranslatorProvider) ViberWebhookHandler {
	if webhookDriver == nil {
		panic("webhookDriver == nil")
	}
	if botHost == nil {
		panic("botHost == nil")
	}
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return ViberWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:        ViberPlatform{},
			BotHost:            botHost,
			WebhookDriver:      webhookDriver,
			TranslatorProvider: translatorProvider,
		},
	}
}

type ViberWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.BotSettingsProvider
}

func (h ViberWebhookHandler) RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(pathPrefix + "/viber/callback/", h.HandleWebhookRequest)
	//http.HandleFunc(pathPrefix + "/viber/callback/", notFound)
	http.HandleFunc(pathPrefix + "/viber/setwebhook", h.SetWebhook)
}

func (h ViberWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleWebhook(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h ViberWebhookHandler) SetWebhook(w http.ResponseWriter, r *http.Request) {
	logger := h.Logger(r)
	client := h.GetHttpClient(r)
	botCode := r.URL.Query().Get("code")
	if botCode == "" {
		http.Error(w, "Missing required parameter: code", http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	botSettings, ok := h.botsBy(c).Code[botCode]
	if !ok {
		http.Error(w, fmt.Sprintf("Bot not found by code: %v", botCode), http.StatusBadRequest)
		return
	}
	bot := viberbotapi.NewViberBotApiWithHttpClient(botSettings.Token, client)
	//bot.Debug = true

	webhookUrl := fmt.Sprintf("https://%v/bot/viber/callback/%v", r.Host, url.QueryEscape(botSettings.Code))

	if _, err := bot.SetWebhook(webhookUrl, nil); err != nil {
		logger.Errorf(c, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Webhook set"))
	}
}

var reEvent = regexp.MustCompile(`"event"\s*:\s*"(\w+)"`)

func (h ViberWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	logger := h.BotHost.Logger(r)
	code := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	c := appengine.NewContext(r) //TODO: Remove dependency on AppEngine, should be passed indside.
	botSettings, ok := h.botsBy(c).Code[code]
	if !ok {
		errMess := fmt.Sprintf("Unknown public account: [%v]", code)
		err = bots.AuthFailedError(errMess)
		return
	}

	sig := r.URL.Query().Get("sig")
	var sigMAC []byte
	if sigMAC, err = hex.DecodeString(sig); err != nil {
		err = errors.Wrapf(err, "Failed to decode sig parameter using 'base64.RawURLEncoding'")
		return
	}

	//viberinterface.CallbackBase{}.UnmarshalJSON()
	body, _ := ioutil.ReadAll(r.Body)
	if len(body) < 1024 * 3 {
		logger.Debugf(c, "Request body: %v", (string)(body))
	} else {
		logger.Debugf(c, "Request len(body): %v", len(body))
	}

	mac := hmac.New(sha256.New, []byte(botSettings.Token))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(expectedMAC, sigMAC) {
		err = errors.New(fmt.Sprintf("Unexpected signature value:\n\tExpected: %v\n\tGot: %v",
			hex.EncodeToString(expectedMAC), sig))
		return
	}

	match := reEvent.FindStringSubmatch(string(body))
	if len(match) == 0 {
		err = errors.New("Unknown event type")
		return
	}

	event := match[1]

	switch event {
	case "message":
		textMessage := &viberinterface.TextMessage{}
		if err = textMessage.UnmarshalJSON(body); err != nil {
			err = errors.Wrap(err, "Failed to parse body to TextMessage")
			return
		}
		logger.Debugf(c, "TextMessage: %v", textMessage)
		//entriesWithInputs := append(entriesWithInputs, )
	case "webhook":
		setWebhookCallback := &viberinterface.SetWebhookResponse{}
		if err = setWebhookCallback.UnmarshalJSON(body); err != nil {
			err = errors.Wrap(err, "Failed to unmarshal request body to 'viberinterface.SetWebhookResponse'")
			return
		}
		logger.Infof(c, "Viber 'set-webhook' callback event")
		return
	default:
		log.Warningf(c, "Unknown callback event: [%v]", event)
	}
	botContext = &bots.BotContext{
			BotHost:     h.BotHost,
			BotSettings: botSettings,
		}
	return
}

func (h ViberWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) bots.WebhookContext {
	return NewViberWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (h ViberWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if twhc, ok := whc.(*ViberWebhookContext); ok {
		return NewViberWebhookResponder(w, twhc)
	} else {
		panic(fmt.Sprintf("Expected ViberWebhookContext, got: %T", whc))
	}
}

func (h ViberWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(ViberPlatformID, appContext, r)
}
