package botsfwtest

import (
	"context"

	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
)

// FakeResponder implements botsfw.WebhookResponder and records every message
// passed to SendMessage so tests can inspect what the command produced.
type FakeResponder struct {
	// Sent accumulates every MessageFromBot delivered to SendMessage.
	Sent []botmsg.MessageFromBot
}

var _ botsfw.WebhookResponder = (*FakeResponder)(nil)

// SendMessage records the message and returns a zero response with no error.
func (r *FakeResponder) SendMessage(_ context.Context, m botmsg.MessageFromBot, _ botmsg.BotAPISendMessageChannel) (botsfw.OnMessageSentResponse, error) {
	r.Sent = append(r.Sent, m)
	return botsfw.OnMessageSentResponse{}, nil
}

// DeleteMessage is a no-op.
func (r *FakeResponder) DeleteMessage(_ context.Context, _ string) error {
	return nil
}
