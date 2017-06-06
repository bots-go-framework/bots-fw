package telegram_bot

import (
	"encoding/json"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"io/ioutil"
	"net/http"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"github.com/strongo/app/log"
	//"github.com/kylelemons/go-gypsy/yaml"
	//"bytes"
	"strings"
	"time"
	"github.com/pquerna/ffjson/ffjson"
)

func NewTelegramWebhookHandler(botsBy bots.SettingsProvider, webhookDriver bots.WebhookDriver, botHost bots.BotHost, translatorProvider bots.TranslatorProvider) TelegramWebhookHandler {
	if webhookDriver == nil {
		panic("webhookDriver == nil")
	}
	if botHost == nil {
		panic("botHost == nil")
	}
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return TelegramWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:        TelegramPlatform{},
			BotHost:            botHost,
			WebhookDriver:      webhookDriver,
			TranslatorProvider: translatorProvider,
		},
	}
}

type TelegramWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.SettingsProvider
}
var _ bots.WebhookHandler = (*TelegramWebhookHandler)(nil)

func (h TelegramWebhookHandler) RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(pathPrefix+"/telegram/webhook", h.HandleWebhookRequest)
	http.HandleFunc(pathPrefix+"/telegram/webhook/", notFound)
	http.HandleFunc(pathPrefix+"/telegram/setwebhook", func(w http.ResponseWriter, r *http.Request) {
		h.SetWebhook(h.Context(r), w, r)
	})
}

func (h TelegramWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	c := h.Context(r)
	//log.Debugf(c, "TelegramWebhookHandler.HandleWebhookRequest()")
	switch r.Method {
	case http.MethodPost:
		defer func() {
			if r := recover(); r != nil {
				log.Criticalf(c,"Unhandled exception in Telegram handler: %v", r)
			}
		}()
		h.HandleWebhook(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h TelegramWebhookHandler) SetWebhook(c context.Context, w http.ResponseWriter, r *http.Request) {
	ctxWithDeadline, _ := context.WithTimeout(c, 30*time.Second)
	client := h.GetHttpClient(ctxWithDeadline)
	botCode := r.URL.Query().Get("code")
	if botCode == "" {
		http.Error(w, "Missing required parameter: code", http.StatusBadRequest)
		return
	}
	botSettings, ok := h.botsBy(c).Code[botCode]
	if !ok {
		http.Error(w, fmt.Sprintf("Bot not found by code: %v", botCode), http.StatusBadRequest)
		return
	}
	bot := tgbotapi.NewBotAPIWithClient(botSettings.Token, client)
	bot.EnableDebug(c)
	//bot.Debug = true

	webhookUrl := fmt.Sprintf("https://%v/bot/telegram/webhook?token=%v", r.Host, bot.Token)

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(webhookUrl)); err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Webhook set"))
	}
}

func (h TelegramWebhookHandler) GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	token := r.URL.Query().Get("token")
	botSettings, ok := h.botsBy(c).ApiToken[token]
	if !ok {
		errMess := fmt.Sprintf("Unknown token: [%v]", token)
		err = bots.AuthFailedError(errMess)
		return
	}
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	if len(bodyBytes) < 1024 * 3 {
		s := string(bodyBytes)
		s = strings.Replace(s, `,"`, ",\n\"" , -1)
		s = strings.Replace(s, `:{`, `:{` + "\n", -1)
		log.Debugf(c, "Request body: %v", s)
		//if node, err := yaml.Parse(bytes.NewReader(bodyBytes)); err != nil {
		//	log.Debugf(c, "Request body: %v", (string)(bodyBytes))
		//} else {
		//	log.Debugf(c, "Request JSON body as YAML (%T):\n%v", node, yaml.Render(node))
		//}
	} else {
		log.Debugf(c, "Request len(body): %v", len(bodyBytes))
	}
	var update tgbotapi.Update
	err = ffjson.UnmarshalFast(bodyBytes, &update)
	if err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			log.Errorf(c, "json.UnmarshalTypeError %v - %v - %v", ute.Value, ute.Type, ute.Offset)
		} else if se, ok := err.(*json.SyntaxError); ok {
			log.Errorf(c, "json.SyntaxError: Offset=%v", se.Offset)
		} else {
			log.Errorf(c, "json.Error: %T: %v", err, err.Error())
		}
		return
	}
	botContext = bots.NewBotContext(h.BotHost, botSettings)
	input := NewTelegramWebhookInput(update)
	if input == nil {
		err = errors.New("Unexpected whi")
		return
	}
	entriesWithInputs = []bots.EntryInputs{
			{
				Entry:  TelegramWebhookEntry{update: update},
				Inputs: []bots.WebhookInput{input},
			},
		}
	return
}

func (h TelegramWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) bots.WebhookContext {
	return NewTelegramWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (h TelegramWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if twhc, ok := whc.(*TelegramWebhookContext); ok {
		return NewTelegramWebhookResponder(w, twhc)
	} else {
		panic(fmt.Sprintf("Expected TelegramWebhookContext, got: %T", whc))
	}
}

func (h TelegramWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(TelegramPlatformID, appContext, r)
}
