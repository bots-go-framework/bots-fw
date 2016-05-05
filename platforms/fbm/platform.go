package fbm_strongo_bot

import "github.com/strongo/bots-framework/core"

type FbmPlatform struct {
}

var _ bots.BotPlatform = (*FbmPlatform)(nil)

func (p FbmPlatform) Id() string {
	return "fbm"
}

func (p FbmPlatform) Version() string {
	return "1"
}
