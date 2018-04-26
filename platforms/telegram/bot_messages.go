package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

// InlineBotMessage is wrapper for Telegram bot message
type InlineBotMessage tgbotapi.InlineConfig

// BotMessageType returns BotMessageTypeInlineResults
func (InlineBotMessage) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeInlineResults
}

// CallbackAnswer is callback answer message
type CallbackAnswer tgbotapi.AnswerCallbackQueryConfig

// BotMessageType returns BotMessageTypeCallbackAnswer
func (CallbackAnswer) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeCallbackAnswer
}

// LeaveChat is leave chat message from bot
type LeaveChat tgbotapi.LeaveChatConfig

// BotMessageType return BotMessageTypeLeaveChat
func (LeaveChat) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeLeaveChat
}

// ExportChatInviteLink is TG message
type ExportChatInviteLink tgbotapi.ExportChatInviteLink

// BotMessageType returns BotMessageTypeExportChatInviteLink
func (ExportChatInviteLink) BotMessageType() bots.BotMessageType {
	return bots.BotMessageTypeExportChatInviteLink
}
