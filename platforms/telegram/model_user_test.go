package telegram_bot

import (
	"testing"
	"google.golang.org/appengine/datastore"
)

func TestTelegramUser(t *testing.T) {
	var _ datastore.PropertyLoadSaver = (*TelegramUserEntity)(nil)
}
