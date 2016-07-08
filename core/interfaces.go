package bots

import (
	"net/http"
	"time"
)

type BotPlatform interface {
	Id() string
	Version() string
}

func UtmSource(p BotPlatform) string {
	return "bot-" + p.Id()
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Criticalf(format string, args ...interface{})
}

type BotHost interface {
	GetLogger(r *http.Request) Logger
	GetHttpClient(r *http.Request) *http.Client
	GetBotCoreStores(platform string, appContext AppContext, r *http.Request) BotCoreStores
}

type BotContext struct { // TODO: Rename to BotWebhookContext or just WebhookContext (replace old one)
	BotHost     BotHost
	BotSettings BotSettings
}

type WebhookHandler interface {
	RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request))
	HandleWebhookRequest(w http.ResponseWriter, r *http.Request)
	GetBotContextAndInputs(r *http.Request) (botContext BotContext, entriesWithInputs []EntryInputs, err error)
	CreateBotCoreStores(appContext AppContext, r *http.Request) BotCoreStores
	CreateWebhookContext(appContext AppContext, r *http.Request, botContext BotContext, webhookInput WebhookInput, botCoreStores BotCoreStores) WebhookContext //TODO: Can we get rid of http.Request? Needed for botHost.GetHttpClient()
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
	WebhookInputMessage
	WebhookInputPostback
	WebhookInputDelivery
	WebhookInputAttachment
	//WebhookInputContact
	WebhookInputInlineQuery // Telegram only?
	WebhookInputCallbackQuery
	WebhookInputChosenInlineResult // Telegram only?
)

var WebhookInputTypeNames = map[WebhookInputType]string{
	//WebhookInputContact:				  "Contact",
	WebhookInputUnknown:            "unknown",
	WebhookInputMessage:            "Message",
	WebhookInputPostback:           "Postback",
	WebhookInputDelivery:           "Delivery",
	WebhookInputAttachment:         "Attachment",
	WebhookInputInlineQuery:        "InlineQuery",
	WebhookInputCallbackQuery:      "CallbackQuery",
	WebhookInputChosenInlineResult: "ChosenInlineResult",
}

type WebhookInput interface { // '/entry/messaging' for Facebook
	GetSender() WebhookSender
	GetRecipient() WebhookRecipient
	GetTime() time.Time

	InputType() WebhookInputType

	InputMessage() WebhookMessage
	InputPostback() WebhookPostback
	InputDelivery() WebhookDelivery
	InputInlineQuery() WebhookInlineQuery
	InputCallbackQuery() WebhookCallbackQuery
	InputChosenInlineResult() WebhookChosenInlineResult
}

type WebhookActor interface {
	GetID() interface{}
	GetFirstName() string
	GetLastName() string
	GetUserName() string
	Platform() string
}

type WebhookSender interface {
	WebhookActor
}

type WebhookRecipient interface {
	WebhookActor
}

type WebhookMessage interface {
	IntID() int64
	StringID() string
	Sequence() int // 'seq' for Facebook, '???' for Telegram
	Text() string
	Chat() WebhookChat
	Contact() WebhookContact
}

type WebhookContact interface {
	PhoneNumber() string
	FirstName() string
	LastName() string
	UserID() interface{}
}

type WebhookChat interface {
	GetID() interface{}
	GetTitle() string
	GetType() string
}

type WebhookPostback interface {
	Payload() string
}

type WebhookDelivery interface {
	Payload() string
}

type WebhookInlineQuery interface {
	GetID() interface{}
	GetInlineQueryID() string
	GetFrom() WebhookSender
	GetQuery() string
	GetOffset() string
	//GetLocation() - TODO: Not implemented yet
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

type WebhookResponder interface {
	SendMessage(m MessageFromBot, channel BotApiSendMessageChannel) error
}

type InputMessage interface {
	Text() string
}

type BotCoreStores struct {
	BotChatStore
	BotUserStore
	AppUserStore
}

type BotApiSendMessageChannel string

const (
	BotApiSendMessageOverHTTPS = BotApiSendMessageChannel("https")
	BotApiSendMessageOverResponse = BotApiSendMessageChannel("response")
)