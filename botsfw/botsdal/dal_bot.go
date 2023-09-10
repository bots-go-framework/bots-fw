package botsdal

import "github.com/dal-go/dalgo/dal"

const botsCollection = "bots"

// NewBotKey creates a dalgo key to specific bot record
func NewBotKey(platformID, botID string) *dal.Key {
	platformKey := NewPlatformKey(platformID)
	return dal.NewKeyWithParentAndID(platformKey, botsCollection, botID)
}
