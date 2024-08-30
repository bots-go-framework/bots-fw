package botsfsobject

// BotUser provides info about current bot user
type BotUser interface {
	// GetBotUserID returns bot user ID
	GetBotUserID() string

	// GetFirstName returns user's first name
	GetFirstName() string

	// GetLastName returns user's last name
	GetLastName() string
}
