package botmsg

import "errors"

var _ BotMessage = ChatAction{}
var _ BotMessage = (*ChatAction)(nil)

// Chat action values, per https://core.telegram.org/bots/api#sendchataction.
const (
	// ChatActionTyping indicates the bot is typing a text message.
	ChatActionTyping = "typing"
	// ChatActionUploadPhoto indicates the bot is uploading a photo.
	ChatActionUploadPhoto = "upload_photo"
	// ChatActionRecordVideo indicates the bot is recording a video.
	ChatActionRecordVideo = "record_video"
	// ChatActionUploadVideo indicates the bot is uploading a video.
	ChatActionUploadVideo = "upload_video"
	// ChatActionRecordVoice indicates the bot is recording a voice message.
	ChatActionRecordVoice = "record_voice"
	// ChatActionUploadVoice indicates the bot is uploading a voice message.
	ChatActionUploadVoice = "upload_voice"
	// ChatActionUploadDocument indicates the bot is uploading a document.
	ChatActionUploadDocument = "upload_document"
	// ChatActionChooseSticker indicates the bot is choosing a sticker.
	ChatActionChooseSticker = "choose_sticker"
	// ChatActionFindLocation indicates the bot is finding a location.
	ChatActionFindLocation = "find_location"
)

// ChatAction tells the messenger to show a chat action indicator (e.g. "typing")
// in the target chat for a few seconds. Modeled after
// https://core.telegram.org/bots/api#sendchataction.
//
// It carries no text; send it as MessageFromBot{BotMessage: ChatAction{...}} via
// the responder. When ChatID is empty the platform adapter targets the current chat.
type ChatAction struct {
	// ChatID is the target chat. When empty, the adapter uses the current chat.
	ChatID string `json:"chat_id,omitempty"`
	// Action is the chat action to show, e.g. ChatActionTyping. Required.
	Action string `json:"action"`
}

// BotMessageType implements BotMessage.
func (ChatAction) BotMessageType() Type {
	return TypeChatAction
}

// Validate reports whether the chat action is well-formed.
func (v ChatAction) Validate() error {
	if v.Action == "" {
		return errors.New("missing required parameter Action")
	}
	return nil
}
