package botplan

// LivePanel marks a plan as an update-in-place-if-possible of an earlier message
// (capability vocabulary: live-panel; capability-map telegram/edit-message).
//
// PanelKey is the app's stable identifier for the logical panel (e.g. a
// spot-day card). The neutral layer does not know which platform message that
// maps to — the mapping from PanelKey to a concrete message reference is a
// projection concern the caller owns (architecture.md §4.2: "PanelKey→caller-
// supplied message ref mapping stays outside"). The renderer only records the
// intent to update in place; a platform with no edit endpoint (WhatsApp —
// capability-map whatsapp/edit-message absent) degrades it to an append.
type LivePanel struct {
	PanelKey string
}

// ProactiveSpec marks a plan as a proactive (bot-initiated) send rather than a
// reply, and names the purpose that selects an approved template when the
// platform's send window is closed (capability vocabulary: gated-send).
//
// A nil *ProactiveSpec on a MessagePlan means "this is a reply" — deliverable
// as a free-form message with no template involved. A non-nil ProactiveSpec on
// WhatsApp is gated by the 24-hour customer-service window (capability-map
// whatsapp/customer-service-window): in-window it renders free-form, out-of-
// window it must map Purpose→an approved template (capability-map
// whatsapp/template-messages). On Telegram a proactive send is an ordinary
// message — Telegram has no window and no templates (capability-map
// telegram/proactive-messaging, telegram/message-templates absent) — so
// Purpose/Params are ignored there.
type ProactiveSpec struct {
	// Purpose is the template-catalog purpose key (e.g. "intent_notice").
	Purpose string
	// Locale selects the template localisation (e.g. "en" or "en_US"). Empty
	// lets the catalog fall back to its default.
	Locale string
	// Params maps template placeholder names to values. On WhatsApp these become
	// the template's body parameters, in the order the TemplateDef declares.
	Params map[string]string
}

// MediaRef is a single image to accompany the message (capability vocabulary:
// media; capability-map telegram/send-photo, whatsapp/send-image).
//
// Exactly one of ImageURL or MediaID identifies the asset. MediaID is a
// platform-uploaded asset handle (Meta recommends it over URL — capability-map
// whatsapp/send-image); ImageURL is a publicly hosted asset. Caption is optional
// and, where the platform supports it, rendered alongside the image.
type MediaRef struct {
	ImageURL string
	MediaID  string
	Caption  string
}
