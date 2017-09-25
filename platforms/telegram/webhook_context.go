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
	"errors"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
)

type TelegramWebhookContext struct {
	*bots.WebhookContextBase
	tgInput TelegramWebhookInput
	//update         tgbotapi.Update // TODO: Consider removing?
	responseWriter http.ResponseWriter
	responder      bots.WebhookResponder
	//whi          telegramWebhookInput
}

var _ bots.WebhookContext = (*TelegramWebhookContext)(nil)

func (twhc *TelegramWebhookContext) NewEditMessage(text string, format bots.MessageFormat) (m bots.MessageFromBot, err error) {
	m, err = twhc.NewEditMessageTextAndKeyboard(text, nil)
	m.Format = format
	return
}

func (twhc *TelegramWebhookContext) CreateOrUpdateTgChatInstance() (err error) {
	c := twhc.Context()
	log.Debugf(c, "*TelegramWebhookContext.CreateOrUpdateTgChatInstance()")
	tgUpdate := twhc.tgInput.TgUpdate()
	if tgUpdate.CallbackQuery == nil {
		return
	}
	if chatInstanceID := tgUpdate.CallbackQuery.ChatInstance; chatInstanceID != "" {
		chatID := tgUpdate.CallbackQuery.Message.Chat.ID
		if chatID == 0 {
			return
		}
		tgChatEntity := twhc.ChatEntity().(TelegramChatEntity)
		if tgChatEntity.GetTgChatInstanceID() != "" {
			return
		}
		tgChatEntity.SetTgChatInstanceID(chatInstanceID)
		var chatInstance TelegramChatInstance
		preferredLanguage := tgChatEntity.GetPreferredLanguage()
		if DAL.DB == nil {
			panic("telegram_bot.DAL.DB is nil")
		}
		err = DAL.DB.RunInTransaction(c, func(c context.Context) (err error) {
			if chatInstance, err = DAL.TgChatInstance.GetTelegramChatInstanceByID(c, chatInstanceID); err != nil {
				if !db.IsNotFound(err) {
					return
				}
				chatInstance = DAL.TgChatInstance.NewTelegramChatInstance(chatInstanceID, chatID, preferredLanguage)
			} else { // Update if needed
				if chatInstance.GetTgChatID() != chatID {
					err = fmt.Errorf("chatInstance.GetTgChatID():%d != chatID:%d", chatInstance.GetTgChatID(), chatID)
				} else if chatInstance.GetPreferredLanguage() == preferredLanguage {
					return
				}
				chatInstance.SetPreferredLanguage(preferredLanguage)
			}
			if err = DAL.TgChatInstance.SaveTelegramChatInstance(c, chatInstance); err != nil {
				return
			}
			return
		}, db.SingleGroupTransaction)
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

func (twhc *TelegramWebhookContext) NewEditMessageTextAndKeyboard(text string, kbMarkup *tgbotapi.InlineKeyboardMarkup) (m bots.MessageFromBot, err error) {
	//TODO: !!! panic from here is not handled properly (!silent!failure!) - verify and add unit tests
	m = twhc.NewMessage(text)

	inlineMessageID, chatID, messageID := getTgMessageIDs(twhc.tgInput.TgUpdate())

	hasKeyboard := kbMarkup != nil && len(kbMarkup.InlineKeyboard) > 0

	if text != "" {
		editMessageTextConfig := tgbotapi.NewEditMessageText(chatID, messageID, inlineMessageID, text)
		if hasKeyboard {
			editMessageTextConfig.ReplyMarkup = kbMarkup
		}
		m.TelegramEditMessageText = editMessageTextConfig
	} else if hasKeyboard {
		editMessageMarkupConfig := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, inlineMessageID, kbMarkup)
		m.TelegramEditMessageMarkup = &editMessageMarkupConfig
	} else {
		err = errors.New("can't edit Telegram message as  m.Text is empty string and no keyboard")
		return
	}
	m.IsEdit = true
	return
}

func newTelegramWebhookContext(
	appContext bots.BotAppContext,
	r *http.Request, botContext bots.BotContext,
	input TelegramWebhookInput,
	botCoreStores bots.BotCoreStores,
	gaMeasurement *measurement.BufferedSender,
) *TelegramWebhookContext {
	whcb := bots.NewWebhookContextBase(
		r,
		appContext,
		TelegramPlatform{},
		botContext,
		input.(bots.WebhookInput),
		botCoreStores,
		gaMeasurement,
	)
	return &TelegramWebhookContext{
		//update: update,
		WebhookContextBase: whcb,
		tgInput: input.(TelegramWebhookInput),
		//whi: whi,
	}
}

func (twhc TelegramWebhookContext) IsInGroup() bool {
	chat := twhc.tgInput.TgUpdate().Chat()
	return chat != nil && chat.IsGroup()
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
	//return errors.New(fmt.Sprintf("Expected *TelegramChatEntityBase, got: %T", chatEntity))
}

var ErrChatInstanceIsNotSet = errors.New("update.CallbackQuery.ChatInstance is empty string")

func (twhc *TelegramWebhookContext) BotChatID() (chatID string, err error) {
	tgUpdate := twhc.tgInput.TgUpdate()
	if cbq := tgUpdate.CallbackQuery; cbq  != nil {
		if cbq.Message != nil && cbq.Message.Chat != nil {
			return strconv.FormatInt(cbq.Message.Chat.ID, 10), nil
		}
		if cbq.ChatInstance == "" {
			err = ErrChatInstanceIsNotSet
			return
		}
		c := twhc.Context()
		if chatInstance, err := DAL.TgChatInstance.GetTelegramChatInstanceByID(c, cbq.ChatInstance); err != nil  {
			return "", err
		} else {
			if tgChatID := chatInstance.GetTgChatID(); tgChatID != 0 {
				return strconv.FormatInt(tgChatID, 10), nil
			}
			return "", nil
		}
	}
	return twhc.WebhookContextBase.BotChatID()
}