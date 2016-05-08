package telegram_bot

import (
	"errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
}
var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func NewTelegramWebhookContext(r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput) *TelegramWebhookContext {
	return &TelegramWebhookContext{
		//update: update,
		WebhookContextBase: bots.NewWebhookContextBase(r, botContext, webhookInput, botContext.BotHost.GetBotChatStore("telegram", r)),
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

func (tc TelegramWebhookContext) ApiUser() bots.BotApiUser {
	return TelegramBotApiUser{user: tc.TelegramApiUser()}
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
	return whc.WebhookContextBase.ChatEntity(whc)
}

func (whc *TelegramWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	if telegramChat, ok := whc.ChatEntity().(*TelegramChat); ok && telegramChat != nil {
		return whc.InputMessage().Sequence() > telegramChat.LastProcessedUpdateID
	}
	return false
}

func (whc *TelegramWebhookContext) TelegramApiUser() tgbotapi.User {
	panic("Not implemented yet") //return whc.update.Message.From
}

func (whc *TelegramWebhookContext) GetUser() (*datastore.Key, bots.AppUser, error) {
	return whc.getUserByTelegramID(whc.Context(), whc.TelegramApiUser().ID, true)
}

func (whc *TelegramWebhookContext) getUserByTelegramID(ctx context.Context, telegramUserID int, createIfMissing bool) (*datastore.Key, bots.AppUser, error) {
	//telegramUser := TelegramUser{}
	//gae_host.NewGaeTelegramChatStore()
	//err := GetBotUserEntity(ctx, NewTelegramUserEntityKey(ctx, telegramUserID), &telegramUser)
	//if err != nil {
	//	return nil, nil, err
	//}
	//userKey := datastore.NewKey(ctx, common.AppUserKind, "", telegramUser.UserID, nil)
	//user := common.AppUser{}
	//err = nds.Get(ctx, userKey, &user)
	//return userKey, &user, err
	return nil, nil, nil
}

func (whc *TelegramWebhookContext) NewChatEntity() bots.BotChat {
	return new(TelegramChat)
}

func (whc *TelegramWebhookContext) MakeChatEntity() bots.BotChat {
	telegramChat := whc.InputMessage().Chat()
	chatEntity := TelegramChat{
		BotChatEntity: bots.BotChatEntity{
			Type:  telegramChat.GetType(),
			Title: telegramChat.GetTitle(),
		},
		BotUserID: (int64)(whc.TelegramApiUser().ID),
	}
	return &chatEntity
}

func (whc *TelegramWebhookContext) GetOrCreateUserEntity() (bots.BotUser, error) {
	return nil, bots.NotImplementedError
	//return gae_host.GetOrCreateUserEntity(whc.Context(), whc.update)
}

func (whc *TelegramWebhookContext) GetOrCreateUser() (*datastore.Key, bots.AppUser, error) {
	return whc.getUserByTelegramID(whc.Context(), whc.TelegramApiUser().ID, true)
}

func (tc *TelegramWebhookContext) NewTgMessage(text string) tgbotapi.MessageConfig {
	log.Infof(tc.Context(), "NewTgMessage(): tc.update.Message.Chat.ID: %v", tc.InputMessage().Chat().GetID())
	if intID, ok := tc.BotChatID().(int64); ok {
		return tgbotapi.NewMessage((int)(intID), text)
	} else {
		panic("tc.BotChatID.(int64) is not OK")
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

func (tc *TelegramWebhookContext) ReplyByBot(m bots.MessageFromBot) error {
	return nil
}
