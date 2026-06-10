package botsframework

import "testing"

func TestPackage(t *testing.T) {
	if got := packageName(); got != "botsframework" {
		t.Errorf("packageName() = %q, want %q", got, "botsframework")
	}
}
