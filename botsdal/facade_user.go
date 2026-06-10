package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
)

type Bot struct {
	Platform botsfwconst.Platform
	ID       string
	User     botinput.User
}

type AppUserDal interface {
	// CreateAppUserFromBotUser creates app user record using bot user data
	CreateAppUserFromBotUser(ctx context.Context, tx dal.ReadwriteTransaction, bot Bot) (
		appUser record.DataWithID[string, botsfwmodels.AppUserData],
		botUser BotUser,
		err error,
	)
}
