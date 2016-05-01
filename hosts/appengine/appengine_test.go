package gae_host

import (
	"testing"
	"github.com/strongo/strongo-bots"
)

func TestGaeLoggerIsInterfaceOfLogger(t *testing.T) {
	_ = bots.Logger(GaeLogger{})
}

func TestGaeBotHostIsInterfaceOfBotHost(t *testing.T) {
	_ = bots.BotHost(GaeBotHost{})
}