package fbm_bot

import (
	"testing"
	"github.com/strongo/bots-framework/core"
)

func TestFbmPostbackInputIsWebhookCallbackQuery(t *testing.T) {
	var _ bots.WebhookCallbackQuery = FbmPostbackInput{}
}
