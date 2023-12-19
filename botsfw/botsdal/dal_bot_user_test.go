package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"testing"
)

func TestGetBotUser(t *testing.T) {
	type args struct {
		//collection string
		platform  string
		botID     string
		botUserID string
	}
	tests := []struct {
		name        string
		args        args
		shouldPanic bool
		checkResult func(botUser record.DataWithID[string, botsfwmodels.BotUserData], err error)
	}{
		{name: "empty", shouldPanic: true},
	}
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
			botUser, err := GetBotUser(nil, tx, tt.args.platform, tt.args.botID, tt.args.botUserID, func() botsfwmodels.BotUserData {
				return nil
			})
			tt.checkResult(botUser, err)
		})
	}
}

func TestCreateBotUserRecord(t *testing.T) {
	type args struct {
		platform    string
		botID       string
		botUserID   string
		botUserData botsfwmodels.BotUserData
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
						t.Errorf("CreateBotUserRecord() did not panic")
					}
				}()
			}
			var tx dal.ReadwriteTransaction
			err := CreateBotUserRecord(ctx, tx, tt.args.platform, tt.args.botID, tt.args.botUserID, tt.args.botUserData)
			tt.checkResult(err)
		})
	}
}
