package botinput

// WebhookSharedUserMessage represents a message with information about shared user
type WebhookSharedUserMessage interface {
	WebhookMessage
	GetSharedUsers() []SharedUserMessageItem
}

type SharedUserMessageItem interface {
	GetBotUserID() string
	GetUsername() string
	GetFirstName() string
	GetLastName() string
	GetPhotos() []PhotoMessageItem
}

type PhotoMessageItem interface {
	GetFileID() string
	GetUniqueFileID() string
	GetWidth() int
	GetHeight() int
	GetFileSize() int
}
