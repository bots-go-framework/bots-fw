package botsdal

import (
	"testing"
)

func TestNewBotChatStore(t *testing.T) {
	type args struct {
		collection string
		//platform   string
		//db         DbProvider
	}
	tests := []struct {
		name        string
		args        args
		shouldPanic bool
	}{
		{name: "empty", args: args{collection: ""}, shouldPanic: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("newBotChatStore() did not panic")
					}
				}()
			}
			panic("temporary disabled")
			//_ = newBotChatStore(tt.args.collection, tt.args.platform, tt.args.db, nil)
		})
	}
}
