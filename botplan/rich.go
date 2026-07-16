package botplan

// Rich is a neutral rich-text model: an ordered list of lines, each carrying a
// kind (paragraph, list-item, quote) and a sequence of styled spans.
//
// It is deliberately a pragmatic MVP, not a full document model. It captures
// exactly the structure the ToGethered scenarios use — emphasis, inline code,
// anchored links, bulleted lists and quoted lines — and nothing more. Crucially
// it holds NO HTML or markdown strings: the renderer, not the app, decides how a
// BoldSpan becomes "<b>…</b>" (Telegram) or "*…*" (WhatsApp) or is dropped
// (plain text). See capability-map telegram/text-formatting and
// whatsapp/text-formatting.
type Rich struct {
	Lines []Line
}

// LineKind classifies a whole line of rich text.
type LineKind int

const (
	// LineParagraph is an ordinary text line.
	LineParagraph LineKind = iota
	// LineListItem is a bulleted list item. Renderers prefix it with the
	// platform's bullet marker ("• " for HTML, "* " for WhatsApp markers).
	LineListItem
	// LineQuote is a quoted line. Renderers emit <blockquote> (Telegram) or a
	// leading "> " (WhatsApp); plain text falls back to a leading "> ".
	LineQuote
)

// Line is one line of rich text: a kind plus the spans that compose it.
type Line struct {
	Kind  LineKind
	Spans []Span
}

// SpanKind classifies an inline run of text within a Line.
type SpanKind int

const (
	// SpanText is unstyled text.
	SpanText SpanKind = iota
	// SpanBold is bold text.
	SpanBold
	// SpanItalic is italic text.
	SpanItalic
	// SpanCode is inline monospace/code text.
	SpanCode
	// SpanLink is an anchored hyperlink: Text is the anchor, URL the target.
	// On platforms without anchor-text links (WhatsApp — see capability-map
	// whatsapp/text-formatting noAnchorTextLinks) the renderer degrades this to
	// "anchor: url" so the destination stays reachable.
	SpanLink
)

// Span is a styled inline run of text. For SpanLink, URL is the target and Text
// is the anchor; for every other kind URL is empty.
type Span struct {
	Kind SpanKind
	Text string
	URL  string // set only when Kind == SpanLink
}

// --- ergonomic constructors (spans) ---

// Text returns an unstyled span.
func Text(s string) Span { return Span{Kind: SpanText, Text: s} }

// Bold returns a bold span.
func Bold(s string) Span { return Span{Kind: SpanBold, Text: s} }

// Italic returns an italic span.
func Italic(s string) Span { return Span{Kind: SpanItalic, Text: s} }

// Code returns an inline-code span.
func Code(s string) Span { return Span{Kind: SpanCode, Text: s} }

// Link returns an anchored-link span (anchor text + target URL).
func Link(anchor, url string) Span { return Span{Kind: SpanLink, Text: anchor, URL: url} }

// --- ergonomic constructors (lines) ---

// Para returns a paragraph line from the given spans.
func Para(spans ...Span) Line { return Line{Kind: LineParagraph, Spans: spans} }

// Item returns a list-item line from the given spans.
func Item(spans ...Span) Line { return Line{Kind: LineListItem, Spans: spans} }

// Quote returns a quote line from the given spans.
func Quote(spans ...Span) Line { return Line{Kind: LineQuote, Spans: spans} }

// RichText builds a single-paragraph Rich from a plain string. A convenience for
// the common case of an unformatted message.
func RichText(s string) Rich {
	return Rich{Lines: []Line{Para(Text(s))}}
}

// PlainString flattens Rich to unstyled text, one line per Line, dropping all
// styling. List items gain a "• " prefix and quotes a "> " prefix so the shape
// survives. Renderers use their own markup path; this is a neutral fallback and
// a convenience for tests and logging.
func (r Rich) PlainString() string {
	var out string
	for i, line := range r.Lines {
		if i > 0 {
			out += "\n"
		}
		switch line.Kind {
		case LineListItem:
			out += "• "
		case LineQuote:
			out += "> "
		}
		for _, sp := range line.Spans {
			if sp.Kind == SpanLink && sp.Text != sp.URL && sp.Text != "" {
				out += sp.Text + ": " + sp.URL
				continue
			}
			if sp.Kind == SpanLink {
				out += sp.URL
				continue
			}
			out += sp.Text
		}
	}
	return out
}

// IsEmpty reports whether the Rich carries no lines or only empty lines.
func (r Rich) IsEmpty() bool {
	for _, line := range r.Lines {
		for _, sp := range line.Spans {
			if sp.Text != "" || sp.URL != "" {
				return false
			}
		}
	}
	return true
}
