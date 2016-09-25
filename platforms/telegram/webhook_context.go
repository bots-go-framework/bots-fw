package telegram_bot

import (
	"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	//"google.golang.org/appengine/log"
	"github.com/strongo/measurement-protocol"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
}

var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func (whc *TelegramWebhookContext) NewEditCallbackMessage(messageText string) bots.MessageFromBot {
	chatID, _ := whc.BotChatID().(int64)
	messageID := whc.InputCallbackQuery().GetMessage().IntID()
	editMessageTextConfig := tgbotapi.NewEditMessageText(chatID, (int)(messageID), messageText)
	editMessageTextConfig.ParseMode = "HTML"
	m := whc.NewMessage("")
	m.TelegramEditMessageText = editMessageTextConfig
	return m
}

func NewEditCallbackMessageKeyboard(whc bots.WebhookContext, kbMarkup tgbotapi.InlineKeyboardMarkup) bots.MessageFromBot {
	//whct := whc.(*TelegramWebhookContext)
	chatID, _ := whc.BotChatID().(int64)
	messageID := whc.InputCallbackQuery().GetMessage().IntID()
	editMessageMarkupConfig := tgbotapi.NewEditMessageReplyMarkup(chatID, (int)(messageID), kbMarkup)
	m := whc.NewMessage("")
	m.TelegramEditMessageMarkup = &editMessageMarkupConfig
	return m
}

func NewTelegramWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) *TelegramWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		TelegramPlatform{},
		botContext,
		webhookInput,
		botCoreStores,
		gaMeasurement,
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

//func (tc TelegramBotApiUser) IdAsInt64() int64 {
//	return int64(tc.user.ID)
//}

func (whc *TelegramWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc *TelegramWebhookContext) BotApi() *tgbotapi.BotAPI {
	return tgbotapi.NewBotAPIWithClient(whc.BotContext.BotSettings.Token, whc.GetHttpClient())
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
	return
}

func (whc *TelegramWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := whc.AppUserIntID()
	appUser := whc.BotAppContext().NewBotAppUserEntity()
	err := whc.BotAppUserStore.GetAppUserByID(appUserID, appUser)
	return appUser, err
}

func (whc *TelegramWebhookContext) MessageText() string {
	inputMessage := whc.WebhookInput.InputMessage()
	if inputMessage != nil {
		return inputMessage.Text()
	}
	return ""
}

func (whc *TelegramWebhookContext) BotChatID() interface{} {
	id := whc.BotChatIntID()
	if id == 0 {
		return nil
	}
	return id
}

func (whc *TelegramWebhookContext) BotChatIntID() (chatId int64) {
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
	telegramChat := whc.Chat()
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
