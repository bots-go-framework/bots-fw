package botsdal

import "github.com/dal-go/dalgo/dal"

const botPlatformsCollection = "botPlatforms"

func NewPlatformKey(platform string) *dal.Key {
	if platform == "" {
		panic("platform is required parameter")
	}
	return dal.NewKeyWithID(botPlatformsCollection, platform)
}
