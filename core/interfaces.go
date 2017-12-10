package bots

import (
	"github.com/strongo/db"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type BotPlatform interface {
	Id() string
	Version() string
}

type BotHost interface {
	Context(r *http.Request) context.Context
	GetHttpClient(c context.Context) *http.Client
	GetBotCoreStores(platform string, appContext BotAppContext, r *http.Request) BotCoreStores
	DB() db.Database
}

type BotContext struct {
	// TODO: Rename to BotWebhookContext or just WebhookContext (replace old one)
	BotHost     BotHost
	BotSettings BotSettings
}

func NewBotContext(host BotHost, settings BotSettings) *BotContext {
	if settings.Code == "" {
		panic("ReferredTo settings.Code is empty string")
	}
	return &BotContext{
		BotHost:     host,
		BotSettings: settings,
	}
}

type WebhookEntry interface {
	GetID() interface{}
	GetTime() time.Time
}

type WebhookInputType int

const (
	WebhookInputUnknown WebhookInputType = iota
	WebhookInputText                     // Facebook, Telegram, Viber
	WebhookInputVoice
	WebhookInputPhoto
	WebhookInputAudio
	WebhookInputContact // Facebook, Telegram, Viber
	WebhookInputPostback
	WebhookInputDelivery
	WebhookInputAttachment
	WebhookInputInlineQuery // Telegram
	WebhookInputCallbackQuery
	WebhookInputReferral            // FBM
	WebhookInputChosenInlineResult  // Telegram
	WebhookInputSubscribed          // Viber
	WebhookInputUnsubscribed        // Viber
	WebhookInputConversationStarted // Viber
	WebhookInputNewChatMembers      // Telegram groups
	WebhookInputLeftChatMembers
	WebhookInputSticker // Telegram

	WebhookInputNotImplemented //
)

var WebhookInputTypeNames = map[WebhookInputType]string{
	//WebhookInputContact:				  "Contact",
	WebhookInputUnknown:             "unknown",
	WebhookInputReferral:            "Referral",
	WebhookInputText:                "Text",
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
}

type WebhookInput interface {
	// '/entry/messaging' for Facebook
	GetSender() WebhookSender
	GetRecipient() WebhookRecipient
	GetTime() time.Time
	InputType() WebhookInputType
	BotChatID() (string, error)
	Chat() WebhookChat
	LogRequest()
}

type WebhookActor interface {
	Platform() string
	GetID() interface{}
	IsBotUser() bool
	GetFirstName() string
	GetLastName() string
	GetUserName() string
	GetLanguage() string
}

type WebhookSender interface {
	GetAvatar() string // Extension to support avatar (Viber)
	WebhookActor
}

type WebhookUser interface {
	// Extension to support language & country (Viber)
	GetCountry() string
	WebhookSender
}

type WebhookRecipient interface {
	WebhookActor
}

type WebhookMessage interface {
	IntID() int64
	StringID() string
	Chat() WebhookChat
	//Sequence() int // 'seq' for Facebook, '???' for Telegram
}

type WebhookTextMessage interface {
	WebhookMessage
	Text() string
	IsEdited() bool
}

type WebhookStickerMessage interface {
	WebhookMessage
	// TODO: Define sticker message interface
}

type WebhookVoiceMessage interface {
	WebhookMessage
	// TODO: Define voice message interface
}

type WebhookPhotoMessage interface {
	WebhookMessage
	// TODO: Define voice message interface
}

type WebhookAudioMessage interface {
	WebhookMessage
	// TODO: Define voice message interface
}

type WebhookReferralMessage interface {
	// https://developers.facebook.com/docs/messenger-platform/webhook-reference/referral
	Type() string
	Source() string
	RefData() string
}

type WebhookContactMessage interface {
	PhoneNumber() string
	FirstName() string
	LastName() string
	UserID() interface{}
}

type WebhookNewChatMembersMessage interface {
	BotChatID() (string, error)
	NewChatMembers() []WebhookActor
}

type WebhookLeftChatMembersMessage interface {
	BotChatID() (string, error)
	LeftChatMembers() []WebhookActor
}

type WebhookChat interface {
	GetID() string
	GetType() string
	IsGroupChat() bool
}

type WebhookPostback interface {
	PostbackMessage() interface{}
	Payload() string
}

type WebhookSubscribed interface {
	SubscribedMessage() interface{}
}

type WebhookUnsubscribed interface {
	UnsubscribedMessage() interface{}
}

type WebhookConversationStarted interface {
	ConversationStartedMessage() interface{}
}

type WebhookInlineQuery interface {
	GetID() interface{}
	GetInlineQueryID() string
	GetFrom() WebhookSender
	GetQuery() string
	GetOffset() string
	//GetLocation() - TODO: Not implemented yet
}

type WebhookDelivery interface {
	Payload() string
}

type WebhookChosenInlineResult interface {
	GetResultID() string
	GetInlineMessageID() string // Telegram only?
	GetFrom() WebhookSender
	GetQuery() string
	//GetLocation() - TODO: Not implemented yet
}

type WebhookCallbackQuery interface {
	GetID() interface{}
	GetInlineMessageID() string // Telegram only?
	GetFrom() WebhookSender
	GetMessage() WebhookMessage
	GetData() string
	Chat() WebhookChat
}

type WebhookAttachment interface {
	Type() string       // Enum(image, video, audio) for Facebook
	PayloadUrl() string // 'payload.url' for Facebook
}

type MessengerResponse interface {
}

type OnMessageSentResponse struct {
	StatusCode      int
	TelegramMessage MessengerResponse // TODO: change to some interface
}

type WebhookResponder interface {
	SendMessage(c context.Context, m MessageFromBot, channel BotApiSendMessageChannel) (OnMessageSentResponse, error)
}

type InputMessage interface {
	Text() string
}

type BotCoreStores struct {
	BotChatStore
	BotUserStore
	BotAppUserStore
}

type BotApiSendMessageChannel string

const (
	BotApiSendMessageOverHTTPS    = BotApiSendMessageChannel("https")
	BotApiSendMessageOverResponse = BotApiSendMessageChannel("response")
)
