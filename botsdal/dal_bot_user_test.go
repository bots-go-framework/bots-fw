package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	"testing"
)

func TestGetBotUser(t *testing.T) {
	type args struct {
		//collection string
		platform  string
		botUserID string
	}
	tests := []struct {
		name        string
		args        args
		shouldPanic bool
		checkResult func(botUser BotUser, err error)
	}{
		{name: "empty", shouldPanic: true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("newBotUserStore() did not panic")
					}
				}()
			}
			var tx dal.ReadwriteTransaction
			botUser, err := GetPlatformUser(ctx, tx, tt.args.platform, tt.args.botUserID, nil)
			tt.checkResult(botUser, err)
		})
	}
}

func TestCreateBotUserRecord(t *testing.T) {
	type args struct {
		platform    string
		botUserID   string
		botUserData botsfwmodels.PlatformUserData
	}
	tests := []struct {
		name        string
		args        args
		shouldPanic bool
		checkResult func(err error)
	}{
		{
			name:        "empty",
			shouldPanic: true,
			checkResult: func(err error) {
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("CreatePlatformUserRecord() did not panic")
					}
				}()
			}
			var tx dal.ReadwriteTransaction
			err := CreatePlatformUserRecord(ctx, tx, tt.args.platform, tt.args.botUserID, tt.args.botUserData)
			tt.checkResult(err)
		})
	}
}
