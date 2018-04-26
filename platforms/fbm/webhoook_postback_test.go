package fbm

import (
	"github.com/strongo/bots-framework/core"
	"testing"
)

func TestFbmPostbackInputIsWebhookCallbackQuery(t *testing.T) {
	var _ bots.WebhookCallbackQuery = postbackInput{}
}
