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

// WebhookMessage represents a single message
type WebhookMessage interface {
	WebhookInput
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
	//GetInlineMessageID() string // Telegram only?
	//GetChatInstanceID() string  // Telegram only?
	GetFrom() WebhookSender
	GetMessage() WebhookMessage
	GetData() string
	Chat() WebhookChat
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
type WebhookSuccessfulPayment interface {
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
	GetFrom() WebhookSender
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
