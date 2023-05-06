package botsfw

import "github.com/strongo/app/user"

// BotUserData hold common properties for bot user entities
type BotUserData struct {
	BotEntity
	user.LastLogin

	FirstName string // required
	LastName  string // optional
	UserName  string // optional
}
