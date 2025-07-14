package botinput

// SharedUsersMessage represents a message with information about shared user
type SharedUsersMessage interface {
	Message
	GetRequestID() int
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
	GetFileUniqueID() string
	GetWidth() int
	GetHeight() int
	GetFileSize() int
}
