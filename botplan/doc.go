// Package botplan defines the neutral message-plan model and the capability
// descriptor that per-platform renderers consult (the "capability-driven
// rendering" seam, layer 5 of the ToGethered conversation architecture).
//
// An application composes a MessagePlan once — a platform-neutral statement of
// intent: some rich text, optionally a prompt (choose one of N), a URL action,
// a live-panel marker, a proactive-send spec, and/or a single image. It never
// mentions Telegram or WhatsApp. A per-platform Renderer then turns that plan
// into one or more botmsg.MessageFromBot values, degrading capabilities the
// target platform lacks (a WhatsApp button grid becomes reply-buttons, then a
// list, then a paged list; a live-panel edit becomes an append; an out-of-window
// proactive send becomes an approved template).
//
// Two ideas keep platform conditionals out of the application:
//
//   - The neutral types carry no HTML or markdown strings. Text is modelled as
//     lines of typed spans (Rich) so each renderer can emit HTML (Telegram),
//     WhatsApp markers, or plain text without the app choosing.
//
//   - Descriptor records the facts a renderer genuinely consults — button
//     ceilings, edit/delete support, markup dialect, window gating. Every field
//     mirrors a record in can-i-use/capability-map.json (cited per field) so the
//     descriptors can be CI-checked against the platform facts.
//
// Renderers live in the adapter repos (bots-fw-telegram, bots-fw-whatsapp); this
// package owns only the neutral vocabulary and the interface they implement.
package botplan
