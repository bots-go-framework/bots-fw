package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
)

const botUsersCollection = "botUsers"

// NewPlatformUserKey creates a dalgo key to specific bot user record
func NewPlatformUserKey(platformID botsfwconst.Platform, botUserID string) *dal.Key {
	platformKey := NewPlatformKey(platformID)
	if botUserID == "" {
		panic("botUserID is required parameter")
	}
	return dal.NewKeyWithParentAndID(platformKey, botUsersCollection, botUserID)
}

// GetPlatformUser loads bot user data
func GetPlatformUser(
	ctx context.Context,
	tx dal.ReadSession,
	platformID botsfwconst.Platform, botUserID string,
	platformUserData botsfwmodels.PlatformUserData,
) (botUser BotUser, err error) {
	botUserKey := NewPlatformUserKey(platformID, botUserID)
	dataWithID, err := dal.GetRecordWithIDIntoData(ctx, tx, botUserKey, botUserID, platformUserData)
	return BotUser(dataWithID), err
}

// CreatePlatformUserRecord creates bot user record in database
func CreatePlatformUserRecord(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	platformID botsfwconst.Platform, botUserID string,
	platformUserData botsfwmodels.PlatformUserData,
) (err error) {
	if validatableData, ok := platformUserData.(interface{ Validate() error }); ok {
		if err = validatableData.Validate(); err != nil {
			return err
		}
	}
	key := NewPlatformUserKey(platformID, botUserID)
	_, err = dal.InsertRecordWithDataAndID(ctx, tx, key, botUserID, platformUserData)
	return err
}
