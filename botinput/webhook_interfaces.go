package botinput

import (
	"fmt"
	"strconv"
	"time"
)

// WebhookEntry represents a single message from a messenger user
type WebhookEntry interface {
	GetID() interface{}
	GetTime() time.Time
}

// WebhookInputType is enum of input type
type WebhookInputType int

const (
	// WebhookInputUnknown is an unknown input type
	WebhookInputUnknown WebhookInputType = iota
	// WebhookInputNotImplemented is not implemented input type
	WebhookInputNotImplemented
	// WebhookInputText is a text input type
	WebhookInputText // Facebook, Telegram, Viber
	// WebhookInputVoice is voice input type
	WebhookInputVoice
	// WebhookInputPhoto is a photo input type
	WebhookInputPhoto
	// WebhookInputAudio is an audio input type
	WebhookInputAudio
	// WebhookInputContact is a contact input type
	WebhookInputContact // Facebook, Telegram, Viber
	// WebhookInputPostback is unknown input type
	WebhookInputPostback
	// WebhookInputDelivery is a postback input type
	WebhookInputDelivery
	// WebhookInputAttachment is a delivery report input type
	WebhookInputAttachment
	// WebhookInputInlineQuery is an attachment input type
	WebhookInputInlineQuery // Telegram
	// WebhookInputCallbackQuery is inline input type
	WebhookInputCallbackQuery
	// WebhookInputReferral is a callback input type
	WebhookInputReferral // FBM
	// WebhookInputChosenInlineResult is chosen inline result input type
	WebhookInputChosenInlineResult // Telegram
	// WebhookInputSubscribed is subscribed input type
	WebhookInputSubscribed // Viber
	// WebhookInputUnsubscribed is unsubscribed input type
	WebhookInputUnsubscribed // Viber
	// WebhookInputConversationStarted is conversation started input type
	WebhookInputConversationStarted // Viber
	// WebhookInputNewChatMembers is new botChat members input type
	WebhookInputNewChatMembers // Telegram groups
	// WebhookInputLeftChatMembers is left botChat members input type
	WebhookInputLeftChatMembers
	// WebhookInputSticker is a sticker input type
	WebhookInputSticker
	WebhookInputSharedUsers // Telegram
)

var webhookInputTypeNames = map[WebhookInputType]string{
	WebhookInputUnknown:             "unknown",
	WebhookInputNotImplemented:      "NotImplemented",
	WebhookInputText:                "Text",
	WebhookInputVoice:               "Voice",
	WebhookInputPhoto:               "Photo",
	WebhookInputAudio:               "Audio",
	WebhookInputReferral:            "Referral",
	WebhookInputContact:             "Contact",
	WebhookInputPostback:            "Postback",
	WebhookInputDelivery:            "Delivery",
	WebhookInputAttachment:          "Attachment",
	WebhookInputInlineQuery:         "InlineQuery",
	WebhookInputCallbackQuery:       "CallbackQuery",
	WebhookInputChosenInlineResult:  "ChosenInlineResult",
	WebhookInputSubscribed:          "Subscribed",          // Viber
	WebhookInputUnsubscribed:        "Unsubscribed",        // Viber
	WebhookInputConversationStarted: "ConversationStarted", // Telegram
	WebhookInputNewChatMembers:      "NewChatMembers",      // Telegram
	WebhookInputSticker:             "Sticker",             // Telegram
	WebhookInputLeftChatMembers:     "LeftChatMembers",     // Telegram
	WebhookInputSharedUsers:         "SharedUsers",         // Telegram
}

func GetWebhookInputTypeIdNameString(whInputType WebhookInputType) string {
	name, ok := webhookInputTypeNames[whInputType]
	if ok {
		return fmt.Sprintf("%d:%s", whInputType, name)
	}
	return strconv.Itoa(int(whInputType))
}

// WebhookInput represent a single message
// '/entry/messaging' for Facebook Messenger
type WebhookInput interface {
	GetSender() WebhookUser
	GetRecipient() WebhookRecipient
	GetTime() time.Time
	InputType() WebhookInputType
	BotChatID() (string, error)
	Chat() WebhookChat
	LogRequest() // TODO: should not be part of Input? If should - specify why
}

// WebhookActor represents sender
type WebhookActor interface {
	Platform() string // TODO: Consider removing this?
	GetID() any
	IsBotUser() bool
	GetFirstName() string
	GetLastName() string
	GetUserName() string
	GetLanguage() string
}

// WebhookSender represents sender with avatar
type WebhookSender interface {
	GetAvatar() string // Extension to support avatar (Viber)
	WebhookActor
}

// WebhookUser represents sender with country
type WebhookUser interface {
	WebhookSender

	// GetCountry is an extension to support language & country (Viber)
	GetCountry() string
}

// WebhookRecipient represents receiver
type WebhookRecipient interface {
	WebhookActor
}

// WebhookMessage represents single message
type WebhookMessage interface {
	IntID() int64
	StringID() string
	Chat() WebhookChat
	//Sequence() int // 'seq' for Facebook, '???' for Telegram
}

// WebhookTextMessage represents a single text message
type WebhookTextMessage interface {
	WebhookMessage
	Text() string
	IsEdited() bool
}

// WebhookStickerMessage represents single sticker message
type WebhookStickerMessage interface {
	WebhookMessage
	// TODO: Define sticker message interface
}

// WebhookVoiceMessage represents a single voice message
type WebhookVoiceMessage interface {
	WebhookMessage
	// TODO: Define voice message interface
}

// WebhookPhotoMessage represents a single photo message
type WebhookPhotoMessage interface {
	WebhookMessage
	// TODO: Define photo message interface
}

// WebhookAudioMessage represents a single audio message
type WebhookAudioMessage interface {
	WebhookMessage
	// TODO: Define audio message interface
}

// WebhookReferralMessage represents a single referral message
// https://developers.facebook.com/docs/messenger-platform/webhook-reference/referral
type WebhookReferralMessage interface {
	Type() string
	Source() string
	RefData() string
}

// WebhookContactMessage represents a single contact message
type WebhookContactMessage interface {
	GetPhoneNumber() string
	GetFirstName() string
	GetLastName() string
	GetBotUserID() string
	GetVCard() string
}

// WebhookNewChatMembersMessage represents a single message about a new member of a botChat
type WebhookNewChatMembersMessage interface {
	BotChatID() (string, error)
	NewChatMembers() []WebhookActor
}

// WebhookLeftChatMembersMessage represents a single message about a member leaving a botChat
type WebhookLeftChatMembersMessage interface {
	BotChatID() (string, error)
	LeftChatMembers() []WebhookActor
}

// WebhookChat represents botChat of a messenger
type WebhookChat interface {
	GetID() string
	GetType() string
	IsGroupChat() bool
}

// WebhookPostback represents a single postback message
type WebhookPostback interface {
	PostbackMessage() interface{}
	Payload() string
}

// WebhookSubscribed represents a subscription message
type WebhookSubscribed interface {
	SubscribedMessage() interface{}
}

// WebhookUnsubscribed represents a message when user unsubscribe
type WebhookUnsubscribed interface {
	UnsubscribedMessage() interface{}
}

// WebhookConversationStarted represents a single message about new conversation
type WebhookConversationStarted interface {
	ConversationStartedMessage() interface{}
}

// WebhookInlineQuery represents a single inline message
type WebhookInlineQuery interface {
	GetID() interface{}
	GetInlineQueryID() string
	GetFrom() WebhookSender
	GetQuery() string
	GetOffset() string
	//GetLocation() - TODO: Not implemented yet
}

// WebhookDelivery represents a single delivery report message
type WebhookDelivery interface {
	Payload() string
}

// WebhookChosenInlineResult represents a single report message on chosen inline result
type WebhookChosenInlineResult interface {
	GetResultID() string
	GetInlineMessageID() string // Telegram only?
	GetFrom() WebhookSender
	GetQuery() string
	//GetLocation() - TODO: Not implemented yet
}

// WebhookCallbackQuery represents a single callback query message
type WebhookCallbackQuery interface {
	GetID() string
	GetInlineMessageID() string // Telegram only?
	GetFrom() WebhookSender
	GetMessage() WebhookMessage
	GetData() string
	Chat() WebhookChat
}

// WebhookAttachment represents attachment to a message
type WebhookAttachment interface {
	Type() string       // Enum(image, video, audio) for Facebook
	PayloadUrl() string // 'payload.url' for Facebook
}
