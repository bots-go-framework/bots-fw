package botplan

// TextMarkup identifies how a platform expresses rich-text styling, so a Rich
// renderer knows which dialect to emit.
type TextMarkup int

const (
	// MarkupPlain: no styling; spans render as their text. (No platform in the
	// pilot set is plain-only, but a renderer may downgrade to it.)
	MarkupPlain TextMarkup = iota
	// MarkupHTML: an HTML subset with a parse-mode opt-in and anchor-text links.
	// Telegram (capability-map telegram/text-formatting: htmlTags include
	// b/i/code/a, hyperlinks=true).
	MarkupHTML
	// MarkupMarkers: always-on inline markers, no parse-mode, no anchor-text
	// links. WhatsApp (capability-map whatsapp/text-formatting: markers map,
	// noParseModeParameter, noAnchorTextLinks).
	MarkupMarkers
)

// Descriptor is the static per-platform capability fact-sheet a renderer
// consults. It is the F5 half of the seam: every field mirrors a record in
// can-i-use/capability-map.json (cited in the field's doc comment) so the
// descriptor a renderer ships can be CI-checked against the platform facts.
//
// It carries only the facts the two pilot renderers genuinely consult when
// shaping a MessagePlan — not the whole capability map. Dynamic facts (is the
// window open right now, is this template approved) are NOT here; they live on
// RenderTarget and in the template catalog, because they vary per send.
type Descriptor struct {
	// MaxPromptButtons is the largest number of choices that render as inline/
	// reply buttons before the renderer must degrade to a list or pages.
	// Telegram: no fixed small cap (inline grid) — set high. WhatsApp:
	// capability-map whatsapp/reply-buttons constraints.maxButtons = 3.
	MaxPromptButtons int

	// MaxListRows is the largest number of choices that render as a single
	// selectable list before paging is required. Telegram has no list message
	// (buttons only) — 0 means "not applicable". WhatsApp: capability-map
	// whatsapp/list-messages constraints.maxRowsAcrossAllSections = 10.
	MaxListRows int

	// MaxButtonLabelChars is the button/row label ceiling; longer labels are
	// truncated. Telegram: no documented hard limit — 0 means "unbounded".
	// WhatsApp: capability-map whatsapp/reply-buttons buttonLabelMaxChars = 20.
	MaxButtonLabelChars int

	// SupportsEdit reports whether the platform can update a message in place
	// (LivePanel). Telegram: capability-map telegram/edit-message native = true.
	// WhatsApp: capability-map whatsapp/edit-message absent = false → append.
	SupportsEdit bool

	// SupportsDelete reports whether the platform can retract a message.
	// Telegram: capability-map telegram/delete-message native = true. WhatsApp:
	// capability-map whatsapp/delete-message absent = false.
	SupportsDelete bool

	// SupportsCallbackAck reports whether a tap can be acknowledged (toast /
	// spinner-stop). Telegram: capability-map telegram/callback-query
	// semanticsRequiresAck = true. WhatsApp: capability-map whatsapp/callback-ack
	// absent = false — there is nothing to ack.
	SupportsCallbackAck bool

	// WindowGated reports whether proactive sends are gated by a service window
	// that may force a template. Telegram: false (capability-map
	// telegram/proactive-messaging native, no window). WhatsApp: true
	// (capability-map whatsapp/customer-service-window partial, 24h).
	WindowGated bool

	// SupportsInlineURLButton reports whether a fully dynamic URL button can ride
	// a normal (non-template) message. Telegram: capability-map
	// telegram/inline-keyboard (URL buttons) = true. WhatsApp: capability-map
	// whatsapp/cta-url-button native = true, but only in-window — the renderer
	// combines this flag with RenderTarget.WindowOpen.
	SupportsInlineURLButton bool

	// SupportsButtonGrid reports whether buttons arrange in a multi-column grid
	// (so ActionPrompt.LayoutRows is meaningful). Telegram: capability-map
	// telegram/inline-keyboard constraints.layout = "grid" = true. WhatsApp:
	// capability-map whatsapp/reply-buttons grid = false.
	SupportsButtonGrid bool

	// SupportsMedia reports whether the platform can send an image. Telegram:
	// capability-map telegram/send-photo native = true. WhatsApp: capability-map
	// whatsapp/send-image native = true (though the client may not yet implement
	// it — a platform fact, not a client fact).
	SupportsMedia bool

	// TextMarkup is the styling dialect the Rich renderer must emit. See
	// TextMarkup. Telegram: MarkupHTML. WhatsApp: MarkupMarkers.
	TextMarkup TextMarkup

	// SupportsAnchorTextLinks reports whether a link span can render as anchor
	// text over a hidden URL. Telegram: capability-map telegram/text-formatting
	// hyperlinks = true. WhatsApp: capability-map whatsapp/text-formatting
	// noAnchorTextLinks → false, so links render as "anchor: url".
	SupportsAnchorTextLinks bool
}
