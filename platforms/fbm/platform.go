package fbm

import "github.com/strongo/bots-framework/core"

// PlatformID is platform ID for FBM
const PlatformID = "fbm"

// Platform describes FBM platform
type Platform struct {
}

var _ bots.BotPlatform = (*Platform)(nil)

// ID returns ID of FBM platform
func (p Platform) ID() string {
	return PlatformID
}

// Version returns version of the platform
func (p Platform) Version() string {
	return "1"
}
