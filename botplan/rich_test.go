package botplan

import "testing"

func TestRichText(t *testing.T) {
	r := RichText("hello")
	if len(r.Lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(r.Lines))
	}
	if r.Lines[0].Kind != LineParagraph {
		t.Errorf("want LineParagraph, got %v", r.Lines[0].Kind)
	}
	if got := r.PlainString(); got != "hello" {
		t.Errorf("PlainString = %q, want %q", got, "hello")
	}
}

func TestRichPlainString(t *testing.T) {
	tests := []struct {
		name string
		rich Rich
		want string
	}{
		{
			name: "styled spans flatten to their text",
			rich: Rich{Lines: []Line{Para(Text("go "), Bold("fast"), Text(" and "), Italic("slow"))}},
			want: "go fast and slow",
		},
		{
			name: "code span",
			rich: Rich{Lines: []Line{Para(Text("run "), Code("go test"))}},
			want: "run go test",
		},
		{
			name: "link with distinct anchor becomes anchor: url",
			rich: Rich{Lines: []Line{Para(Link("View day", "https://x.io/d/1"))}},
			want: "View day: https://x.io/d/1",
		},
		{
			name: "link with anchor equal to url becomes just url",
			rich: Rich{Lines: []Line{Para(Link("https://x.io", "https://x.io"))}},
			want: "https://x.io",
		},
		{
			name: "list items get bullet prefix",
			rich: Rich{Lines: []Line{Item(Text("kite")), Item(Text("surf"))}},
			want: "• kite\n• surf",
		},
		{
			name: "quote gets prefix",
			rich: Rich{Lines: []Line{Quote(Text("who's around?"))}},
			want: "> who's around?",
		},
		{
			name: "multi-line mixed",
			rich: Rich{Lines: []Line{
				Para(Bold("Saturday")),
				Item(Text("kitesurf 15:00")),
				Quote(Text("moved from 14:00")),
			}},
			want: "Saturday\n• kitesurf 15:00\n> moved from 14:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rich.PlainString(); got != tt.want {
				t.Errorf("PlainString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRichIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		rich Rich
		want bool
	}{
		{"nil lines", Rich{}, true},
		{"empty spans", Rich{Lines: []Line{Para()}}, true},
		{"whitespace-free empty span", Rich{Lines: []Line{Para(Text(""))}}, true},
		{"has text", Rich{Lines: []Line{Para(Text("x"))}}, false},
		{"link-only url", Rich{Lines: []Line{Para(Link("", "u"))}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rich.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpanConstructors(t *testing.T) {
	tests := []struct {
		name string
		span Span
		kind SpanKind
	}{
		{"text", Text("a"), SpanText},
		{"bold", Bold("a"), SpanBold},
		{"italic", Italic("a"), SpanItalic},
		{"code", Code("a"), SpanCode},
		{"link", Link("a", "u"), SpanLink},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.span.Kind != tt.kind {
				t.Errorf("Kind = %v, want %v", tt.span.Kind, tt.kind)
			}
		})
	}
	if l := Link("anchor", "url"); l.Text != "anchor" || l.URL != "url" {
		t.Errorf("Link built %+v", l)
	}
}

func TestLineConstructors(t *testing.T) {
	if Para().Kind != LineParagraph {
		t.Error("Para kind")
	}
	if Item().Kind != LineListItem {
		t.Error("Item kind")
	}
	if Quote().Kind != LineQuote {
		t.Error("Quote kind")
	}
}
