package botsfw

import (
	"github.com/strongo/app/user"
)

// BotUser interface provides information about bot user
type BotUser interface {
	// GetAppUserIntID returns app user ID
	// Deprecated: use GetAppUserStrID instead
	GetAppUserIntID() int64 // TODO: decommission?
	GetAppUserStrID() string
	IsAccessGranted() bool
	SetAccessGranted(value bool) bool
	SetAppUserIntID(appUserID int64)
	user.UpdatedTimeSetter // SetUpdatedTime(time.Time) // to satisfy github.com/strongo/app/user.UpdatedTimeSetter
}
