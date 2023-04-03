package botsfw

import "testing"

func TestNewChatID(t *testing.T) {
	type args struct {
		botID     string
		botChatID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", ""}, ""},
		{"should_pass", args{"b1", "chat1"}, "b1:chat1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == "" {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewChatID() did not panic")
					}
				}()
			}
			if got := NewChatID(tt.args.botID, tt.args.botChatID); got != tt.want {
				t.Errorf("NewChatID() = %v, want %v", got, tt.want)
			}
		})
	}
}
