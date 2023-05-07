package botsfw

import "github.com/bots-go-framework/bots-fw-models/botsfwmodels"

// BotRecordsMaker is an interface for making bot records
// This should be implemented by platform adapters
// (for example by https://github.com/bots-go-framework/bots-fw-telegram)
type BotRecordsMaker interface {
	MakeBotUserDto() botsfwmodels.BotUser
	MakeBotChatDto() botsfwmodels.BotChat
}
