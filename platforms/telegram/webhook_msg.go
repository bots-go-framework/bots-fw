package telegram_bot

type TelegramMessageType string
const (
	TelegramMessageTypeRegular = "message"
	TelegramMessageTypeEdited = "edited_message"
	TelegramMessageTypeChannelPost = "channel_post"
	TelegramMessageTypeEditedChannelPost = "edited_channel_post"
)
