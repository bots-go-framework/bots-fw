package botinput

import (
	"fmt"
	"strconv"
	"time"
)

// Entry represents a single message from a messenger user
type Entry interface {
	GetID() any
	GetTime() time.Time
}

func GetBotInputTypeIdNameString(whInputType Type) string {
	name, ok := webhookInputTypeNames[whInputType]
	if ok {
		return fmt.Sprintf("%d:%s", whInputType, name)
	}
	return strconv.Itoa(int(whInputType))
}

// InputMessage represent a single message
// '/entry/messaging' for Facebook Messenger
type InputMessage interface {
	GetSender() User
	GetRecipient() Recipient
	GetTime() time.Time
	InputType() Type
	BotChatID() (string, error)
	Chat() Chat
	LogRequest() // TODO: should not be part of Input? If should - specify why
}

// Actor represents sender
type Actor interface {
	Platform() string // TODO: Consider removing this?
	GetID() any
	IsBotUser() bool
	GetFirstName() string
	GetLastName() string
	GetUserName() string
	GetLanguage() string
}

// Sender represents sender with avatar
type Sender interface {
	GetAvatar() string // Extension to support avatar (Viber)
	Actor
}

// User represents sender with country
type User interface {
	Sender

	// GetCountry is an extension to support language & country (Viber)
	GetCountry() string
}

// Recipient represents receiver
type Recipient interface {
	Actor
}

// Message represents a single input  message
type Message interface {
	InputMessage
	IntID() int64
	StringID() string
	Chat() Chat
	//Sequence() int // 'seq' for Facebook, '???' for Telegram
}

// TextMessage represents a single text message
type TextMessage interface {
	Message
	Text() string
	IsEdited() bool
}

// StickerMessage represents single sticker message
type StickerMessage interface {
	Message
	// TODO: Define sticker message interface
}

// VoiceMessage represents a single voice message
type VoiceMessage interface {
	Message
	// TODO: Define voice message interface
}

// PhotoMessage represents a single photo message
type PhotoMessage interface {
	Message
	// TODO: Define photo message interface
}

// AudioMessage represents a single audio message
type AudioMessage interface {
	Message
	// TODO: Define audio message interface
}

// ReferralMessage represents a single referral message
// https://developers.facebook.com/docs/messenger-platform/webhook-reference/referral
type ReferralMessage interface {
	Type() string
	Source() string
	RefData() string
}

// ContactMessage represents a single contact message
type ContactMessage interface {
	GetPhoneNumber() string
	GetFirstName() string
	GetLastName() string
	GetBotUserID() string
	GetVCard() string
}

// NewChatMembersMessage represents a single message about a new member of a botChat
type NewChatMembersMessage interface {
	BotChatID() (string, error)
	NewChatMembers() []Actor
}

// LeftChatMembersMessage represents a single message about a member leaving a botChat
type LeftChatMembersMessage interface {
	BotChatID() (string, error)
	LeftChatMembers() []Actor
}

// Chat represents botChat of a messenger
type Chat interface {
	GetID() string
	GetType() string
	IsGroupChat() bool
}

// Postback represents a single postback message
type Postback interface {
	PostbackMessage() any
	Payload() string
}

// Subscribed represents a subscription message
type Subscribed interface {
	SubscribedMessage() any
}

// Unsubscribed represents a message when user unsubscribe
type Unsubscribed interface {
	UnsubscribedMessage() any
}

// ConversationStarted represents a single message about new conversation
type ConversationStarted interface {
	ConversationStartedMessage() any
}

// InlineQuery represents a single inline message
type InlineQuery interface {
	GetID() any
	GetInlineQueryID() string
	GetFrom() Sender
	GetQuery() string
	GetOffset() string
	//GetLocation() - TODO: Not implemented yet
}

// Delivery represents a single delivery report message
type Delivery interface {
	Payload() string
}

// ChosenInlineResult represents a single report message on chosen inline result
type ChosenInlineResult interface {
	GetResultID() string
	GetInlineMessageID() string // Telegram only?
	GetFrom() Sender
	GetQuery() string
	//GetLocation() - TODO: Not implemented yet
}

// CallbackQuery represents a single callback query message
type CallbackQuery interface {
	GetID() string
	//GetInlineMessageID() string // Telegram only?
	//GetChatInstanceID() string  // Telegram only?
	GetFrom() Sender
	GetMessage() Message
	GetData() string
	Chat() Chat
}

//type SuccessfulPayment struct {
//	Currency                string
//	TotalAmount             int
//	Payload                 string
//	IsRecurring             bool
//	IsFirstRecurring        bool
//	MessengerChargeID       string
//	PaymentProviderChargeID string
//	//
//	SubscriptionExpirationDate *time.Time
//}

type webhookPayment interface {
	GetCurrency() string
	GetTotalAmount() int
	GetInvoicePayload() string
	GetMessengerChargeID() string
	GetPaymentProviderChargeID() string
}
type SuccessfulPayment interface {
	webhookPayment
	GetSubscriptionExpirationDate() time.Time
	GetIsRecurring() bool
	GetIsFirstRecurring() bool
	GetShippingOptionID() string
	GetOrderInfo() OrderInfo
}

type WebhookRefundedPayment interface {
	webhookPayment
}

type WebhookPreCheckoutQuery interface {
	GetPreCheckoutQueryID() string
	GetCurrency() string
	GetTotalAmount() int
	GetInvoicePayload() string
	GetFrom() Sender
	GetShippingOptionID() string
	GetOrderInfo() OrderInfo
}

type OrderInfo interface {
	GetUserName() string //
	GetPhoneNumber() string
	GetEmailAddress() string
	GetShippingAddress() ShippingAddress
}

type ShippingAddress interface {
	GetCountryCode() string // Two-letter ISO 3166-1 alpha-2 country code
	GetState() string       // if applicable
	GetCity() string
	GetStreetLine1() string // First line for the address
	GetStreetLine2() string // Second line for the address
	GetPostCode() string
}

// WebhookAttachment represents attachment to a message
type WebhookAttachment interface {
	Type() string       // Enum(image, video, audio) for Facebook
	PayloadUrl() string // 'payload.url' for Facebook
}
