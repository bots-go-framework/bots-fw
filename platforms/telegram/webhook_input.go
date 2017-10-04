package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"time"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
)

type telegramWebhookInput struct {
	update *tgbotapi.Update // TODO: Make a pointer?
	logRequest func()
}

type TelegramWebhookInput interface {
	TgUpdate() *tgbotapi.Update
}

func (whi telegramWebhookInput) LogRequest() {
	if whi.logRequest != nil {
		whi.logRequest()
	}
}

var _ TelegramWebhookInput = (*telegramWebhookInput)(nil)

type TelegramWebhookUpdateProvider interface {
	TgUpdate() *tgbotapi.Update
}

func (whi telegramWebhookInput) TgUpdate() *tgbotapi.Update {
	return whi.update
}

var _ bots.WebhookInput = (*TelegramWebhookTextMessage)(nil)
var _ bots.WebhookInput = (*TelegramWebhookContactMessage)(nil)
var _ bots.WebhookInput = (*TelegramWebhookInlineQuery)(nil)
var _ bots.WebhookInput = (*TelegramWebhookChosenInlineResult)(nil)
var _ bots.WebhookInput = (*TelegramWebhookCallbackQuery)(nil)
var _ bots.WebhookInput = (*TelegramWebhookNewChatMembersMessage)(nil)

func (whi telegramWebhookInput) GetID() interface{} {
	return whi.update.UpdateID
}

func NewTelegramWebhookInput(update *tgbotapi.Update, logRequest func()) (bots.WebhookInput, error) {
	input := telegramWebhookInput{update: update, logRequest: logRequest}

	switch {

	case update.InlineQuery != nil:
		return NewTelegramWebhookInlineQuery(input), nil

	case update.CallbackQuery != nil:
		return NewTelegramWebhookCallbackQuery(input), nil

	case update.ChosenInlineResult != nil:
		return NewTelegramWebhookChosenInlineResult(input), nil

	default:

		message2input := func(tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) bots.WebhookInput {
			switch {
			case tgMessage.Text != "":
				return NewTelegramWebhookTextMessage(input, tgMessageType, tgMessage)
			case tgMessage.Contact != nil:
				return NewTelegramWebhookContact(input)
			case tgMessage.NewChatMembers != nil:
				return NewTelegramWebhookNewChatMembersMessage(input)
			case tgMessage.LeftChatMember != nil:
				return NewTelegramWebhookLeftChatMembersMessage(input)
			case tgMessage.Voice != nil:
				return NewTelegramWebhookVoiceMessage(input, tgMessageType, tgMessage)
			case tgMessage.Photo != nil:
				return NewTelegramWebhookPhotoMessage(input, tgMessageType, tgMessage)
			case tgMessage.Audio != nil:
				return NewTelegramWebhookAudioMessage(input, tgMessageType, tgMessage)
			case tgMessage.Sticker != nil:
				return NewTelegramWebhookStickerMessage(input, tgMessageType, tgMessage)
			default:
				return nil
			}
		}

		switch {

		case update.Message != nil:
			return message2input(TelegramMessageTypeRegular, update.Message), nil

		case update.EditedMessage != nil:
			return message2input(TelegramMessageTypeEdited, update.EditedMessage), nil

		case update.ChannelPost != nil:
			channelPost, _ := ffjson.MarshalFast(update.ChannelPost)
			return nil, errors.WithMessage(bots.ErrNotImplemented, "ChannelPost is not supported at the moment: " + string(channelPost))
			//return message2input(TelegramMessageTypeChannelPost, update.ChannelPost), nil

		case update.EditedChannelPost != nil:
			editedChannelPost, _ := ffjson.MarshalFast(update.EditedChannelPost)
			return nil, errors.WithMessage(bots.ErrNotImplemented, "EditedChannelPost is not supported at the moment: " + string(editedChannelPost))
			//	return message2input(TelegramMessageTypeEditedChannelPost, update.EditedChannelPost), nil

		default:
			return nil, bots.ErrNotImplemented

		}
	}
}

func (whi telegramWebhookInput) GetSender() bots.WebhookSender {
	switch {
	case whi.update.Message != nil:
		return TelegramSender{tgUser: whi.update.Message.From}
	case whi.update.EditedMessage != nil:
		return TelegramSender{tgUser: whi.update.EditedMessage.From}
	case whi.update.CallbackQuery != nil:
		return TelegramSender{tgUser: whi.update.CallbackQuery.From}
	case whi.update.InlineQuery != nil:
		return TelegramSender{tgUser: whi.update.InlineQuery.From}
	case whi.update.ChosenInlineResult != nil:
		return TelegramSender{tgUser: whi.update.ChosenInlineResult.From}
	//case whi.update.ChannelPost != nil:
	//	chat := whi.update.ChannelPost.Chat
	//	return TelegramSender{  // TODO: Seems to be dirty hack.
	//		tgUser: &tgbotapi.User{
	//			ID: int(chat.ID),
	//			Name: chat.Name,
	//			FirstName: chat.FirstName,
	//			LastName: chat.LastName,
	//		},
	//	}
	default:
		panic("Unknown From sender")
	}
}

func (whi telegramWebhookInput) GetRecipient() bots.WebhookRecipient {
	panic("Not implemented")
}

func (whi telegramWebhookInput) GetTime() time.Time {
	if whi.update.Message != nil {
		return whi.update.Message.Time()
	}
	if whi.update.EditedMessage != nil {
		return whi.update.EditedMessage.Time()
	}
	return time.Time{}
}

func (whi telegramWebhookInput) StringID() string {
	return ""
}

func (whi telegramWebhookInput) TelegramChatID() int64 {
	if whi.update.Message != nil {
		return whi.update.Message.Chat.ID
	}
	if whi.update.EditedMessage != nil {
		return whi.update.EditedMessage.Chat.ID
	}
	panic("Can't get Telgram chat ID from `update.Message` or `update.EditedMessage`.")
}

func (whi telegramWebhookInput) Chat() bots.WebhookChat {
	update := whi.update
	if update.Message != nil {
		return TelegramWebhookChat{
			chat: update.Message.Chat,
		}
	} else if update.EditedMessage != nil {
		return TelegramWebhookChat{
			chat: update.EditedMessage.Chat,
		}
	} else if callbackQuery := update.CallbackQuery; callbackQuery != nil && callbackQuery.Message != nil {
		return TelegramWebhookChat{
			chat: callbackQuery.Message.Chat,
		}
	}
	return nil
}
