package viber

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// NewViberWebhookHandler creates new Viber webhook handler
func NewViberWebhookHandler(botsBy bots.SettingsProvider, translatorProvider bots.TranslatorProvider) bots.WebhookHandler {
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return viberWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:        Platform{},
			TranslatorProvider: translatorProvider,
		},
	}
}

type viberWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.SettingsProvider
}

var _ bots.WebhookHandler = (*viberWebhookHandler)(nil)

func (h viberWebhookHandler) HandleUnmatched(whc bots.WebhookContext) (m bots.MessageFromBot) {
	return
}

func (h viberWebhookHandler) RegisterHttpHandlers(driver bots.WebhookDriver, host bots.BotHost, router *httprouter.Router, pathPrefix string) {
	if router == nil {
		panic("router == nil")
	}
	h.BaseHandler.Register(driver, host)
	router.POST(pathPrefix+"/viber/callback/", h.HandleWebhookRequest)
	router.GET(pathPrefix+"/viber/set-webhook", h.SetWebhook)
}

func (h viberWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	switch r.Method {
	case http.MethodPost:
		h.HandleWebhook(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

var reEvent = regexp.MustCompile(`"event"\s*:\s*"(\w+)"`)

func (h viberWebhookHandler) GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	code := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	botSettings, ok := h.botsBy(c).ByCode[code]
	if !ok {
		errMess := fmt.Sprintf("Unknown public account: [%v]", code)
		err = bots.ErrAuthFailed(errMess)
		return
	}

	sig := r.URL.Query().Get("sig")
	var sigMAC []byte
	if sigMAC, err = hex.DecodeString(sig); err != nil {
		err = errors.Wrapf(err, "Failed to decode sig parameter using 'base64.RawURLEncoding'")
		return
	}

	//viber interface.CallbackBase{}.UnmarshalJSON()
	body, _ := io.ReadAll(r.Body)
	if len(body) < 1024*3 {
		log.Debugf(c, "Request body: %v", (string)(body))
	} else {
		log.Debugf(c, "Request len(body): %v", len(body))
	}

	mac := hmac.New(sha256.New, []byte(botSettings.Token))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(expectedMAC, sigMAC) {
		err = fmt.Errorf("Unexpected signature value:\n\tExpected: %v\n\tGot: %v",
			hex.EncodeToString(expectedMAC), sig)
		return
	}

	match := reEvent.FindStringSubmatch(string(body))
	if len(match) == 0 {
		err = errors.New("Unknown event type")
		return
	}

	event := match[1]

	unmarshal := func(m interface {
		UnmarshalJSON(input []byte) error
	}) (err error) {
		if err = m.UnmarshalJSON(body); err != nil {
			err = errors.Wrapf(err, "Failed to unmarshal request body to %T", m)
		}
		log.Debugf(c, "%T: %v", m, m)
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
				Entry:  WebhookEntry{},
				Inputs: []bots.WebhookInput{newViberWebhookTextMessage(message)},
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
				Entry:  WebhookEntry{},
				Inputs: []bots.WebhookInput{newViberWebhookInputConversationStarted(message)},
			},
		}
	case "webhook":
		message := &viberinterface.SetWebhookResponse{}
		if err = unmarshal(message); err != nil {
			return
		}
		log.Infof(c, "Viber 'set-webhook' callback event")
		return // Do not create bot context
	default:
		log.Warningf(c, "Unknown callback event: [%v]", event)
		return
	}
	botContext = bots.NewBotContext(h.BotHost, botSettings)
	return
}

func (viberWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement bots.GaQueuer) bots.WebhookContext {
	return newViberWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (viberWebhookHandler) GetResponder(_ http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if viberWhc, ok := whc.(*viberWebhookContext); ok {
		return newViberWebhookResponder(viberWhc)
	}
	panic(fmt.Sprintf("Expected viberWebhookContext, got: %T", whc))
}

func (h viberWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(PlatformID, appContext, r)
}
