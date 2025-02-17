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
	User     botinput.WebhookUser
}

type AppUserDal interface {

	// CreateAppUserFromBotUser creates app user record using bot user data
	CreateAppUserFromBotUser(ctx context.Context, tx dal.ReadwriteTransaction, bot Bot) (
		appUser record.DataWithID[string, botsfwmodels.AppUserData],
		botUser BotUser,
		err error,
	)

	//GetAppUserByBotUserID(ctx context.Context, tx dal.ReadwriteTransaction, platform, botID, botUserID string) (
	//	appUser record.DataWithID[string, botsfwmodels.AppUserData], err error,
	//)
	//UpdateAppUser(ctx context.Context, tx dal.ReadwriteTransaction, appUser record.DataWithID[string, botsfwmodels.AppUserData]) error
	//LinkAppUserToBotUser(ctx context.Context, platform, botID, botUserID, appUserID string) (err error)
}

//type appUserDal struct {
//}
//
//func DefaultAppUserDal() AppUserDal {
//	return appUserDal{}
//}
//
//func (a appUserDal) CreateAppUserFromBotUser(ctx context.Context, tx dal.ReadwriteTransaction, platform, botID string, botUser botinput.WebhookUser) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (a appUserDal) GetAppUserByBotUserID(ctx context.Context, tx dal.ReadwriteTransaction, platform, botID, botUserID string) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (a appUserDal) UpdateAppUser(ctx context.Context, tx dal.ReadwriteTransaction, appUser record.DataWithID[string, botsfwmodels.AppUserData]) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (a appUserDal) LinkAppUserToBotUser(ctx context.Context, platform, botID, botUserID, appUserID string) (err error) {
//	//TODO implement me
//	panic("implement me")
//}
