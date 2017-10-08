package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type InlineBotMessage tgbotapi.InlineConfig

func (InlineBotMessage) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeInlineResults
}

type CallbackAnswer tgbotapi.AnswerCallbackQueryConfig

func (CallbackAnswer) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeCallbackAnswer
}

type LeaveChat tgbotapi.LeaveChatConfig

func (LeaveChat) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeLeaveChat
}

type ExportChatInviteLink tgbotapi.ExportChatInviteLink

func (ExportChatInviteLink) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeExportChatInviteLink
}
