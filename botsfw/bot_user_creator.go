package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
)

type BotUserCreator func(c context.Context, botID string, apiUser botinput.WebhookActor) (botsfwmodels.PlatformUserData, error)
