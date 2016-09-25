package fbm_strongo_bot

import (
	"github.com/strongo/bots-framework/core"
	"net/http"
	"github.com/strongo/measurement-protocol"
	"github.com/strongo/bots-api-fbm"
	"fmt"
	"strings"
	"net/url"
	"strconv"
	"github.com/pkg/errors"
)

type FbmWebhookContext struct {
	*bots.WebhookContextBase
	//update         fbm_bot_api.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
}

var _ bots.WebhookContext = (*FbmWebhookContext)(nil)

func NewFbmWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) *FbmWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		FbmPlatform{},
		botContext,
		webhookInput,
		botCoreStores,
		gaMeasurement,
	)
	return &FbmWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
	}
}

func (whc *FbmWebhookContext) NewEditCallbackMessage(messageText string) bots.MessageFromBot {
	//chatID, _ := whc.BotChatID().(int64)
	//messageID := whc.InputCallbackQuery().GetMessage().IntID()
	//editMessageTextConfig := fbm_bot_api.NewEditMessageText(chatID, (int)(messageID), messageText)
	//editMessageTextConfig.ParseMode = "HTML"
	//m := whc.NewMessage("")
	//m.FbmEditMessageText = editMessageTextConfig
	//return m
	panic("not implemented")
}

//func (whc *FbmWebhookContext) NewEditCallbackMessageKeyboard(kbMarkup fbm_bot_api.InlineKeyboardMarkup) bots.MessageFromBot {
//	//chatID, _ := whc.BotChatID().(int64)
//	//messageID := whc.InputCallbackQuery().GetMessage().IntID()
//	//editMessageMarkupConfig := fbm_bot_api.NewEditMessageReplyMarkup(chatID, (int)(messageID), kbMarkup)
//	//m := whc.NewMessage("")
//	//m.FbmEditMessageMarkup = &editMessageMarkupConfig
//	//return m
//	panic("not implemented")
//}


func (tc FbmWebhookContext) Close() error {
	return nil
}

func (tc FbmWebhookContext) Responder() bots.WebhookResponder {
	return tc.responder
}

type FbmBotApiUser struct {
	user fbm_bot_api.Sender
}

func (tc FbmBotApiUser) FirstName() string {
	return tc.user.ID //tc.user.FirstName
}

func (tc FbmBotApiUser) LastName() string {
	return tc.user.ID
}

//func (tc FbmBotApiUser) IdAsString() string {
//	return ""
//}

//func (tc FbmBotApiUser) IdAsInt64() int64 {
//	return int64(tc.user.ID)
//}

func (whc *FbmWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

//func (whc *FbmWebhookContext) BotApi() *fbm_bot_api.BotAPI {
//	return fbm_bot_api.NewBotAPIWithClient(whc.BotContext.BotSettings.Token, whc.GetHttpClient())
//}

func (whc *FbmWebhookContext) AppUserIntID() (appUserIntID int64) { // TODO: This method is duplicating telegram
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

func (whc *FbmWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := whc.AppUserIntID()
	appUser := whc.BotAppContext().NewBotAppUserEntity()
	err := whc.BotAppUserStore.GetAppUserByID(appUserID, appUser)
	return appUser, err
}

func (whc *FbmWebhookContext) MessageText() string {
	inputMessage := whc.WebhookInput.InputMessage()
	if inputMessage != nil {
		return inputMessage.Text()
	}
	return ""
}

func (whc *FbmWebhookContext) BotChatID() interface{} {
	return whc.WebhookInput.Chat().GetID()
}

func (whc *FbmWebhookContext) BotChatIntID() (chatId int64) {
	webhookInput := whc.WebhookInput
	switch webhookInput.InputType() {
	case bots.WebhookInputMessage:
		chatId = webhookInput.Chat().GetID().(int64)
	case bots.WebhookInputCallbackQuery:
		callbackQuery := webhookInput.InputCallbackQuery()
		if callbackQuery == nil {
			return 0
		}
		chat := callbackQuery.Chat()
		if chat != nil {
			chatId = chat.GetID().(int64)
		} else {
			data := callbackQuery.GetData()
			if strings.Contains(data, "chat=") {
				c := whc.Context()
				values, err := url.ParseQuery(data)
				if err != nil {
					whc.Logger().Errorf(c, "Failed to GetData() from webhookInput.InputCallbackQuery()")
					return 0
				}
				chatIdAsStr := values.Get("chat")
				if chatId, err = strconv.ParseInt(chatIdAsStr, 10, 64); err != nil {
					whc.Logger().Errorf(c, "Failed to parse 'chat' parameter to int: %v", err)
					return 0
				}
			}
		}
	}

	return chatId
}

func (whc *FbmWebhookContext) ChatEntity() bots.BotChat {
	if whc.BotChatID() == nil {
		return nil
	}
	botChatEntity, err := whc.WebhookContextBase.ChatEntity(whc)
	if err != nil {
		panic(fmt.Sprintf("Failed to get BotChat entity: %v", err))
	}
	return botChatEntity
}

func (whc *FbmWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	if chat, ok := whc.ChatEntity().(*FbmChat); ok && chat != nil {
		return whc.InputMessage().Sequence() > chat.LastSeq
	}
	return false
}

func (whc *FbmWebhookContext) NewChatEntity() bots.BotChat {
	return new(FbmChat)
}

func (whc *FbmWebhookContext) getFbmSenderID() string {
	senderID := whc.GetSender().GetID()
	if fbmUserID, ok := senderID.(string); ok {
		return fbmUserID
	}
	panic("string expected")
}

func (whc *FbmWebhookContext) MakeChatEntity() bots.BotChat {
	fbmChat := whc.Chat()
	chatEntity := FbmChat{
		BotChatEntity: bots.BotChatEntity{
			Type:  fbmChat.GetType(),
			Title: fbmChat.GetFullName(),
		},
		FbmUserID: whc.getFbmSenderID(),
	}
	return &chatEntity
}

func (tc *FbmWebhookContext) NewFbmMessage(text string) fbm_bot_api.SendMessage {
	////inputMessage := tc.InputMessage()
	////if inputMessage != nil {
	////ctx := tc.Context()
	////chat := inputMessage.Chat()
	////chatID := chat.GetID()
	////log.Infof(ctx, "NewTgMessage(): tc.update.Message.Chat.ID: %v", chatID)
	//botChatID := tc.BotChatID()
	//if botChatID == nil {
	//	panic(fmt.Sprintf("Not able to send message as BotChatID() returned nil. text: %v", text))
	//}
	//if int64ID, ok := botChatID.(int64); ok {
	//	return fbm_bot_api.NewMessage(int64ID, text)
	//} else {
	//	if intID, ok := botChatID.(int); ok {
	//		return fbm_bot_api.NewMessage(int64(intID), text)
	//	} else {
	//		panic(fmt.Sprintf("OK=%v;Expected int or int64, got: %T", ok, botChatID))
	//	}
	//}
	////}
	////panic(fmt.Sprintf("Expected to be called just for inputType == Message, got: %v", tc.InputType()))
	panic("Not implemented")
}

func (tc *FbmWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	if chat, ok := chatEntity.(*FbmChat); ok {
		chat.LastSeq = tc.InputMessage().Sequence()
		return nil
	}
	return errors.New(fmt.Sprintf("Expected *FbmChat, got: %T", chatEntity))
}
