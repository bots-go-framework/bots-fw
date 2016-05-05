package telegram_bot

import "github.com/strongo/bots-framework/core"

type TelegramPlatform struct {
}

var _ bots.BotPlatform = (*TelegramPlatform)(nil)

func (p TelegramPlatform) Id() string {
	return "telegram"
}

func (p TelegramPlatform) Version() string {
	return "2.0"
}
