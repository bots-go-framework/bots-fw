package botsfw

import (
	"context"
	"github.com/strongo/i18n"
)

// TranslatorProvider translates texts
type TranslatorProvider func(c context.Context) i18n.Translator
