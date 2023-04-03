package botsfw

import "testing"

func TestNewWebhookContextBase(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("NewWebhookContextBase() did not panic")
			}
		}()
		NewWebhookContextBase(nil, nil, nil, BotContext{}, nil, BotCoreStores{}, nil, nil, nil)
	})
}
