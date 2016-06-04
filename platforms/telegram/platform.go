package telegram_bot

import "github.com/strongo/bots-framework/core"

type TelegramPlatform struct {
}

var _ bots.BotPlatform = (*TelegramPlatform)(nil)

const TelegramPlatformID = "telegram"

func (p TelegramPlatform) Id() string {
	return TelegramPlatformID
}

func (p TelegramPlatform) Version() string {
	return "2.0"
}
