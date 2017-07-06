package telegram_bot

import (
	//"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	//"github.com/strongo/app/log"
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
	//whi          TelegramWebhookInput
}

var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func (twhc *TelegramWebhookContext) NewEditCallbackMessage(messageText string) (bots.MessageFromBot, error) {
	return twhc.NewEditCallbackMessageTextAndKeyboard(messageText, tgbotapi.InlineKeyboardMarkup{})
}

func (twhc *TelegramWebhookContext) NewEditCallbackMessageTextAndKeyboard(text string, kbMarkup tgbotapi.InlineKeyboardMarkup) (m bots.MessageFromBot, err error) {
	m = twhc.NewMessage(text)
	update := twhc.Input().(TelegramWebhookCallbackQuery).update

	var (
		inlineMessageID string
		chatID int64
		messageID int
	)

	if inlineMessageID := update.CallbackQuery.InlineMessageID; inlineMessageID == "" {
		messageID = update.Message.MessageID
		chatID = update.Message.Chat.ID
	}
	if text == "" {
		editMessageMarkupConfig := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, inlineMessageID, kbMarkup)
		m.TelegramEditMessageMarkup = &editMessageMarkupConfig
	} else {
		editMessageTextConfig := tgbotapi.NewEditMessageText(chatID, messageID, inlineMessageID, text)
		if len(kbMarkup.InlineKeyboard) > 0 {
			editMessageTextConfig.ReplyMarkup = &kbMarkup
		}
		m.TelegramEditMessageText = editMessageTextConfig
	}

	return m, nil
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
		//whi: whi,
	}
}

func (twhc TelegramWebhookContext) Close(c context.Context) error {
	return nil
}

func (twhc TelegramWebhookContext) Responder() bots.WebhookResponder {
	return twhc.responder
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

func (twhc *TelegramWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (twhc *TelegramWebhookContext) BotApi() *tgbotapi.BotAPI {
	return tgbotapi.NewBotAPIWithClient(twhc.BotContext.BotSettings.Token, twhc.GetHttpClient())
}

func (twhc *TelegramWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := twhc.AppUserIntID()
	appUser := twhc.BotAppContext().NewBotAppUserEntity()
	err := twhc.BotAppUserStore.GetAppUserByID(twhc.Context(), appUserID, appUser)
	return appUser, err
}

func (twhc *TelegramWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	return true
	//if telegramChat, ok := whc.ChatEntity().(*TelegramChatEntity); ok && telegramChat != nil {
	//	return whc.Input().whi.update.UpdateID > telegramChat.LastProcessedUpdateID
	//}
	//return false
}

func (twhc *TelegramWebhookContext) NewChatEntity() bots.BotChat {
	return new(TelegramChatEntity)
}

func (twhc *TelegramWebhookContext) getTelegramSenderID() int {
	senderID := twhc.Input().GetSender().GetID()
	if tgUserID, ok := senderID.(int); ok {
		return tgUserID
	}
	panic("int expected")
}

func (twhc *TelegramWebhookContext) NewTgMessage(text string) tgbotapi.MessageConfig {
	//inputMessage := tc.InputMessage()
	//if inputMessage != nil {
	//ctx := tc.Context()
	//entity := inputMessage.Chat()
	//chatID := entity.GetID()
	//log.Infof(ctx, "NewTgMessage(): tc.update.Message.Chat.ID: %v", chatID)
	botChatID := twhc.BotChatID()
	if botChatID == "" {
		panic(fmt.Sprintf("Not able to send message as BotChatID() returned empty string. text: %v", text))
	}
	botChatIntID, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Not able to parse BotChatID(%v) as int: %v", botChatID, err))
	}
	return tgbotapi.NewMessage(botChatIntID, text)
}

func (twhc *TelegramWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	return nil
	//if telegramChat, ok := chatEntity.(*TelegramChatEntity); ok {
	//	telegramChat.LastProcessedUpdateID = tc.whi.update.UpdateID
	//	return nil
	//}
	//return errors.New(fmt.Sprintf("Expected *TelegramChatEntity, got: %T", chatEntity))
}
