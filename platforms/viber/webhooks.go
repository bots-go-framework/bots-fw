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

func (h ViberWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	logger := h.BotHost.Logger(r)
	code := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	c := appengine.NewContext(r) //TODO: Remove dependency on AppEngine, should be passed indside.
	botSettings, ok := h.botsBy(c).Code[code]
	if !ok {
		errMess := fmt.Sprintf("Unknown public account: [%v]", code)
		err = bots.AuthFailedError(errMess)
		return
	}
	//viberinterface.CallbackBase{}.UnmarshalJSON()
	body, _ := ioutil.ReadAll(r.Body)
	if len(body) < 1024 * 3 {
		logger.Debugf(c, "Request body: %v", (string)(body))
	} else {
		logger.Debugf(c, "Request len(body): %v", len(body))
	}


	if match := reEvent.FindStringSubmatch(string(body)); len(match) > 0 {
		logger.Debugf(c, "Viber callback event: %v", match[1])
	}

	//var update viberbotapi.Update
	//err = json.Unmarshal(bytes, &update)
	//if err != nil {
	//	if ute, ok := err.(*json.UnmarshalTypeError); ok {
	//		logger.Errorf(c, "json.UnmarshalTypeError %v - %v - %v", ute.Value, ute.Type, ute.Offset)
	//	} else if se, ok := err.(*json.SyntaxError); ok {
	//		logger.Errorf(c, "json.SyntaxError: Offset=%v", se.Offset)
	//	} else {
	//		logger.Errorf(c, "json.Error: %T: %v", err, err.Error())
	//	}
	//	return
	//}
	botContext = bots.BotContext{
			BotHost:     h.BotHost,
			BotSettings: botSettings,
		}
		//, []bots.EntryInputs{
		//	//{
		//	//	Entry:  ViberWebhookEntry{update: update},
		//	//	Inputs: []bots.WebhookInput{NewViberWebhookInput(update)},
		//	//},
		//}
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
