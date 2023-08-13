package botsfw

import "testing"

func TestNewWebhookContextBase(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("NewWebhookContextBase() did not panic")
			}
		}()
		args := CreateWebhookContextArgs{}
		NewWebhookContextBase(args, nil, nil, nil, nil)
	})
}
