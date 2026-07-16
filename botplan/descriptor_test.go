package botplan

import "testing"

// TestDescriptorZeroValue documents that the zero Descriptor is the most
// conservative platform: no buttons, no edit/delete/ack, plain text, no media.
// A renderer that forgets to set a field therefore degrades rather than
// over-promises.
func TestDescriptorZeroValue(t *testing.T) {
	var d Descriptor
	if d.MaxPromptButtons != 0 || d.MaxListRows != 0 {
		t.Error("zero descriptor should advertise no buttons")
	}
	if d.SupportsEdit || d.SupportsDelete || d.SupportsCallbackAck ||
		d.WindowGated || d.SupportsInlineURLButton || d.SupportsButtonGrid ||
		d.SupportsMedia || d.SupportsAnchorTextLinks {
		t.Error("zero descriptor should advertise no capabilities")
	}
	if d.TextMarkup != MarkupPlain {
		t.Error("zero descriptor markup should be plain")
	}
}

func TestTextMarkupDistinct(t *testing.T) {
	if MarkupPlain == MarkupHTML || MarkupHTML == MarkupMarkers || MarkupPlain == MarkupMarkers {
		t.Error("markup constants must be distinct")
	}
}

func TestPlatformConstants(t *testing.T) {
	if PlatformTelegram != "telegram" || PlatformWhatsApp != "whatsapp" {
		t.Errorf("platform constants changed: %q %q", PlatformTelegram, PlatformWhatsApp)
	}
}
