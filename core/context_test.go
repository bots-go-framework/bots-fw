package bots

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"net/http"
)

type TestWebhookContext struct {
}

func (whc TestWebhookContext) MessageText() string {
	return "test message"
}

func (whc TestWebhookContext) Translate(key string) string {
	return key
}

func (whc TestWebhookContext) TranslateNoWarning(key string) string {
	return key
}

func (whc TestWebhookContext) NewChatEntity() BotChat  { panic("Not implemented") }
func (whc TestWebhookContext) MakeChatEntity() BotChat { panic("Not implemented") }

func (whc TestWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	panic("Not implemented")
}
func (whc TestWebhookContext) Context() context.Context { panic("Not implemented") }

func (whc TestWebhookContext) ChatKey() *datastore.Key                     { panic("Not implemented") }
func (whc TestWebhookContext) NewChatKey(c context.Context) *datastore.Key { panic("Not implemented") }
func (whc TestWebhookContext) ChatEntity() BotChat                         { panic("Not implemented") }

func (whc TestWebhookContext) CommandTitle(title, icon string) string        { panic("Not implemented") }
func (whc TestWebhookContext) CommandTitleNoTrans(title, icon string) string { panic("Not implemented") }

func (whc TestWebhookContext) Locale() Locale               { panic("Not implemented") }
func (whc TestWebhookContext) SetLocale(code5 string) error { panic("Not implemented") }

func (whc TestWebhookContext) NewMessage(text string) MessageFromBot { panic("Not implemented") }
func (whc TestWebhookContext) NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetHttpClient() *http.Client                  { panic("Not implemented") }
func (whc TestWebhookContext) IsNewerThen(chatEntity BotChat) bool          { panic("Not implemented") }
func (whc TestWebhookContext) UpdateLastProcessed(chatEntity BotChat) error { panic("Not implemented") }

func (whc TestWebhookContext) UserID() int64                                { panic("Not implemented") }
func (whc TestWebhookContext) CurrentUserKey() *datastore.Key               { panic("Not implemented") }
func (whc TestWebhookContext) GetAppUser() (*datastore.Key, AppUser, error) { panic("Not implemented") }

func (whc TestWebhookContext) GetLogger() Logger { panic("Not implemented") }

var _ WebhookContext = TestWebhookContext{}
