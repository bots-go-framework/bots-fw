package telegram_bot

import (
	"google.golang.org/appengine/datastore"
	"testing"
)

func TestTelegramUser(t *testing.T) {
	var _ datastore.PropertyLoadSaver = (*TelegramUserEntity)(nil)
}
