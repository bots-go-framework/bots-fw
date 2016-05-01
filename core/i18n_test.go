package bots

import (
	"testing"
)

func TestLocale_String(t *testing.T) {
	l := Locale{Code5: "c0de5", IsRtl: true, NativeTitle: "Код05", EnglishTitle: "C0de5", FlagIcon: "fl0g"}
	actualLs := l.String()
	expectingLs := `Locale{Code5: "c0de5", IsRtl: true, NativeTitle: "Код05", EnglishTitle: "C0de5", FlagIcon: "fl0g"}`
	if actualLs != expectingLs {
		t.Errorf("Unexpected result of func (Locale) String(). Got: %v. Exepcted: %v", actualLs, expectingLs)
	}
}