package viber_bot

import "github.com/strongo/bots-framework/core"

type ViberPlatform struct {
}

var _ bots.BotPlatform = (*ViberPlatform)(nil)

const ViberPlatformID = "viber"

func (p ViberPlatform) Id() string {
	return ViberPlatformID
}

func (p ViberPlatform) Version() string {
	return "1"
}
