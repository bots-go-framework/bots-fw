package mock_botinput

import (
	"github.com/bots-go-framework/bots-fw/botinput"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestNewMockWebhookInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInput := NewMockWebhookInput(ctrl)
	if mockInput == nil {
		t.Fatalf("NewMockWebhookInput() should not return nil")
	}
	mockInput.EXPECT().InputType().Return(botinput.WebhookInputText)
	if v := mockInput.InputType(); v != botinput.WebhookInputText {
		t.Errorf("InputType() = %v, want %v", v, botinput.WebhookInputText)
	}
}
