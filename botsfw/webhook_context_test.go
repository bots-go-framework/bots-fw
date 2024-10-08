package botsfw

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"

	//"github.com/dal-go/dalgo/dal"
	"github.com/strongo/i18n"
	"net/http"
	"time"

	"context"
)

type TestWebhookContext struct {
	*WebhookContextBase
}

var _ WebhookContext = (*TestWebhookContext)(nil)

func (whc TestWebhookContext) BotChatIntID() int64 {
	return 0
}

func (whc TestWebhookContext) IsInGroup() (bool, error) {
	return false, nil
}

func (whc TestWebhookContext) Close(c context.Context) error {
	return nil
}

func (whc TestWebhookContext) CreateBotUser(c context.Context, botID string, apiUser botinput.WebhookActor) (botsfwmodels.PlatformUserData, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotChatEntityByID(c context.Context, botID, botChatID string) (botsfwmodels.BotChatData, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotCode() string {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotToken() string {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetBotUserByID(_ context.Context, botUserID string) (botsfwmodels.PlatformUserData, error) {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetRecipient() botinput.WebhookRecipient {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetSender() botinput.WebhookSender {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetTime() time.Time {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputChosenInlineResult() botinput.WebhookChosenInlineResult {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputCallbackQuery() botinput.WebhookCallbackQuery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputDelivery() botinput.WebhookDelivery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputInlineQuery() botinput.WebhookInlineQuery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputMessage() botinput.WebhookMessage {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputPostback() botinput.WebhookPostback {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputType() botinput.WebhookInputType {
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

func (whc TestWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	panic("Not implemented")
}
func (whc TestWebhookContext) Context() context.Context { panic("Not implemented") }

func (whc TestWebhookContext) ChatData() botsfwmodels.BotChatData { panic("Not implemented") }

func (whc TestWebhookContext) CommandText(title, icon string) string        { panic("Not implemented") }
func (whc TestWebhookContext) CommandTextNoTrans(title, icon string) string { panic("Not implemented") }

func (whc TestWebhookContext) Locale() i18n.Locale          { panic("Not implemented") }
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

func (whc TestWebhookContext) IsNewerThen(chatEntity botsfwmodels.BotChatData) bool {
	panic("Not implemented")
}
func (whc TestWebhookContext) UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error {
	panic("Not implemented")
}

func (whc TestWebhookContext) UserID() int64 { panic("Not implemented") }

// func (whc TestWebhookContext) CurrentUserKey() *dal.Key                     { panic("Not implemented") }
func (whc TestWebhookContext) AppUserData() (botsfwmodels.AppUserData, error) {
	panic("Not implemented")
}

var _ WebhookContext = TestWebhookContext{}
