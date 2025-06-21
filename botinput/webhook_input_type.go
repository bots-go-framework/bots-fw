package botinput

//go:generate stringer -type=WebhookInputType

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
	WebhookInputPreCheckoutQuery
	WebhookInputSuccessfulPayment
	WebhookInputRefundedPayment
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
	WebhookInputSuccessfulPayment:   "SuccessfulPayment",   // Telegram
	WebhookInputRefundedPayment:     "RefundedPayment",     // Telegram
}
