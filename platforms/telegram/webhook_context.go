package telegram_bot

import (
	//"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	//"google.golang.org/appengine/log"
	"github.com/strongo/measurement-protocol"
	"net/http"
	"strconv"
	"golang.org/x/net/context"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
	//input          telegramWebhookInput
}

var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func (whc *TelegramWebhookContext) NewEditCallbackMessage(messageText string) bots.MessageFromBot {
	chatID, err := strconv.ParseInt(whc.BotChatID(), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse whc.BotChatID() to int: %v", err))
	}
	callbackQuery := whc.Input().(bots.WebhookCallbackQuery)
	message := callbackQuery.GetMessage()
	messageID := message.IntID()
	editMessageTextConfig := tgbotapi.NewEditMessageText(chatID, (int)(messageID), messageText)
	editMessageTextConfig.ParseMode = "HTML"
	m := whc.NewMessage("")
	m.TelegramEditMessageText = editMessageTextConfig
	return m
}

func NewEditCallbackMessageKeyboard(whc bots.WebhookContext, kbMarkup tgbotapi.InlineKeyboardMarkup) bots.MessageFromBot {
	//whct := whc.(*TelegramWebhookContext)
	chatID, err := strconv.ParseInt(whc.BotChatID(), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse whc.BotChatID() to int: %v", err))
	}
	messageID := whc.Input().(bots.WebhookCallbackQuery).GetMessage().IntID()
	editMessageMarkupConfig := tgbotapi.NewEditMessageReplyMarkup(chatID, (int)(messageID), kbMarkup)
	m := whc.NewMessage("")
	m.TelegramEditMessageMarkup = &editMessageMarkupConfig
	return m
}

func NewTelegramWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, input bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) *TelegramWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		TelegramPlatform{},
		botContext,
		input,
		botCoreStores,
		gaMeasurement,
	)
	return &TelegramWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
		//input: input,
	}
}

func (tc TelegramWebhookContext) Close(c context.Context) error {
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

func (whc *TelegramWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := whc.AppUserIntID()
	appUser := whc.BotAppContext().NewBotAppUserEntity()
	err := whc.BotAppUserStore.GetAppUserByID(whc.Context(), appUserID, appUser)
	return appUser, err
}

func (whc *TelegramWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	return true
	//if telegramChat, ok := whc.ChatEntity().(*TelegramChat); ok && telegramChat != nil {
	//	return whc.Input().input.update.UpdateID > telegramChat.LastProcessedUpdateID
	//}
	//return false
}

func (whc *TelegramWebhookContext) NewChatEntity() bots.BotChat {
	return new(TelegramChat)
}

func (whc *TelegramWebhookContext) getTelegramSenderID() int {
	senderID := whc.Input().GetSender().GetID()
	if tgUserID, ok := senderID.(int); ok {
		return tgUserID
	}
	panic("int expected")
}

func (tc *TelegramWebhookContext) NewTgMessage(text string) tgbotapi.MessageConfig {
	//inputMessage := tc.InputMessage()
	//if inputMessage != nil {
	//ctx := tc.Context()
	//chat := inputMessage.Chat()
	//chatID := chat.GetID()
	//log.Infof(ctx, "NewTgMessage(): tc.update.Message.Chat.ID: %v", chatID)
	botChatID := tc.BotChatID()
	if botChatID == "" {
		panic(fmt.Sprintf("Not able to send message as BotChatID() returned empty string. text: %v", text))
	}
	botChatIntID, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Not able to parse BotChatID(%v) as int: %v", botChatID, err))
	}
	return tgbotapi.NewMessage(botChatIntID, text)
}

func (tc *TelegramWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	return nil
	//if telegramChat, ok := chatEntity.(*TelegramChat); ok {
	//	telegramChat.LastProcessedUpdateID = tc.input.update.UpdateID
	//	return nil
	//}
	//return errors.New(fmt.Sprintf("Expected *TelegramChat, got: %T", chatEntity))
}
