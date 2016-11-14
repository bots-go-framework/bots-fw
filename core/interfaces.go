package bots

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type BotPlatform interface {
	Id() string
	Version() string
}

func UtmSource(p BotPlatform) string {
	return p.Id()
}

//type strongo.Logger interface {
//	Debugf(format string, args ...interface{})
//	Infof(format string, args ...interface{})
//	Warningf(format string, args ...interface{})
//	Errorf(format string, args ...interface{})
//	Criticalf(format string, args ...interface{})
//}

type BotHost interface {
	Logger(r *http.Request) strongo.Logger
	Context(r *http.Request) context.Context
	GetHttpClient(r *http.Request) *http.Client
	GetBotCoreStores(platform string, appContext BotAppContext, r *http.Request) BotCoreStores
}

type BotContext struct { // TODO: Rename to BotWebhookContext or just WebhookContext (replace old one)
	BotHost     BotHost
	BotSettings BotSettings
}

type WebhookHandler interface {
	RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request))
	HandleWebhookRequest(w http.ResponseWriter, r *http.Request)
	GetBotContextAndInputs(r *http.Request) (botContext *BotContext, entriesWithInputs []EntryInputs, err error)
	CreateBotCoreStores(appContext BotAppContext, r *http.Request) BotCoreStores
	CreateWebhookContext(appContext BotAppContext, r *http.Request, botContext BotContext, webhookInput WebhookInput, botCoreStores BotCoreStores, gaMeasurement *measurement.BufferedSender) WebhookContext //TODO: Can we get rid of http.Request? Needed for botHost.GetHttpClient()
	GetResponder(w http.ResponseWriter, whc WebhookContext) WebhookResponder
	//ProcessInput(input WebhookInput, entry *WebhookEntry)
}

type WebhookEntry interface {
	GetID() interface{}
	GetTime() time.Time
}

type WebhookInputType int

const (
	WebhookInputUnknown WebhookInputType = iota
	WebhookInputText              // Facebook, Telegram, Viber
	WebhookInputContact              // Facebook, Telegram, Viber
	WebhookInputPostback
	WebhookInputDelivery
	WebhookInputAttachment
	WebhookInputInlineQuery          // Telegram
	WebhookInputCallbackQuery
	WebhookInputChosenInlineResult   // Telegram
	WebhookInputSubscribed           // Viber
	WebhookInputUnsubscribed         // Viber
	WebhookInputConversationStarted  // Viber
)

var WebhookInputTypeNames = map[WebhookInputType]string{
	//WebhookInputContact:				  "Contact",
	WebhookInputUnknown:            "unknown",
	WebhookInputText:               "Text",
	WebhookInputContact:            "Contact",
	WebhookInputPostback:           "Postback",
	WebhookInputDelivery:           "Delivery",
	WebhookInputAttachment:         "Attachment",
	WebhookInputInlineQuery:        "InlineQuery",
	WebhookInputCallbackQuery:      "CallbackQuery",
	WebhookInputChosenInlineResult: "ChosenInlineResult",
	WebhookInputSubscribed:          "Subscribed",  // Viber
	WebhookInputUnsubscribed:        "Unsubscribed", // Viber
	WebhookInputConversationStarted: "ConversationStarted",
}

type WebhookInput interface { // '/entry/messaging' for Facebook
	GetSender() WebhookSender
	GetRecipient() WebhookRecipient
	GetTime() time.Time
	InputType() WebhookInputType
	Chat() WebhookChat
}

type WebhookActor interface {
	GetID() interface{}
	GetFirstName() string
	GetLastName() string
	GetUserName() string
	Platform() string
}

type WebhookSender interface { // TODO: Extend to support avatar & language (Viber)
	//GetAvatar() string
	//GetLanguage() string
	WebhookActor
}

type WebhookRecipient interface {
	WebhookActor
}

type WebhookMessage interface {
	IntID() int64
	StringID() string
	//Sequence() int // 'seq' for Facebook, '???' for Telegram
}

type WebhookTextMessage interface {
	WebhookMessage
	Text() string
}

type WebhookContactMessage interface {
	PhoneNumber() string
	FirstName() string
	LastName() string
	UserID() interface{}
}

type WebhookChat interface {
	GetID() interface{}
	GetFullName() string
	GetType() string
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

type OnMessageSentResponse struct {
	TelegramMessage tgbotapi.Message
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
