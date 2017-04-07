package viber_bot

import (
	"errors"
	"fmt"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/bots-framework/core"
	//"github.com/strongo/app/log"
	"github.com/strongo/measurement-protocol"
	"net/http"
	"golang.org/x/net/context"
	"github.com/strongo/app/log"
)

type ViberWebhookContext struct {
	*bots.WebhookContextBase
	//update         viberbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
}

var _ bots.WebhookContext = (*ViberWebhookContext)(nil)

func (whc *ViberWebhookContext) NewEditCallbackMessage(messageText string) bots.MessageFromBot {
	panic("Not supported by Viber")
}

func NewViberWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) *ViberWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		ViberPlatform{},
		botContext,
		webhookInput,
		botCoreStores,
		gaMeasurement,
	)
	return &ViberWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
	}
}

func (tc ViberWebhookContext) Close(c context.Context) error {
	return nil
}

func (tc ViberWebhookContext) Responder() bots.WebhookResponder {
	return tc.responder
}

type ViberBotApiUser struct {
	//user viberbotapi.User
}

func (tc ViberBotApiUser) FirstName() string {
	return ""
	//return tc.user.FirstName
}

func (tc ViberBotApiUser) LastName() string {
	return ""
	//return tc.user.LastName
}

//func (tc ViberBotApiUser) IdAsString() string {
//	return ""
//}

//func (tc ViberBotApiUser) IdAsInt64() int64 {
//	return int64(tc.user.ID)
//}

func (whc *ViberWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc *ViberWebhookContext) BotApi() *viberbotapi.ViberBotApi {
	return viberbotapi.NewViberBotApiWithHttpClient(whc.BotContext.BotSettings.Token, whc.GetHttpClient())
}

func (whc *ViberWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	log.Warningf(whc.Context(), "IsNewerThen")
	//if viberChat, ok := whc.ChatEntity().(*ViberChat); ok && viberChat != nil {
	//	return whc.InputMessage().Sequence() > viberChat.LastProcessedUpdateID
	//}
	return true
}

func (whc *ViberWebhookContext) NewChatEntity() bots.BotChat {
	return new(ViberUserChatEntity)
}

func (whc *ViberWebhookContext) getViberSenderID() string {
	senderID := whc.GetSender().GetID()
	if viberUserID, ok := senderID.(string); ok {
		return viberUserID
	}
	panic("string expected")
}

func (tc *ViberWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	if _, ok := chatEntity.(*ViberUserChatEntity); ok {
		//viberChat.LastProcessedUpdateID = tc.InputMessage().Sequence()
		return nil
	}
	return errors.New(fmt.Sprintf("Expected *ViberChat, got: %T", chatEntity))
}