package telegram_bot

type TelegramPlatform struct {
}

func (p TelegramPlatform) Id() string {
	return "telegram"
}

func (p TelegramPlatform) Version() string {
	return "2.0"
}