package botsfw

import (
	"github.com/strongo/app/user"
)

// BotEntity holds properties common to al bot entities
type BotEntity struct {
	AccessGranted bool
	user.OwnedByUserWithID
}

// IsAccessGranted indicates if access to the bot has been granted
func (e *BotEntity) IsAccessGranted() bool {
	return e.AccessGranted
}

// SetAccessGranted mark that access has been granted
func (e *BotEntity) SetAccessGranted(value bool) bool {
	if e.AccessGranted != value {
		e.AccessGranted = value
		return true
	}
	return false
}
