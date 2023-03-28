package telegram

import (
	//"errors"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/dalgo/dal"
	//"github.com/strongo/log"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"net/http"
	"strconv"
)

type tgWebhookContext struct {
	*bots.WebhookContextBase
	tgInput TgWebhookInput
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
	//whi          tgWebhookInput

	// This 3 props are cache for getLocalAndChatIDByChatInstance()
	isInGroup bool
	locale    string
	chatID    string
}

var _ bots.WebhookContext = (*tgWebhookContext)(nil)

func (twhc *tgWebhookContext) NewEditMessage(text string, format bots.MessageFormat) (m bots.MessageFromBot, err error) {
	m.Text = text
	m.Format = format
	m.IsEdit = true
	return
}

func (twhc *tgWebhookContext) CreateOrUpdateTgChatInstance() (err error) {
	c := twhc.Context()
	log.Debugf(c, "*tgWebhookContext.CreateOrUpdateTgChatInstance()")
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
		tgChatEntity := twhc.ChatEntity().(TgChatEntity)
		if tgChatEntity.GetTgChatInstanceID() != chatInstanceID {
			tgChatEntity.SetTgChatInstanceID(chatInstanceID)
			//if err = twhc.SaveBotChat(c, twhc.GetBotCode(), twhc.MustBotChatID(), tgChatEntity.(bots.BotChat)); err != nil {
			//	return
			//}
		}

		var chatInstance ChatInstance
		preferredLanguage := tgChatEntity.GetPreferredLanguage()
		if DAL.DB == nil {
			panic("telegram.DAL.DB is nil")
		}
		if err = DAL.DB.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
			log.Debugf(c, "CreateOrUpdateTgChatInstance() => checking tg chat instance within tx")
			changed := false
			if chatInstance, err = DAL.TgChatInstance.GetTelegramChatInstanceByID(c, tx, chatInstanceID); err != nil {
				if !dal.IsNotFound(err) {
					return
				}
				log.Debugf(c, "CreateOrUpdateTgChatInstance() => new tg chat instance")
				chatInstance = DAL.TgChatInstance.NewTelegramChatInstance(chatInstanceID, chatID, preferredLanguage)
				changed = true
			} else { // Update if needed
				log.Debugf(c, "CreateOrUpdateTgChatInstance() => existing tg chat instance")
				if tgChatInstanceId := chatInstance.Data.GetTgChatID(); tgChatInstanceId != chatID {
					err = fmt.Errorf("chatInstance.GetTgChatID():%d != chatID:%d", tgChatInstanceId, chatID)
				} else if prefLang := chatInstance.Data.GetPreferredLanguage(); prefLang != preferredLanguage {
					chatInstance.Data.SetPreferredLanguage(preferredLanguage)
					changed = true
				}
			}
			if changed {
				log.Debugf(c, "Saving tg chat instance...")
				if err = DAL.TgChatInstance.SaveTelegramChatInstance(c, chatInstance); err != nil {
					return
				}
			}
			return
		}, dal.TxWithCrossGroup()); err != nil {
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
	input TgWebhookInput,
	botCoreStores bots.BotCoreStores,
	gaMeasurement bots.GaQueuer,
) *tgWebhookContext {
	twhc := &tgWebhookContext{
		tgInput: input.(TgWebhookInput),
	}
	chat := twhc.tgInput.TgUpdate().Chat()

	isInGroup := func() bool { // Checks if current chat is a group chat
		if chat != nil && chat.IsGroup() {
			return true
		}

		if callbackQuery := twhc.tgInput.TgUpdate().CallbackQuery; callbackQuery != nil && callbackQuery.ChatInstance != "" {
			c := botContext.BotHost.Context(r)
			var isGroupChat bool
			err := twhc.RunReadwriteTransaction(c, func(ctx context.Context, tx dal.ReadwriteTransaction) error {

				if chatInstance, err := DAL.TgChatInstance.GetTelegramChatInstanceByID(c, tx, callbackQuery.ChatInstance); err != nil {
					return err
				} else if chatInstance.Data != nil {
					isGroupChat = chatInstance.Data.GetTgChatID() < 0
				}
				return nil
			})
			if err != nil {
				if !dal.IsNotFound(err) {
					log.Errorf(c, "failed to get tg chat instance: %v", err)
				}
			}
			return isGroupChat
		}

		return false
	}

	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		Platform{},
		botContext,
		input.(bots.WebhookInput),
		botCoreStores,
		gaMeasurement,
		isInGroup,
		twhc.getLocalAndChatIDByChatInstance,
	)
	twhc.WebhookContextBase = whcb
	return twhc
}

func (twhc tgWebhookContext) Close(c context.Context) error {
	return nil
}

func (twhc tgWebhookContext) Responder() bots.WebhookResponder {
	return twhc.responder
}

type tgBotAPIUser struct {
	user tgbotapi.User
}

func (tc tgBotAPIUser) FirstName() string {
	return tc.user.FirstName
}

func (tc tgBotAPIUser) LastName() string {
	return tc.user.LastName
}

//func (tc tgBotAPIUser) IdAsString() string {
//	return ""
//}

//func (tc tgBotAPIUser) IdAsInt64() int64 {
//	return int64(tc.user.ID)
//}

func (twhc *tgWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (twhc *tgWebhookContext) BotAPI() *tgbotapi.BotAPI {
	botContext := twhc.BotContext()
	return tgbotapi.NewBotAPIWithClient(botContext.BotSettings.Token, botContext.BotHost.GetHTTPClient(twhc.Context()))
}

func (twhc *tgWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	appUserID := twhc.AppUserIntID()
	appUser := twhc.BotAppContext().NewBotAppUserEntity()
	err := twhc.BotAppUserStore.GetAppUserByID(twhc.Context(), appUserID, appUser)
	return appUser, err
}

func (twhc *tgWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	return true
	//if telegramChat, ok := whc.ChatEntity().(*TgChatEntityBase); ok && telegramChat != nil {
	//	return whc.Input().whi.update.UpdateID > telegramChat.LastProcessedUpdateID
	//}
	//return false
}

func (twhc *tgWebhookContext) NewChatEntity() bots.BotChat {
	return new(TgChatEntityBase)
}

func (twhc *tgWebhookContext) getTelegramSenderID() int {
	senderID := twhc.Input().GetSender().GetID()
	if tgUserID, ok := senderID.(int); ok {
		return tgUserID
	}
	panic("int expected")
}

func (twhc *tgWebhookContext) NewTgMessage(text string) tgbotapi.MessageConfig {
	//inputMessage := tc.InputMessage()
	//if inputMessage != nil {
	//ctx := tc.Context()
	//Data := inputMessage.Chat()
	//chatID := Data.GetID()
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

func (twhc *tgWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	return nil
	//if telegramChat, ok := chatEntity.(*TgChatEntityBase); ok {
	//	telegramChat.LastProcessedUpdateID = tc.whi.update.UpdateID
	//	return nil
	//}
	//return fmt.Errorf("Expected *TgChatEntityBase, got: %T", chatEntity)
}

func (twhc *tgWebhookContext) getLocalAndChatIDByChatInstance(c context.Context) (locale, chatID string, err error) {
	log.Debugf(c, "*tgWebhookContext.getLocalAndChatIDByChatInstance()")
	if chatID == "" && locale == "" { // we need to cache to make sure not called within transaction
		if cbq := twhc.tgInput.TgUpdate().CallbackQuery; cbq != nil && cbq.ChatInstance != "" {
			if cbq.Message != nil && cbq.Message.Chat != nil && cbq.Message.Chat.ID != 0 {
				log.Errorf(c, "getLocalAndChatIDByChatInstance() => should not be here")
			} else {
				if chatInstance, err := DAL.TgChatInstance.GetTelegramChatInstanceByID(c, nil, cbq.ChatInstance); err != nil {
					if !dal.IsNotFound(err) {
						return "", "", err
					}
				} else if tgChatID := chatInstance.Data.GetTgChatID(); tgChatID != 0 {
					twhc.chatID = strconv.FormatInt(tgChatID, 10)
					twhc.locale = chatInstance.Data.GetPreferredLanguage()
					twhc.isInGroup = tgChatID < 0
				}
			}
		}
	}
	return twhc.locale, twhc.chatID, nil
}

func (twhc *tgWebhookContext) ChatEntity() bots.BotChat {
	if _, err := twhc.BotChatID(); err != nil {
		log.Errorf(twhc.Context(), errors.WithMessage(err, "whc.BotChatID()").Error())
		return nil
	}
	tgUpdate := twhc.tgInput.TgUpdate()
	if tgUpdate.CallbackQuery != nil {

	}

	return twhc.WebhookContextBase.ChatEntity()
}
