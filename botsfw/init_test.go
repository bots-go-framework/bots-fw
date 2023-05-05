package botsfw

import "testing"

func TestInitializeBotsFw(t *testing.T) {
	InitializeBotsFw(&testLogger{T: t})
}
