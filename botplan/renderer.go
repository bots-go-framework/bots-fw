package botplan

import "github.com/bots-go-framework/bots-fw/botmsg"

// Platform names a messaging platform for renderer selection. It is a string so
// adapters can register without a shared enum the neutral layer would have to
// grow per platform.
type Platform string

const (
	// PlatformTelegram is the Telegram platform.
	PlatformTelegram Platform = "telegram"
	// PlatformWhatsApp is the WhatsApp platform.
	PlatformWhatsApp Platform = "whatsapp"
)

// ChatType distinguishes private chats from group chats, which changes rendering
// (e.g. group posts, mention semantics).
type ChatType int

const (
	// ChatPrivate is a 1:1 chat.
	ChatPrivate ChatType = iota
	// ChatGroup is a group chat.
	ChatGroup
)

// RenderTarget carries the dynamic facts a renderer needs beyond its static
// Descriptor: which platform and chat type, whether the send window is open
// right now, and the locale to render in.
//
// WindowOpen is only meaningful when the platform's Descriptor.WindowGated is
// true (WhatsApp). For Telegram it is effectively always open and the renderer
// ignores it (capability-map telegram/proactive-messaging).
type RenderTarget struct {
	Platform   Platform
	ChatType   ChatType
	WindowOpen bool
	Locale     string
}

// Renderer turns a neutral MessagePlan into one or more concrete
// botmsg.MessageFromBot values for a specific platform, degrading capabilities
// the target lacks.
//
// A single plan may yield several messages: a WhatsApp plan carrying both a
// prompt and a URL action becomes a prompt message followed by a cta_url message
// (capability-map whatsapp/cta-url-button — the two cannot share one message).
// The messages are returned in send order.
//
// Errors are typed so callers can react by policy: ErrNoTemplateForPurpose and
// ErrTemplateMismatch signal an out-of-window proactive send that cannot be
// delivered as-is (scenario-catalogue SYS-TPL-030 — the caller applies the
// scenario's declared degradation), ErrInvalidPlan a malformed plan, and
// ErrUnsupportedTarget a platform the renderer does not implement.
type Renderer interface {
	// Render produces the concrete messages for plan on target, or a typed error.
	Render(plan MessagePlan, target RenderTarget) ([]botmsg.MessageFromBot, error)

	// Descriptor returns the renderer's static capability fact-sheet.
	Descriptor() Descriptor
}
