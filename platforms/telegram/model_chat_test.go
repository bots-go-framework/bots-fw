package telegram_bot

import (
	"google.golang.org/appengine/datastore"
	"testing"
)

func TestTelegramChat(t *testing.T) {
	var _ datastore.PropertyLoadSaver = (*TelegramChatEntityBase)(nil)
}
