package botsfw

import (
	"context"

	"github.com/strongo/i18n"
)

// AppContext provides application-owned presentation services to bots-fw.
// Persistence is injected separately through BotSettings.Store.
type AppContext interface {
	i18n.LocalesProvider
	GetTranslator(ctx context.Context) i18n.Translator
}
