package mock_botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestNewMockWebhookResponder(t *testing.T) {
	ctrl := gomock.NewController(t)
	responderMock := NewMockWebhookResponder(ctrl)
	var _ botsfw.WebhookResponder = responderMock
	if responderMock == nil {
		t.Fatalf("NewMockWebhookResponder() should not return nil")
	}
	ctx := context.Background()
	var m botsfw.MessageFromBot
	responderMock.EXPECT().SendMessage(ctx, m, botsfw.BotAPISendMessageOverHTTPS).Return(botsfw.OnMessageSentResponse{}, nil)
	_, err := responderMock.SendMessage(ctx, m, botsfw.BotAPISendMessageOverHTTPS)
	if err != nil {
		t.Fatalf("SendMessage() should not return error")
	}
}
