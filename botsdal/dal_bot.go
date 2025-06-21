package botsdal

import (
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
)

const botsCollection = "bots"

// NewBotKey creates a dalgo key to specific bot record
func NewBotKey(platformID botsfwconst.Platform, botID string) *dal.Key {
	platformKey := NewPlatformKey(platformID)
	if botID == "" {
		panic("botID is required parameter")
	}
	return dal.NewKeyWithParentAndID(platformKey, botsCollection, botID)
}
