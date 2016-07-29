package telegram_bot

import (
	"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	//"google.golang.org/appengine/log"
	"net/http"
	"strings"
	"net/url"
	"strconv"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder bots.WebhookResponder
}

var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func NewTelegramWebhookContext(appContext bots.AppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores) *TelegramWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		TelegramPlatform{},
		botContext,
		webhookInput,
		botCoreStores,
	)
	return &TelegramWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
	}
}

func (tc TelegramWebhookContext) Close() error {
	return nil
}

func (tc TelegramWebhookContext) Responder() bots.WebhookResponder {
	return tc.responder
}

type TelegramBotApiUser struct {
	user tgbotapi.User
}

func (tc TelegramBotApiUser) FirstName() string {
	return tc.user.FirstName
}

func (tc TelegramBotApiUser) LastName() string {
	return tc.user.LastName
}

//func (tc TelegramBotApiUser) IdAsString() string {
//	return ""
//}

func (tc TelegramBotApiUser) IdAsInt64() int64 {
	return int64(tc.user.ID)
}

func (whc *TelegramWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc *TelegramWebhookContext) BotApi() *tgbotapi.BotAPI {
	botApi, err := tgbotapi.NewBotAPIWithClient(whc.BotContext.BotSettings.Token, whc.GetHttpClient())
	if err != nil {
		panic(err)
	}
	return botApi
}

func (whc *TelegramWebhookContext) AppUserIntID() (appUserIntID int64) {
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
	whc.GetLogger().Debugf("*TelegramWebhookContext.AppUserIntID(): %v", appUserIntID)
	return
}

func (whc *TelegramWebhookContext) GetAppUser() (bots.AppUser, error) {
	appUserID := whc.AppUserIntID()
	appUser := whc.AppContext.NewAppUserEntity()
	err := whc.AppUserStore.GetAppUserByID(appUserID, appUser)
	return appUser, err
}

func (whc *TelegramWebhookContext) MessageText() string {
	inputMessage := whc.WebhookInput.InputMessage()
	if inputMessage != nil {
		return inputMessage.Text()
	}
	return ""
}

func (whc *TelegramWebhookContext) BotChatID() (chatId interface{}) {
	webhookInput := whc.WebhookInput
	switch webhookInput.InputType() {
	case bots.WebhookInputMessage:
		chatId = webhookInput.InputMessage().Chat().GetID()
	case bots.WebhookInputCallbackQuery:
		callbackQuery := webhookInput.InputCallbackQuery()
		if callbackQuery == nil {
			return nil
		}
		chat := callbackQuery.Chat()
		if chat != nil {
			return chat.GetID()
		}
		data := callbackQuery.GetData()
		if strings.Contains(data, "chat=") {
			values, err := url.ParseQuery(data)
			if err != nil {
				whc.GetLogger().Errorf("Failed to GetData() from webhookInput.InputCallbackQuery()")
				return nil
			}
			chatIdAsStr := values.Get("chat")
			chatIdAsInt, err := strconv.Atoi(chatIdAsStr)
			if err != nil {
				whc.GetLogger().Errorf("Failed to parse 'chat' parameter to int: %v", err)
				return nil
			}
			if chatIdAsInt == 0 {
				return nil
			}
			return chatIdAsInt
		}
		return nil
	}

	if chatId == 0 {
		return nil
	}
	return chatId
}

func (whc *TelegramWebhookContext) ChatEntity() bots.BotChat {
	if whc.BotChatID() == nil {
		return nil
	}
	botChatEntity, err := whc.WebhookContextBase.ChatEntity(whc)
	if err != nil {
		panic(fmt.Sprintf("Failed to get BotChat entity: %v", err))
	}
	return botChatEntity
}

func (whc *TelegramWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	if telegramChat, ok := whc.ChatEntity().(*TelegramChat); ok && telegramChat != nil {
		return whc.InputMessage().Sequence() > telegramChat.LastProcessedUpdateID
	}
	return false
}

func (whc *TelegramWebhookContext) NewChatEntity() bots.BotChat {
	return new(TelegramChat)
}

func (whc *TelegramWebhookContext) getTelegramSenderID() int {
	senderID := whc.GetSender().GetID()
	if tgUserID, ok := senderID.(int); ok {
		return tgUserID
	}
	panic("int expected")
}

func (whc *TelegramWebhookContext) MakeChatEntity() bots.BotChat {
	telegramChat := whc.InputMessage().Chat()
	chatEntity := TelegramChat{
		BotChatEntity: bots.BotChatEntity{
			Type:  telegramChat.GetType(),
			Title: telegramChat.GetFullName(),
		},
		TelegramUserID: whc.getTelegramSenderID(),
	}
	return &chatEntity
}

func (tc *TelegramWebhookContext) NewTgMessage(text string) tgbotapi.MessageConfig {
	//inputMessage := tc.InputMessage()
	//if inputMessage != nil {
	//ctx := tc.Context()
	//chat := inputMessage.Chat()
	//chatID := chat.GetID()
	//log.Infof(ctx, "NewTgMessage(): tc.update.Message.Chat.ID: %v", chatID)
	botChatID := tc.BotChatID()
	if botChatID == nil {
		panic(fmt.Sprintf("Not able to send message as BotChatID() returned nil. text: %v", text))
	}
	if int64ID, ok := botChatID.(int64); ok {
		return tgbotapi.NewMessage(int64ID, text)
	} else {
		if intID, ok := botChatID.(int); ok {
			return tgbotapi.NewMessage(int64(intID), text)
		} else {
			panic(fmt.Sprintf("OK=%v;Expected int or int64, got: %T", ok, botChatID))
		}
	}
	//}
	//panic(fmt.Sprintf("Expected to be called just for inputType == Message, got: %v", tc.InputType()))
}

func (tc *TelegramWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	if telegramChat, ok := chatEntity.(*TelegramChat); ok {
		telegramChat.LastProcessedUpdateID = tc.InputMessage().Sequence()
		return nil
	}
	return errors.New(fmt.Sprintf("Expected *TelegramChat, got: %T", chatEntity))
}
