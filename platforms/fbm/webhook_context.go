package fbm_bot

import (
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"context"
	"net/http"
)

type FbmWebhookContext struct {
	*bots.WebhookContextBase
	//update         fbm_api.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
}

var _ bots.WebhookContext = (*FbmWebhookContext)(nil)

func NewFbmWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement bots.GaQueuer) *FbmWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		FbmPlatform{},
		botContext,
		webhookInput,
		botCoreStores,
		gaMeasurement,
		func() bool { return false },
		nil,
	)
	return &FbmWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
	}
}

func (whc *FbmWebhookContext) NewEditMessage(text string, format bots.MessageFormat) (bots.MessageFromBot, error) {
	panic("not implemented")
}

//func (whc *FbmWebhookContext) NewEditMessageKeyboard(kbMarkup fbm_api.InlineKeyboardMarkup) bots.MessageFromBot {
//	//chatID, _ := whc.BotChatID().(int64)
//	//messageID := whc.InputCallbackQuery().GetMessage().IntID()
//	//editMessageMarkupConfig := fbm_api.NewEditMessageReplyMarkup(chatID, (int)(messageID), kbMarkup)
//	//m := whc.NewMessage("")
//	//m.FbmEditMessageMarkup = &editMessageMarkupConfig
//	//return m
//	panic("not implemented")
//}

func (whc FbmWebhookContext) Close(c context.Context) error {
	return nil
}

func (whc FbmWebhookContext) Responder() bots.WebhookResponder {
	return whc.responder
}

type FbmBotApiUser struct {
	user fbm_api.Sender
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

//func (whc *FbmWebhookContext) BotApi() *fbm_api.BotAPI {
//	return fbm_api.NewBotAPIWithClient(whc.BotContext.BotSettings.Token, whc.GetHttpClient())
//}

func (whc *FbmWebhookContext) BotChatIntID() (chatId int64) {
	//webhookInput := whc.input
	//switch webhookInput.InputType() {
	//case bots.WebhookInputMessage:
	//	chatId = webhookInput.Chat().GetID().(int64)
	//case bots.WebhookInputCallbackQuery:
	//	callbackQuery := webhookInput.InputCallbackQuery()
	//	if callbackQuery == nil {
	//		return 0
	//	}
	//	chat := callbackQuery.Chat()
	//	if chat != nil {
	//		chatId = chat.GetID().(int64)
	//	} else {
	//		data := callbackQuery.GetData()
	//		if strings.Contains(data, "chat=") {
	//			c := whc.Context()
	//			values, err := url.ParseQuery(data)
	//			if err != nil {
	//				log.Errorf(c, "Failed to GetData() from webhookInput.InputCallbackQuery()")
	//				return 0
	//			}
	//			chatIdAsStr := values.Get("chat")
	//			if chatId, err = strconv.ParseInt(chatIdAsStr, 10, 64); err != nil {
	//				log.Errorf(c, "Failed to parse 'chat' parameter to int: %v", err)
	//				return 0
	//			}
	//		}
	//	}
	//}
	//
	//return chatId
	panic("Not implemented")
}

func (whc *FbmWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	panic("Not implemented")
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

func (whc *FbmWebhookContext) NewFbmMessage(text string) fbm_api.SendMessage {
	////inputMessage := whc.InputMessage()
	////if inputMessage != nil {
	////ctx := whc.Context()
	////chat := inputMessage.Chat()
	////chatID := chat.GetID()
	////log.Infof(ctx, "NewTgMessage(): whc.update.Message.Chat.ID: %v", chatID)
	//botChatID := whc.BotChatID()
	//if botChatID == nil {
	//	panic(fmt.Sprintf("Not able to send message as BotChatID() returned nil. text: %v", text))
	//}
	//if int64ID, ok := botChatID.(int64); ok {
	//	return fbm_api.NewMessage(int64ID, text)
	//} else {
	//	if intID, ok := botChatID.(int); ok {
	//		return fbm_api.NewMessage(int64(intID), text)
	//	} else {
	//		panic(fmt.Sprintf("OK=%v;Expected int or int64, got: %T", ok, botChatID))
	//	}
	//}
	////}
	////panic(fmt.Sprintf("Expected to be called just for inputType == Message, got: %v", whc.InputType()))
	panic("Not implemented")
}

func (whc *FbmWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	panic("Not implemented")
	//if chat, ok := chatEntity.(*FbmChat); ok {
	//	chat.LastSeq = whc.InputMessage().Sequence()
	//	return nil
	//}
	//return fmt.Errorf("Expected *FbmChat, got: %T", chatEntity)
}
