package botsdal

import (
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/record"
)

type BotUser record.DataWithID[string, botsfwmodels.PlatformUserData]
