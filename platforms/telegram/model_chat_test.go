package telegram

import (
	"google.golang.org/appengine/datastore"
	"testing"
)

func TestTelegramChat(t *testing.T) {
	var _ datastore.PropertyLoadSaver = (*TgChatEntityBase)(nil)
}
