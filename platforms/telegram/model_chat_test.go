package telegram_bot

import (
	"testing"
	"google.golang.org/appengine/datastore"
)

func TestTelegramChat(t *testing.T) {
	var _ datastore.PropertyLoadSaver = (*TelegramChatEntity)(nil)
}
