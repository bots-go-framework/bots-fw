package telegram

import "github.com/strongo/bots-framework/core"

// Platform describes Telegram platform
type Platform struct {
}

var _ bots.BotPlatform = (*Platform)(nil)

// PlatformID is 'telegram'
const PlatformID = "telegram"

// ID returns 'telegram'
func (p Platform) ID() string {
	return PlatformID
}

// Version returns '2.0'
func (p Platform) Version() string {
	return "2.0"
}
