package botinput

//// BotUser provides info about current bot user
//type BotUser interface {
//	// GetBotUserID returns bot user ID
//	GetBotUserID() string
//
//	// GetFirstName returns user's first name
//	GetFirstName() string
//
//	// GetLastName returns user's last name
//	GetLastName() string
//}
//
//func New(botUserID string, fields ...func(user *botUser)) BotUser {
//	svr := &botUser{botUserID: botUserID}
//	for _, f := range fields {
//		f(svr)
//	}
//	return svr
//}
//
//var _ BotUser = (*botUser)(nil)
//
//type botUser struct {
//	botUserID string
//	firstName string
//	lastName  string
//}
//
//func (v *botUser) GetBotUserID() string {
//	return v.botUserID
//}
//
//func (v *botUser) GetFirstName() string {
//	return v.firstName
//}
//
//func (v *botUser) GetLastName() string {
//	return v.lastName
//}
//
//func WithFirstName(s string) func(user *botUser) {
//	return func(v *botUser) {
//		v.firstName = s
//	}
//}
//
//func WithLastName(s string) func(user *botUser) {
//	return func(v *botUser) {
//		v.lastName = s
//	}
//}
