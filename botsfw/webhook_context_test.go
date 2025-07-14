package botsfw

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"

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

func (whc TestWebhookContext) CreateBotUser(c context.Context, botID string, apiUser botinput.Actor) (botsfwmodels.PlatformUserData, error) {
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

func (whc TestWebhookContext) GetRecipient() botinput.Recipient {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetSender() botinput.Sender {
	panic("Not implemented")
}

func (whc TestWebhookContext) GetTime() time.Time {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputChosenInlineResult() botinput.ChosenInlineResult {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputCallbackQuery() botinput.CallbackQuery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputDelivery() botinput.Delivery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputInlineQuery() botinput.InlineQuery {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputMessage() botinput.Message {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputPostback() botinput.Postback {
	panic("Not implemented")
}

func (whc TestWebhookContext) InputType() botinput.Type {
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

func (whc TestWebhookContext) NewMessage(text string) botmsg.MessageFromBot {
	panic("Not implemented")
}
func (whc TestWebhookContext) NewMessageByCode(messageCode string, a ...interface{}) botmsg.MessageFromBot {
	panic("Not implemented")
}

func (whc TestWebhookContext) NewEditMessage(text string, format botmsg.Format) (botmsg.MessageFromBot, error) {
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
