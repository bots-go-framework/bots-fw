package gae_host

import (
	"github.com/strongo/strongo-bots"
	"testing"
)

func TestGaeLoggerIsInterfaceOfLogger(t *testing.T) {
	_ = bots.Logger(GaeLogger{})
}

func TestGaeBotHostIsInterfaceOfBotHost(t *testing.T) {
	_ = bots.BotHost(GaeBotHost{})
}
