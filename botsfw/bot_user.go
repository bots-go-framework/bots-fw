package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
)

type BotUserCreator func(c context.Context, botID string, apiUser WebhookActor) (botsfwmodels.BotUserData, error)
