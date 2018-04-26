package telegram

// TgMessageType represents tpye of Telegram message
type TgMessageType string

const (
	// TgMessageTypeRegular is 'message'
	TgMessageTypeRegular = "message"

	// TgMessageTypeEdited is 'edited_message'
	TgMessageTypeEdited = "edited_message"

	// TgMessageTypeChannelPost is 'channel_post'
	TgMessageTypeChannelPost = "channel_post"

	// TgMessageTypeEditedChannelPost is 'edited_channel_post'
	TgMessageTypeEditedChannelPost = "edited_channel_post"
)
