package botsfw

// BotRecordsMaker is an interface for making bot records
// This should be implemented by platform adapters
// (for example by https://github.com/bots-go-framework/bots-fw-telegram)
type BotRecordsMaker interface {
	MakeBotUserDto() BotUser
	MakeBotChatDto() BotChat
}
