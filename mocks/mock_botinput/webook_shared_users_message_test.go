package mock_botinput

import (
	"github.com/bots-go-framework/bots-fw/botinput"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestNewMockWebhookSharedUsersMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockM := NewMockWebhookSharedUsersMessage(ctrl)
	if mockM == nil {
		t.Fatalf("NewMockWebhookSharedUsersMessage returned nil")
	}
	expectedSharedUsers := []botinput.SharedUserMessageItem{
		NewMockSharedUserMessageItem(ctrl),
	}
	mockM.EXPECT().GetSharedUsers().Return(expectedSharedUsers)
	var _ botinput.WebhookInput = mockM
	var m botinput.WebhookSharedUsersMessage = mockM
	if sharedUser := m.GetSharedUsers(); sharedUser == nil {
		t.Fatalf("GetSharedUsers returned nil")
	} else if sharedUser[0] != expectedSharedUsers[0] {
		t.Fatalf("GetSharedUsers returned unxpcted result %v", sharedUser)
	}
}
