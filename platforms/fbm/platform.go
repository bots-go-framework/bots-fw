package fbm_bot

import "github.com/strongo/bots-framework/core"

const FbmPlatformID = "fbm"

type FbmPlatform struct {
}

var _ bots.BotPlatform = (*FbmPlatform)(nil)

func (p FbmPlatform) Id() string {
	return FbmPlatformID
}

func (p FbmPlatform) Version() string {
	return "1"
}
