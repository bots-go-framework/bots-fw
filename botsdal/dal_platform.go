package botsdal

import (
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
)

const botPlatformsCollection = "botPlatforms"

func NewPlatformKey(platform botsfwconst.Platform) *dal.Key {
	if platform == "" {
		panic("platform is required parameter")
	}
	return dal.NewKeyWithID(botPlatformsCollection, string(platform))
}
