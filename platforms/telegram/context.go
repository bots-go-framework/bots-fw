package telegram_bot

import (
	"errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/log"
	"net/http"
	"fmt"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
}
var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func NewTelegramWebhookContext(appContext bots.AppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores) *TelegramWebhookContext {
	whcb := bots.NewWebhookContextBase(
			r,
			appContext,
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

func (whc *TelegramWebhookContext) AppUserID() int64 {
	return whc.ChatEntity().GetAppUserID()
}

func (whc *TelegramWebhookContext) MessageText() string {
	return whc.WebhookInput.InputMessage().Text()
}

func (whc *TelegramWebhookContext) BotChatID() interface {} {
	chatId := whc.WebhookInput.InputMessage().Chat().GetID()
	whc.GetLogger().Infof("BotChatID(): %v", chatId)
	return chatId
}

func (whc *TelegramWebhookContext) ChatEntity() bots.BotChat {
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
			Title: telegramChat.GetTitle(),
		},
		TelegramUserID: whc.getTelegramSenderID(),
	}
	return &chatEntity
}

func (tc *TelegramWebhookContext) NewTgMessage(text string) tgbotapi.MessageConfig {
	log.Infof(tc.Context(), "NewTgMessage(): tc.update.Message.Chat.ID: %v", tc.InputMessage().Chat().GetID())
	botChatID := tc.BotChatID()
	if intID, ok := botChatID.(int); ok {
		return tgbotapi.NewMessage(intID, text)
	} else {
		panic(fmt.Sprintf("Expected int, got: %T", botChatID))
	}
}

func (tc *TelegramWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	telegramChat, ok := chatEntity.(*TelegramChat)
	if !ok {
		return errors.New("Failed to cast: chatEntity.(*TelegramChat)")
	}
	telegramChat.LastProcessedUpdateID = tc.InputMessage().Sequence()
	return nil
}

func (tc *TelegramWebhookContext) ReplyByBot(w http.ResponseWriter, m bots.MessageFromBot) error {
	return nil
}
