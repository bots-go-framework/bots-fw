package telegram_bot

import (
	//"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	//"github.com/strongo/log"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"net/http"
	"strconv"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	tgInput TelegramWebhookInput
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
	//whi          telegramWebhookInput

	// This 3 props are cache for getLocalAndChatIdByChatInstance()
	isInGroup bool
	locale    string
	chatID    string
}

var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func (twhc *TelegramWebhookContext) NewEditMessage(text string, format bots.MessageFormat) (m bots.MessageFromBot, err error) {
	m.Text = text
	m.Format = format
	m.IsEdit = true
	return
}

func (twhc *TelegramWebhookContext) CreateOrUpdateTgChatInstance() (err error) {
	c := twhc.Context()
	log.Debugf(c, "*TelegramWebhookContext.CreateOrUpdateTgChatInstance()")
	tgUpdate := twhc.tgInput.TgUpdate()
	if tgUpdate.CallbackQuery == nil {
		log.Debugf(c, "CreateOrUpdateTgChatInstance() => tgUpdate.CallbackQuery == nil")
		return
	}
	if chatInstanceID := tgUpdate.CallbackQuery.ChatInstance; chatInstanceID == "" {
		log.Debugf(c, "CreateOrUpdateTgChatInstance() => no chatInstanceID")
	} else {
		chatID := tgUpdate.CallbackQuery.Message.Chat.ID
		log.Debugf(c, "CreateOrUpdateTgChatInstance() => chatID: %v, chatInstanceID: %v", chatID, chatInstanceID)
		if chatID == 0 {
			return
		}
		tgChatEntity := twhc.ChatEntity().(TelegramChatEntity)
		if tgChatEntity.GetTgChatInstanceID() != chatInstanceID {
			tgChatEntity.SetTgChatInstanceID(chatInstanceID)
			//if err = twhc.SaveBotChat(c, twhc.GetBotCode(), twhc.MustBotChatID(), tgChatEntity.(bots.BotChat)); err != nil {
			//	return
			//}
		}

		var chatInstance TelegramChatInstance
		preferredLanguage := tgChatEntity.GetPreferredLanguage()
		if DAL.DB == nil {
			panic("telegram_bot.DAL.DB is nil")
		}
		if err = DAL.DB.RunInTransaction(c, func(c context.Context) (err error) {
			log.Debugf(c, "CreateOrUpdateTgChatInstance() => checking tg chat instance within tx")
			changed := false
			if chatInstance, err = DAL.TgChatInstance.GetTelegramChatInstanceByID(c, chatInstanceID); err != nil {
				if !db.IsNotFound(err) {
					return
				}
				log.Debugf(c, "CreateOrUpdateTgChatInstance() => new tg chat instance")
				chatInstance = DAL.TgChatInstance.NewTelegramChatInstance(chatInstanceID, chatID, preferredLanguage)
				changed = true
			} else { // Update if needed
				log.Debugf(c, "CreateOrUpdateTgChatInstance() => existing tg chat instance")
				if chatInstance.GetTgChatID() != chatID {
					err = fmt.Errorf("chatInstance.GetTgChatID():%d != chatID:%d", chatInstance.GetTgChatID(), chatID)
				} else if chatInstance.GetPreferredLanguage() == preferredLanguage {
					return
				}
				chatInstance.SetPreferredLanguage(preferredLanguage)
				changed = true
			}
			if changed {
				log.Debugf(c, "Saving tg chat instance...")
				if err = DAL.TgChatInstance.SaveTelegramChatInstance(c, chatInstance); err != nil {
					return
				}
			}
			return
		}, db.CrossGroupTransaction); err != nil {
			err = errors.WithMessage(err, "failed to create or update Telegram chat instance")
			return
		}
	}
	return
}

func getTgMessageIDs(update *tgbotapi.Update) (inlineMessageID string, chatID int64, messageID int) {
	if update.CallbackQuery != nil {
		if update.CallbackQuery.InlineMessageID != "" {
			inlineMessageID = update.CallbackQuery.InlineMessageID
		} else if update.CallbackQuery.Message != nil {
			messageID = update.CallbackQuery.Message.MessageID
			chatID = update.CallbackQuery.Message.Chat.ID
		}
	} else if update.Message != nil {
		messageID = update.Message.MessageID
		chatID = update.Message.Chat.ID
	} else if update.EditedMessage != nil {
		messageID = update.EditedMessage.MessageID
		chatID = update.EditedMessage.Chat.ID
	} else if update.ChannelPost != nil {
		messageID = update.ChannelPost.MessageID
		chatID = update.ChannelPost.Chat.ID
	} else if update.ChosenInlineResult != nil {
		if update.ChosenInlineResult.InlineMessageID != "" {
			inlineMessageID = update.ChosenInlineResult.InlineMessageID
		}
	} else if update.EditedChannelPost != nil {
		messageID = update.EditedChannelPost.MessageID
		chatID = update.EditedChannelPost.Chat.ID
	}

	return
}

func newTelegramWebhookContext(
	appContext bots.BotAppContext,
	r *http.Request, botContext bots.BotContext,
	input TelegramWebhookInput,
	botCoreStores bots.BotCoreStores,
	gaMeasurement bots.GaQueuer,
) *TelegramWebhookContext {
	twhc := &TelegramWebhookContext{
		tgInput: input.(TelegramWebhookInput),
	}
	chat := twhc.tgInput.TgUpdate().Chat()

	isInGroup := func() bool {
		if chat != nil && chat.IsGroup() {
			return true
		}

		if callbackQuery := twhc.tgInput.TgUpdate().CallbackQuery; callbackQuery != nil && callbackQuery.ChatInstance != "" {
			c := botContext.BotHost.Context(r)
			if chatInstance, err := DAL.TgChatInstance.GetTelegramChatInstanceByID(c, callbackQuery.ChatInstance); err == nil {
				return chatInstance.GetTgChatID() < 0
			} else if !db.IsNotFound(err) {
				log.Errorf(c, "failed to get tg chat instance: %v", err)
			}
		}

		return false
	}

	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		TelegramPlatform{},
		botContext,
		input.(bots.WebhookInput),
		botCoreStores,
		gaMeasurement,
		isInGroup,
		twhc.getLocalAndChatIdByChatInstance,
	)
	twhc.WebhookContextBase = whcb
	return twhc
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
	return tgbotapi.NewBotAPIWithClient(twhc.BotContext.BotSettings.Token, twhc.BotContext.BotHost.GetHttpClient(twhc.Context()))
}

func (twhc *TelegramWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := twhc.AppUserIntID()
	appUser := twhc.BotAppContext().NewBotAppUserEntity()
	err := twhc.BotAppUserStore.GetAppUserByID(twhc.Context(), appUserID, appUser)
	return appUser, err
}

func (twhc *TelegramWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	return true
	//if telegramChat, ok := whc.ChatEntity().(*TelegramChatEntityBase); ok && telegramChat != nil {
	//	return whc.Input().whi.update.UpdateID > telegramChat.LastProcessedUpdateID
	//}
	//return false
}

func (twhc *TelegramWebhookContext) NewChatEntity() bots.BotChat {
	return new(TelegramChatEntityBase)
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
	botChatID, err := twhc.BotChatID()
	if err != nil {
		panic(err)
	}
	if botChatID == "" {
		panic(fmt.Sprintf("Not able to send message as BotChatID() returned empty string. text: %v", text))
	}
	botChatIntID, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Not able to parse BotChatID(%v) as int: %v", botChatID, err))
	}
	//tgbotapi.NewEditMessageText()
	return tgbotapi.NewMessage(botChatIntID, text)
}

func (twhc *TelegramWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	return nil
	//if telegramChat, ok := chatEntity.(*TelegramChatEntityBase); ok {
	//	telegramChat.LastProcessedUpdateID = tc.whi.update.UpdateID
	//	return nil
	//}
	//return fmt.Errorf("Expected *TelegramChatEntityBase, got: %T", chatEntity)
}

func (twhc *TelegramWebhookContext) getLocalAndChatIdByChatInstance(c context.Context) (locale, chatID string, err error) {
	log.Debugf(c, "*TelegramWebhookContext.getLocalAndChatIdByChatInstance()")
	if chatID == "" && locale == "" { // we need to cache to make sure not called within transaction
		if cbq := twhc.tgInput.TgUpdate().CallbackQuery; cbq != nil && cbq.ChatInstance != "" {
			if cbq.Message != nil && cbq.Message.Chat != nil && cbq.Message.Chat.ID != 0 {
				log.Errorf(c, "getLocalAndChatIdByChatInstance() => should not be here")
			} else {
				if chatInstance, err := DAL.TgChatInstance.GetTelegramChatInstanceByID(c, cbq.ChatInstance); err != nil {
					if !db.IsNotFound(err) {
						return "", "", err
					}
				} else if tgChatID := chatInstance.GetTgChatID(); tgChatID != 0 {
					twhc.chatID = strconv.FormatInt(tgChatID, 10)
					twhc.locale = chatInstance.GetPreferredLanguage()
					twhc.isInGroup = tgChatID < 0
				}
			}
		}
	}
	return twhc.locale, twhc.chatID, nil
}

func (twhc *TelegramWebhookContext) ChatEntity() bots.BotChat {
	if _, err := twhc.BotChatID(); err != nil {
		log.Errorf(twhc.Context(), errors.WithMessage(err, "whc.BotChatID()").Error())
		return nil
	}
	tgUpdate := twhc.tgInput.TgUpdate()
	if tgUpdate.CallbackQuery != nil {

	}

	return twhc.WebhookContextBase.ChatEntity()
}
