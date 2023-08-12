package botsdal

import "github.com/dal-go/dalgo/dal"

const botPlatformsCollection = "botPlatforms"

func NewPlatformKey(platform string) *dal.Key {
	return dal.NewKeyWithID(botPlatformsCollection, platform)
}
