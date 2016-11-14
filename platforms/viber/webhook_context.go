package viber_bot

import (
	"errors"
	"fmt"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/bots-framework/core"
	//"google.golang.org/appengine/log"
	"github.com/strongo/measurement-protocol"
	"net/http"
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

func (tc ViberWebhookContext) Close() error {
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

func (whc *ViberWebhookContext) AppUserIntID() (appUserIntID int64) {
	if chatEntity := whc.ChatEntity(); chatEntity != nil {
		appUserIntID = chatEntity.GetAppUserIntID()
	}
	if appUserIntID == 0 {
		botUser, err := whc.GetOrCreateBotUserEntityBase()
		if err != nil {
			panic(fmt.Sprintf("Failed to get bot user entity: %v", err))
		}
		appUserIntID = botUser.GetAppUserIntID()
	}
	return
}

func (whc *ViberWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := whc.AppUserIntID()
	appUser := whc.BotAppContext().NewBotAppUserEntity()
	err := whc.BotAppUserStore.GetAppUserByID(appUserID, appUser)
	return appUser, err
}

func (whc *ViberWebhookContext) BotChatID() interface{} {
	id := whc.BotChatIntID()
	if id == 0 {
		return nil
	}
	return id
}

func (whc *ViberWebhookContext) BotChatIntID() (chatId int64) {
	panic("Not implemented")
}

func (whc *ViberWebhookContext) ChatEntity() bots.BotChat {
	if whc.BotChatID() == nil {
		return nil
	}
	botChatEntity, err := whc.WebhookContextBase.ChatEntity(whc)
	if err != nil {
		panic(fmt.Sprintf("Failed to get BotChat entity: %v", err))
	}
	return botChatEntity
}

func (whc *ViberWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	whc.Logger().Warningf(whc.Context(), "IsNewerThen")
	//if viberChat, ok := whc.ChatEntity().(*ViberChat); ok && viberChat != nil {
	//	return whc.InputMessage().Sequence() > viberChat.LastProcessedUpdateID
	//}
	return true
}

func (whc *ViberWebhookContext) NewChatEntity() bots.BotChat {
	return new(ViberChat)
}

func (whc *ViberWebhookContext) getViberSenderID() string {
	senderID := whc.GetSender().GetID()
	if viberUserID, ok := senderID.(string); ok {
		return viberUserID
	}
	panic("string expected")
}

func (whc *ViberWebhookContext) MakeChatEntity() bots.BotChat {
	viberChat := whc.Chat()
	chatEntity := ViberChat{
		BotChatEntity: bots.BotChatEntity{
			Type:  viberChat.GetType(),
			Title: viberChat.GetFullName(),
		},
		ViberUserID: whc.getViberSenderID(),
	}
	return &chatEntity
}

func (tc *ViberWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	if _, ok := chatEntity.(*ViberChat); ok {
		//viberChat.LastProcessedUpdateID = tc.InputMessage().Sequence()
		return nil
	}
	return errors.New(fmt.Sprintf("Expected *ViberChat, got: %T", chatEntity))
}