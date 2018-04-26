package viber

import "github.com/strongo/bots-framework/core"

// Platform describes Viber bot platform
type Platform struct {
}

var _ bots.BotPlatform = (*Platform)(nil)

// PlatformID is 'viber'
const PlatformID = "viber"

// ID return ID of Viber platform
func (p Platform) ID() string {
	return PlatformID
}

// Version return supported version of Viber API
func (p Platform) Version() string {
	return "1"
}
