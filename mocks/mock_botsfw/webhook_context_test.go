package mock_botsfw

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestNewMockWebhookContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	whcMock := NewMockWebhookContext(ctrl)
	var _ botsfw.WebhookContext = whcMock
	if whcMock == nil {
		t.Fatalf("NewMockWebhookContext() should not return nil")
	}
	whcMock.EXPECT().AppUserID().Return("123")
	_ = whcMock.AppUserID()
}
