package botsfw

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	botsfw2 "github.com/bots-go-framework/bots-fw/botmsg"
)

var _ WebhookContext = (*whContextDummy)(nil)

// whContextDummy is a dummy implementation of WebhookContext interface
// It exists only to check what is NOT implemented by WebhookContextBase
type whContextDummy struct {
	*WebhookContextBase
}

func (w whContextDummy) NewEditMessage(text string, format botsfw2.Format) (botsfw2.MessageFromBot, error) {
	panic(fmt.Sprintf("must be implemented in platform specific code: text=%s, format=%v", text, format))
}

func (w whContextDummy) UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error {
	panic(fmt.Sprintf("implement me in WebhookContextBase - UpdateLastProcessed(chatEntity=%v)", chatEntity))
}

func (w whContextDummy) AppUserData() (botsfwmodels.AppUserData, error) {
	panic("implement me in WebhookContextBase") //TODO
}

func (w whContextDummy) IsNewerThen(chatEntity botsfwmodels.BotChatData) bool {
	panic(fmt.Sprintf("implement me in WebhookContextBase - IsNewerThen(chatEntity=%v)", chatEntity))
}

func (w whContextDummy) Responder() WebhookResponder {
	//TODO implement me
	panic("implement me")
}
