package botsdal

import (
	"testing"
)

func TestNewBotUserStore(t *testing.T) {
	type args struct {
		//collection string
		//platform string
		//db             DbProvider
		//newBotUserData func(botID string) (botsfwmodels.BotUserData, error)
		//createNewUser  BotUserCreator
	}
	tests := []struct {
		name        string
		args        args
		shouldPanic bool
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
			panic("temporary disabled")
			//_ = newBotUserStore(tt.args.collection, tt.args.platform, tt.args.db, tt.args.newBotUserData, tt.args.createNewUser)
		})
	}
}
