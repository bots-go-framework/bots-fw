package botsfw

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"time"

	"context"
	"github.com/strongo/app"
)

type TestWebhookContext struct {
	*WebhookContextBase
}

var _ WebhookContext = (*TestWebhookContext)(nil)

func (whc TestWebhookContext) BotChatIntID() int64 {
	return 0
}

func (whc TestWebhookContext) IsInGroup() bool {
	return false
}

func (whc TestWebhookContext) Close(c context.Context) error {
	return nil
}

func (whc TestWebhookContext) CreateBotUser(c context.Context, botID string, apiUser WebhookActor) (BotUser, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotChatEntityByID(c context.Context, botID, botChatID string) (BotChat, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotCode() string {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotToken() string {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotUserByID(c context.Context, botUserID interface{}) (BotUser, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetRecipient() WebhookRecipient {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetSender() WebhookSender {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetTime() time.Time {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputChosenInlineResult() WebhookChosenInlineResult {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputCallbackQuery() WebhookCallbackQuery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputDelivery() WebhookDelivery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputInlineQuery() WebhookInlineQuery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputMessage() WebhookMessage {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputPostback() WebhookPostback {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputType() WebhookInputType {
	panic("Not implemented")
}

func (whc TestWebhookContext) MessageText() string {
	return "test message"
}

func (whc TestWebhookContext) Translate(key string, args ...interface{}) string {
	return key
}

func (whc TestWebhookContext) TranslateNoWarning(key string, args ...interface{}) string {
	if len(args) > 0 {
		return fmt.Sprintf(key, args...)
	}
	return key
}

func (whc TestWebhookContext) NewChatEntity() BotChat { panic("Not implemented") }

func (whc TestWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	panic("Not implemented")
}
func (whc TestWebhookContext) Context() context.Context { panic("Not implemented") }

func (whc TestWebhookContext) ChatKey() *dal.Key                     { panic("Not implemented") }
func (whc TestWebhookContext) NewChatKey(c context.Context) *dal.Key { panic("Not implemented") }
func (whc TestWebhookContext) ChatEntity() BotChat                   { panic("Not implemented") }

func (whc TestWebhookContext) CommandText(title, icon string) string        { panic("Not implemented") }
func (whc TestWebhookContext) CommandTextNoTrans(title, icon string) string { panic("Not implemented") }

func (whc TestWebhookContext) Locale() strongo.Locale       { panic("Not implemented") }
func (whc TestWebhookContext) SetLocale(code5 string) error { panic("Not implemented") }

func (whc TestWebhookContext) NewMessage(text string) MessageFromBot { panic("Not implemented") }
func (whc TestWebhookContext) NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot {
	panic("Not implemented")
}

func (whc TestWebhookContext) NewEditMessage(text string, format MessageFormat) (MessageFromBot, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) Responder() WebhookResponder {
	panic("Not implemented")
}

func (whc TestWebhookContext) IsNewerThen(chatEntity BotChat) bool          { panic("Not implemented") }
func (whc TestWebhookContext) UpdateLastProcessed(chatEntity BotChat) error { panic("Not implemented") }

func (whc TestWebhookContext) UserID() int64                   { panic("Not implemented") }
func (whc TestWebhookContext) CurrentUserKey() *dal.Key        { panic("Not implemented") }
func (whc TestWebhookContext) GetAppUser() (BotAppUser, error) { panic("Not implemented") }

var _ WebhookContext = TestWebhookContext{}
