package viber_bot

import (
	"regexp"
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"google.golang.org/appengine"
	"net/http"
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
var _ bots.WebhookHandler = (*ViberWebhookHandler)(nil)


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

var reEvent = regexp.MustCompile(`"event"\s*:\s*"(\w+)"`)

func (h ViberWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	logger := h.BotHost.Logger(r)
	code := r.URL.Path[strings.LastIndex(r.URL.Path, "/") + 1:]
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

	unmarshal := func(m interface{ UnmarshalJSON(input []byte) error }) (err error) {
		if err = m.UnmarshalJSON(body); err != nil {
			err = errors.Wrapf(err, "Failed to unmarshal request body to %T", m)
		}
		logger.Debugf(c, "%T: %v", m, m)
		return
	}

	switch event {
	case "message":
		message := viberinterface.CallbackOnMessage{}
		if err = unmarshal(&message); err != nil {
			return
		}
		entriesWithInputs = []bots.EntryInputs{
			{
				Entry: ViberWebhookEntry{},
				Inputs: []bots.WebhookInput{NewViberWebhookTextMessage(message)},
			},
		}
	//entriesWithInputs := append(entriesWithInputs, )
	case "seen":
		message := &viberinterface.CallbackOnDelivered{}
		if err = unmarshal(message); err != nil {
			return
		}
	case "delievered":
		message := &viberinterface.CallbackOnDelivered{}
		if err = unmarshal(message); err != nil {
			return
		}
	case "failed":
		message := &viberinterface.CallbackOnFailed{}
		if err = unmarshal(message); err != nil {
			return
		}
	case "subscribed":
		message := &viberinterface.CallbackOnSubscribed{}
		if err = unmarshal(message); err != nil {
			return
		}
	case "unsubscribed":
		message := &viberinterface.CallbackOnUnsubscribed{}
		if err = unmarshal(message); err != nil {
			return
		}
	case "conversation_started":
		message := viberinterface.CallbackOnConversationStarted{}
		if err = unmarshal(&message); err != nil {
			return
		}
		entriesWithInputs = []bots.EntryInputs{
			{
				Entry: ViberWebhookEntry{},
				Inputs: []bots.WebhookInput{NewViberWebhookInputConversationStarted(message)},
			},
		}
	case "webhook":
		message := &viberinterface.SetWebhookResponse{}
		if err = unmarshal(message); err != nil {
			return
		}
		logger.Infof(c, "Viber 'set-webhook' callback event")
		return // Do not create bot context
	default:
		log.Warningf(c, "Unknown callback event: [%v]", event)
		return
	}
	botContext = bots.NewBotContext(h.BotHost, botSettings)
	return
}

func (_ ViberWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) bots.WebhookContext {
	return NewViberWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (_ ViberWebhookHandler) GetResponder(_ http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if viberWhc, ok := whc.(*ViberWebhookContext); ok {
		return NewViberWebhookResponder(viberWhc)
	} else {
		panic(fmt.Sprintf("Expected ViberWebhookContext, got: %T", whc))
	}
}

func (handler ViberWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return handler.BotHost.GetBotCoreStores(ViberPlatformID, appContext, r)
}
