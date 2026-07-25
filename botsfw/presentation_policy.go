package botsfw

import (
	"context"
	"fmt"

	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-go-core/botkb"
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
	PersistentBottomKeyboard PersistentBottomKeyboardPolicy
}

// CommandResponseResponder is implemented by host-composed responders that
// select presentation authority from the router's matched command, rather than
// from data supplied by a feature message.
type CommandResponseResponder interface {
	ResponderForCommand(CommandCode) WebhookResponder
}

// NewCommandPolicyResponder makes router-return ownership depend on the command
// codes selected by host composition. Direct sends retain feature ownership.
func NewCommandPolicyResponder(next WebhookResponder, policy PresentationPolicy, hostCommands ...CommandCode) WebhookResponder {
	if next == nil {
		return nil
	}
	owned := make(map[CommandCode]struct{}, len(hostCommands))
	for _, code := range hostCommands {
		owned[code] = struct{}{}
	}
	return commandPolicyResponder{feature: newPolicyResponder(next, policy, false), host: next, policy: policy, hostCommands: owned}
}

type commandPolicyResponder struct {
	feature      WebhookResponder
	host         WebhookResponder
	policy       PresentationPolicy
	hostCommands map[CommandCode]struct{}
}

func (r commandPolicyResponder) SendMessage(ctx context.Context, m botmsg.MessageFromBot, channel botmsg.BotAPISendMessageChannel) (OnMessageSentResponse, error) {
	return r.feature.SendMessage(ctx, m, channel)
}

func (r commandPolicyResponder) DeleteMessage(ctx context.Context, messageID string) error {
	return r.feature.DeleteMessage(ctx, messageID)
}

func (r commandPolicyResponder) ResponderForCommand(code CommandCode) WebhookResponder {
	if _, hostOwned := r.hostCommands[code]; hostOwned {
		return newPolicyResponder(r.host, r.policy, true)
	}
	return r.feature
}

// ResponseResponderForCommand is called by the router after it has selected a
// command. Only a host-composed responder can grant host presentation authority.
func ResponseResponderForCommand(responder WebhookResponder, code CommandCode) WebhookResponder {
	if scoped, ok := responder.(CommandResponseResponder); ok {
		return scoped.ResponderForCommand(code)
	}
	return responder
}

func (p PresentationPolicy) Validate(m botmsg.MessageFromBot, hostOwned bool) error {
	if !hasPersistentBottomKeyboard(m) {
		return nil
	}
	switch p.PersistentBottomKeyboard {
	case "", PersistentBottomKeyboardAllow:
		return nil
	case PersistentBottomKeyboardDeny:
		return fmt.Errorf("persistent bottom keyboard denied by host presentation policy")
	case PersistentBottomKeyboardHostOnly:
		if hostOwned {
			return nil
		}
		return fmt.Errorf("persistent bottom keyboard is reserved for the host")
	default:
		return fmt.Errorf("unknown persistent bottom keyboard policy %q", p.PersistentBottomKeyboard)
	}
}

// NewPolicyResponder wraps both router-returned and direct responder sends,
// provided the host installs the wrapper in the WebhookContext and router.
// NewPolicyResponder creates a feature-owned responder; it cannot send a
// host-only bottom keyboard. Hosts must use NewHostPolicyResponder explicitly.
func NewPolicyResponder(next WebhookResponder, policy PresentationPolicy) WebhookResponder {
	return newPolicyResponder(next, policy, false)
}

func NewHostPolicyResponder(next WebhookResponder, policy PresentationPolicy) WebhookResponder {
	return newPolicyResponder(next, policy, true)
}

func newPolicyResponder(next WebhookResponder, policy PresentationPolicy, hostOwned bool) WebhookResponder {
	if next == nil {
		return nil
	}
	return policyResponder{next: next, policy: policy, hostOwned: hostOwned}
}

type policyResponder struct {
	next      WebhookResponder
	policy    PresentationPolicy
	hostOwned bool
}

func (r policyResponder) SendMessage(ctx context.Context, m botmsg.MessageFromBot, channel botmsg.BotAPISendMessageChannel) (OnMessageSentResponse, error) {
	if err := r.policy.Validate(m, r.hostOwned); err != nil {
		return OnMessageSentResponse{}, err
	}
	return r.next.SendMessage(ctx, m, channel)
}

func hasPersistentBottomKeyboard(m botmsg.MessageFromBot) bool {
	if m.Presentation.PersistentBottomKeyboard || keyboardIsBottom(m.Keyboard) {
		return true
	}
	switch message := m.BotMessage.(type) {
	case *botmsg.TextMessageFromBot:
		return keyboardIsBottom(message.Keyboard)
	}
	return false
}

func keyboardIsBottom(keyboard botkb.Keyboard) bool {
	return keyboard != nil && keyboard.KeyboardType() == botkb.KeyboardTypeBottom
}

func (r policyResponder) DeleteMessage(ctx context.Context, messageID string) error {
	return r.next.DeleteMessage(ctx, messageID)
}
