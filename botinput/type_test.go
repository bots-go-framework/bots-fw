package botinput_test

import (
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
)

func TestType_String(t *testing.T) {
	// In-range value resolves to its generated name.
	if got := botinput.TypeText.String(); got != "TypeText" {
		t.Errorf("TypeText.String() = %q, want %q", got, "TypeText")
	}
	// Out-of-range value falls back to the numeric form.
	if got := botinput.Type(-1).String(); !strings.HasPrefix(got, "Type(") {
		t.Errorf("Type(-1).String() = %q, want a Type(n) fallback", got)
	}
}

func TestGetBotInputTypeIdNameString(t *testing.T) {
	if got := botinput.GetBotInputTypeIdNameString(botinput.TypeText); got != "2:TypeText" {
		t.Errorf("GetBotInputTypeIdNameString(TypeText) = %q, want %q", got, "2:TypeText")
	}
}
