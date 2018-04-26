package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"net/http"
	"strconv"
)

type tgWebhookResponder struct {
	w   http.ResponseWriter
	whc *tgWebhookContext
}

var _ bots.WebhookResponder = (*tgWebhookResponder)(nil)

func newTgWebhookResponder(w http.ResponseWriter, whc *tgWebhookContext) tgWebhookResponder {
	responder := tgWebhookResponder{w: w, whc: whc}
	whc.responder = responder
	return responder
}

func (r tgWebhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotAPISendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	log.Debugf(c, "tgWebhookResponder.SendMessage(channel=%v, isEdit=%v)", channel, m.IsEdit)
	if channel != bots.BotAPISendMessageOverHTTPS && channel != bots.BotAPISendMessageOverResponse {
		panic(fmt.Sprintf("Unknown channel: [%v]. Expected either 'https' or 'response'.", channel))
	}
	//ctx := tc.Context()

	var chattable tgbotapi.Chattable

	parseMode := func() string {
		switch m.Format {
		case bots.MessageFormatHTML:
			return "HTML"
		case bots.MessageFormatMarkdown:
			return "Markdown"
		}
		return ""
	}

	tgUpdate := r.whc.Input().(tgWebhookUpdateProvider).TgUpdate()

	var botMessage bots.BotMessage

	if m.Text == bots.NoMessageToSend {
		log.Debugf(c, bots.NoMessageToSend)
		return
	} else if botMessage = m.BotMessage; botMessage != nil {
		log.Debugf(c, "m.BotMessage != nil")
		switch m.BotMessage.BotMessageType() {
		case bots.BotMessageTypeInlineResults:
			chattable = tgbotapi.InlineConfig(m.BotMessage.(InlineBotMessage))
		case bots.BotMessageTypeCallbackAnswer:
			callbackAnswer := tgbotapi.AnswerCallbackQueryConfig(m.BotMessage.(CallbackAnswer))
			if callbackAnswer.CallbackQueryID == "" && tgUpdate.CallbackQuery != nil {
				callbackAnswer.CallbackQueryID = tgUpdate.CallbackQuery.ID
			}
			chattable = callbackAnswer
		case bots.BotMessageTypeLeaveChat:
			leaveChat := tgbotapi.LeaveChatConfig(m.BotMessage.(LeaveChat))
			if leaveChat.ChatID == "" {
				leaveChat.ChatID = strconv.FormatInt(tgUpdate.Chat().ID, 10)
			}
			chattable = leaveChat
		case bots.BotMessageTypeExportChatInviteLink:
			exportChatInviteLink := tgbotapi.ExportChatInviteLink(m.BotMessage.(ExportChatInviteLink))
			if exportChatInviteLink.ChatID == "" {
				exportChatInviteLink.ChatID = strconv.FormatInt(tgUpdate.Chat().ID, 10)
			}
			chattable = exportChatInviteLink
		case bots.BotMessageTypeUndefined:
			err = fmt.Errorf("bot message type %v==undefined", m.BotMessage.BotMessageType())
			return
		default:
			err = fmt.Errorf("unknown bot message type %v==%T", m.BotMessage.BotMessageType(), botMessage)
			return
		}
	} else if m.IsEdit || (tgUpdate.CallbackQuery != nil && tgUpdate.CallbackQuery.InlineMessageID != "" && m.ToChat == nil) {
		if m.IsEdit {
			log.Debugf(c, "m.IsEdit")
		} else if tgUpdate.CallbackQuery != nil {
			log.Debugf(c, "tgUpdate.CallbackQuery != nil")
		}

		// Edit message
		inlineMessageID, chatID, messageID := getTgMessageIDs(tgUpdate)
		if m.EditMessageUID != nil {
			switch m.EditMessageUID.(type) { // TODO: How do we remove duplicates for value & pointer cases?
			case callbackCurrent:
				// do nothing
			case InlineMessageUID:
				inlineMessageID = m.EditMessageUID.(InlineMessageUID).InlineMessageID
				chatID = 0
				messageID = 0
			case *InlineMessageUID:
				inlineMessageID = m.EditMessageUID.(*InlineMessageUID).InlineMessageID
				chatID = 0
				messageID = 0
			case ChatMessageUID:
				chatMessageUID := m.EditMessageUID.(ChatMessageUID)
				inlineMessageID = ""
				if chatMessageUID.ChatID != 0 {
					chatID = chatMessageUID.ChatID
				}
				if chatMessageUID.MessageID != 0 {
					messageID = chatMessageUID.MessageID
				}
			case *ChatMessageUID:
				chatMessageUID := m.EditMessageUID.(*ChatMessageUID)
				inlineMessageID = ""
				if chatMessageUID.ChatID != 0 {
					chatID = chatMessageUID.ChatID
				}
				if chatMessageUID.MessageID != 0 {
					messageID = chatMessageUID.MessageID
				}
			default:
				err = fmt.Errorf("unknown EditMessageUID type %T(%v)", m.EditMessageUID, m.EditMessageUID)
				return
			}
		}
		log.Debugf(c, "Edit message => inlineMessageID: %v, chatID: %d, messageID: %d", inlineMessageID, chatID, messageID)
		if inlineMessageID == "" && chatID == 0 && messageID == 0 {
			err = errors.New("Can't edit Telegram message as inlineMessageID is empty && chatID == 0 && messageID == 0")
			return
		}
		if m.Text == "" && m.Keyboard != nil {
			chattable = tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, inlineMessageID, m.Keyboard.(*tgbotapi.InlineKeyboardMarkup))
		} else if m.Text != "" {
			editMessageTextConfig := tgbotapi.NewEditMessageText(chatID, messageID, inlineMessageID, m.Text)
			editMessageTextConfig.ParseMode = parseMode()
			editMessageTextConfig.DisableWebPagePreview = m.DisableWebPagePreview
			if m.Keyboard != nil {
				editMessageTextConfig.ReplyMarkup = m.Keyboard.(*tgbotapi.InlineKeyboardMarkup)
			}
			chattable = editMessageTextConfig
		} else {
			err = fmt.Errorf("can't edit telegram message as got unknown output: %v", m)
			return
		}
	} else if m.Text != "" {
		messageConfig := r.whc.NewTgMessage(m.Text)
		if m.ToChat != nil {
			messageConfig.ChatID = int64(m.ToChat.(bots.ChatIntID))
		}
		messageConfig.DisableWebPagePreview = m.DisableWebPagePreview
		messageConfig.DisableNotification = m.DisableNotification
		if m.Keyboard != nil {
			messageConfig.ReplyMarkup = m.Keyboard
		}

		messageConfig.ParseMode = parseMode()

		chattable = messageConfig
	} else {
		switch r.whc.InputType() {
		case bots.WebhookInputInlineQuery: // pass
			log.Debugf(c, "No response to WebhookInputInlineQuery")
		case bots.WebhookInputChosenInlineResult: // pass
		default:
			mBytes, err := ffjson.Marshal(m)
			if err != nil {
				log.Errorf(c, "Failed to marshal MessageFromBot to JSON: %v", err)
			}
			inputTypeName := bots.WebhookInputTypeNames[r.whc.InputType()]
			log.Debugf(c, "Not inline answer, Not inline, Not edit inline, Text is empty. r.whc.InputType(): %v\nMessageFromBot:\n%v", inputTypeName, string(mBytes))
			ffjson.Pool(mBytes)
		}
		return
	}

	jsonStr, err := ffjson.Marshal(chattable)
	if err != nil {
		log.Errorf(c, "Failed to marshal message config to json: %v\n\tJSON: %v\n\tchattable: %v", err, jsonStr, chattable)
		ffjson.Pool(jsonStr)
		return resp, err
	}
	var indentedJSON bytes.Buffer
	var indentedJSONStr string
	if indentedErr := json.Indent(&indentedJSON, jsonStr, "", "\t"); indentedErr == nil {
		indentedJSONStr = indentedJSON.String()
	} else {
		indentedJSONStr = string(jsonStr)
	}
	ffjson.Pool(jsonStr)
	log.Debugf(c, "Sending to Telegram, Text: %v\n------------------------\nAs JSON: %v", m.Text, indentedJSONStr)

	//if values, err := chattable.Values(); err != nil {
	//	log.Errorf(c, "Failed to marshal message config to url.Values: %v", err)
	//	return resp, err
	//} else {
	//	log.Debugf(c, "Message for sending to Telegram as URL values: %v", values)
	//}

	switch channel {
	case bots.BotAPISendMessageOverResponse:
		if _, err := tgbotapi.ReplyToResponse(chattable, r.w); err != nil {
			log.Errorf(c, "Failed to send message to Telegram throw HTTP response: %v", err)
		}
		return resp, err
	case bots.BotAPISendMessageOverHTTPS:
		botAPI := tgbotapi.NewBotAPIWithClient(
			r.whc.BotContext.BotSettings.Token,
			r.whc.BotContext.BotHost.GetHTTPClient(c),
		)
		botAPI.EnableDebug(c)
		message, err := botAPI.Send(chattable)
		if err != nil {
			return resp, err
		} else if message.MessageID != 0 {
			log.Debugf(c, "Telegram API: MessageID=%v", message.MessageID)
		} else {
			messageJSON, err := ffjson.Marshal(message)
			if err != nil {
				log.Warningf(c, "Telegram API response as raw: %v", message)
			} else {
				log.Debugf(c, "Telegram API response as JSON: %v", string(messageJSON))
			}
			ffjson.Pool(messageJSON)
		}
		return bots.OnMessageSentResponse{TelegramMessage: message}, nil
	default:
		panic(fmt.Sprintf("Unknown channel: %v", channel))
	}
}
