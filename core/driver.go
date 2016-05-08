package bots

import (
	"net/http"
	"fmt"
	"runtime/debug"
)

// The driver is doing initial request & final response processing
// That includes logging, creating input messages in a general format, sending response
type WebhookDriver interface {
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}

type BotDriver struct {
	botHost BotHost
	router  *WebhooksRouter
}

var _ WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

func NewBotDriver(host BotHost, router *WebhooksRouter) WebhookDriver {
	return BotDriver{botHost: host, router: router}
}

func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler) {
	log := d.botHost.GetLogger(r)
	log.Infof("HandleWebhook() => webhookHandler: %T", webhookHandler)

	botContext, entriesWithInputs, err := webhookHandler.GetBotContextAndInputs(r)

	if err != nil {
		if _, ok := err.(AuthFailedError); ok {
			log.Warningf("Auth failed: %v", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			log.Errorf("Failed to call webhookHandler.GetBotContext(r): %v", err)
			//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	log.Infof("Got %v entries", len(entriesWithInputs))

	defer func() {
		if recovered := recover(); recovered != nil {
			messageText := fmt.Sprintf("Server error (panic): %v", recovered)
			log.Criticalf("Panic recovered: %s\n%s", messageText, debug.Stack())
			//whc.ReplyByBot(whc.NewMessage("\xF0\x9F\x9A\xA8 " + messageText))
		}
	}()

	if err != nil {
		log.Errorf("Failed to create new WebhookContext: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for i, entryWithInputs := range entriesWithInputs {
		log.Infof("Entry[%v]: %v, %v inputs", i, entryWithInputs.Entry.GetID(), len(entryWithInputs.Inputs))
		for j, input := range entryWithInputs.Inputs {
			switch input.InputType() {
			case WebhookInputMessage:
				log.Infof("Input[%v].Message().Text(): %v", j, input.InputMessage().Text())
			default:
				log.Infof("Input[%v].InputType(): %v", j, input.InputType())
			}

			whc := webhookHandler.CreateWebhookContext(r, botContext, input, webhookHandler.GetTranslator(r))
			d.router.Dispatch(whc)
		}
	}
}
