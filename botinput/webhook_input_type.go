package botinput

//go:generate stringer -type=WebhookInputType

// Type is enum of input type
type Type int

const (
	// TypeUnknown is an unknown input type
	TypeUnknown Type = iota
	// TypeNotImplemented is not implemented input type
	TypeNotImplemented
	// TypeText is a text input type
	TypeText // Facebook, Telegram, Viber
	// TypeVoice is voice input type
	TypeVoice
	// TypePhoto is a photo input type
	TypePhoto
	// TypeAudio is an audio input type
	TypeAudio
	// TypeContact is a contact input type
	TypeContact // Facebook, Telegram, Viber
	// TypePostback is unknown input type
	TypePostback
	// TypeDelivery is a postback input type
	TypeDelivery
	// TypeAttachment is a delivery report input type
	TypeAttachment
	// TypeInlineQuery is an attachment input type
	TypeInlineQuery // Telegram
	// TypeCallbackQuery is inline input type
	TypeCallbackQuery
	// TypeReferral is a callback input type
	TypeReferral // FBM
	// TypeChosenInlineResult is chosen inline result input type
	TypeChosenInlineResult // Telegram
	// TypeSubscribed is subscribed input type
	TypeSubscribed // Viber
	// TypeUnsubscribed is unsubscribed input type
	TypeUnsubscribed // Viber
	// TypeConversationStarted is conversation started input type
	TypeConversationStarted // Viber
	// TypeNewChatMembers is new botChat members input type
	TypeNewChatMembers // Telegram groups
	// TypeLeftChatMembers is left botChat members input type
	TypeLeftChatMembers
	// TypeSticker is a sticker input type
	TypeSticker
	TypeSharedUsers // Telegram
	TypePreCheckoutQuery
	TypeSuccessfulPayment
	TypeRefundedPayment
)

var webhookInputTypeNames = map[Type]string{
	TypeUnknown:             "unknown",
	TypeNotImplemented:      "NotImplemented",
	TypeText:                "Text",
	TypeVoice:               "Voice",
	TypePhoto:               "Photo",
	TypeAudio:               "Audio",
	TypeReferral:            "Referral",
	TypeContact:             "Contact",
	TypePostback:            "Postback",
	TypeDelivery:            "Delivery",
	TypeAttachment:          "Attachment",
	TypeInlineQuery:         "InlineQuery",
	TypeCallbackQuery:       "CallbackQuery",
	TypeChosenInlineResult:  "ChosenInlineResult",
	TypeSubscribed:          "Subscribed",          // Viber
	TypeUnsubscribed:        "Unsubscribed",        // Viber
	TypeConversationStarted: "ConversationStarted", // Telegram
	TypeNewChatMembers:      "NewChatMembers",      // Telegram
	TypeSticker:             "Sticker",             // Telegram
	TypeLeftChatMembers:     "LeftChatMembers",     // Telegram
	TypeSharedUsers:         "SharedUsers",         // Telegram
	TypeSuccessfulPayment:   "SuccessfulPayment",   // Telegram
	TypeRefundedPayment:     "RefundedPayment",     // Telegram
}
