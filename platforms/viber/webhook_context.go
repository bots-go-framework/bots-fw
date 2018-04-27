package viber

import (
	"context"
	"fmt"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"net/http"
)

type viberWebhookContext struct {
	*bots.WebhookContextBase
	//update         viberbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
}

var _ bots.WebhookContext = (*viberWebhookContext)(nil)

func (whc *viberWebhookContext) NewEditMessage(text string, format bots.MessageFormat) (m bots.MessageFromBot, err error) {
	panic("Not supported by Viber")
}

func newViberWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement bots.GaQueuer) *viberWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		Platform{},
		botContext,
		webhookInput,
		botCoreStores,
		gaMeasurement,
		func() bool { return false },
		nil,
	)
	return &viberWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
	}
}

func (whc *viberWebhookContext) Close(c context.Context) error {
	return nil
}

func (whc *viberWebhookContext) Responder() bots.WebhookResponder {
	return whc.responder
}

type viberBotAPIUser struct {
	//user viberbotapi.User
}

func (tc viberBotAPIUser) FirstName() string {
	return ""
	//return tc.user.FirstName
}

func (tc viberBotAPIUser) LastName() string {
	return ""
	//return tc.user.LastName
}

//func (tc viberBotAPIUser) IdAsString() string {
//	return ""
//}

//func (tc viberBotAPIUser) IdAsInt64() int64 {
//	return int64(tc.user.ID)
//}

func (whc *viberWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc *viberWebhookContext) BotAPI() *viberbotapi.ViberBotAPI {
	return viberbotapi.NewViberBotAPIWithHTTPClient(whc.BotContext.BotSettings.Token, whc.BotContext.BotHost.GetHTTPClient(whc.Context()))
}

func (whc *viberWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	log.Warningf(whc.Context(), "IsNewerThen")
	//if viberChat, ok := whc.ChatEntity().(*ViberChat); ok && viberChat != nil {
	//	return whc.InputMessage().Sequence() > viberChat.LastProcessedUpdateID
	//}
	return true
}

func (whc *viberWebhookContext) NewChatEntity() bots.BotChat {
	return new(UserChatEntity)
}

func (whc *viberWebhookContext) getViberSenderID() string {
	senderID := whc.GetSender().GetID()
	if viberUserID, ok := senderID.(string); ok {
		return viberUserID
	}
	panic("string expected")
}

func (whc *viberWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	if _, ok := chatEntity.(*UserChatEntity); ok {
		//viberChat.LastProcessedUpdateID = tc.InputMessage().Sequence()
		return nil
	}
	return fmt.Errorf("Expected *ViberChat, got: %T", chatEntity)
}
