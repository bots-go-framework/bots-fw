package botsfw

import (
	"context"
	"fmt"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

type PersistentBottomKeyboardPolicy string

const (
	PersistentBottomKeyboardAllow    PersistentBottomKeyboardPolicy = "allow"
	PersistentBottomKeyboardDeny     PersistentBottomKeyboardPolicy = "deny"
	PersistentBottomKeyboardHostOnly PersistentBottomKeyboardPolicy = "host-only"
)

// PresentationPolicy is supplied by the bot host and is applied before every
// responder send when the responder is wrapped with NewPolicyResponder.
type PresentationPolicy struct {
	PersistentBottomKeyboard  PersistentBottomKeyboardPolicy
	HostMayShowBottomKeyboard bool
}

func (p PresentationPolicy) Validate(m botmsg.MessageFromBot) error {
	if !m.Presentation.PersistentBottomKeyboard {
		return nil
	}
	switch p.PersistentBottomKeyboard {
	case "", PersistentBottomKeyboardAllow:
		return nil
	case PersistentBottomKeyboardDeny:
		return fmt.Errorf("persistent bottom keyboard denied by host presentation policy")
	case PersistentBottomKeyboardHostOnly:
		if p.HostMayShowBottomKeyboard {
			return nil
		}
		return fmt.Errorf("persistent bottom keyboard is reserved for the host")
	default:
		return fmt.Errorf("unknown persistent bottom keyboard policy %q", p.PersistentBottomKeyboard)
	}
}

// NewPolicyResponder wraps both router-returned and direct responder sends,
// provided the host installs the wrapper in the WebhookContext and router.
func NewPolicyResponder(next WebhookResponder, policy PresentationPolicy) WebhookResponder {
	if next == nil {
		return nil
	}
	return policyResponder{next: next, policy: policy}
}

type policyResponder struct {
	next   WebhookResponder
	policy PresentationPolicy
}

func (r policyResponder) SendMessage(ctx context.Context, m botmsg.MessageFromBot, channel botmsg.BotAPISendMessageChannel) (OnMessageSentResponse, error) {
	if err := r.policy.Validate(m); err != nil {
		return OnMessageSentResponse{}, err
	}
	return r.next.SendMessage(ctx, m, channel)
}

func (r policyResponder) DeleteMessage(ctx context.Context, messageID string) error {
	return r.next.DeleteMessage(ctx, messageID)
}
