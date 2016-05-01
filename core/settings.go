package bots

type BotSettings struct {
	Code   string
	Token  string
	VerifyToken string // Used by Facebook
	Locale Locale
}

func NewBotSettings(code, token string, locale Locale) BotSettings {
	if code == "" {
		panic("Missing required parameter: code")
	}
	if token == "" {
		panic("Missing required parameter: token")
	}
	if locale.Code5 == "" {
		panic("Missing required parameter: locale.Code5")
	}
	return BotSettings{
		Code: code,
		Token: token,
		Locale: locale,
	}
}

type BotSettingsBy struct {
	Code     map[string]BotSettings
	ApiToken map[string]BotSettings
	Locale   map[string]BotSettings
}

func NewBotSettingsBy(bots ...BotSettings) BotSettingsBy {
	botsBy := BotSettingsBy{}
	for _, bot := range bots {
		botsBy.Code[bot.Code] = bot
		botsBy.ApiToken[bot.Token] = bot
		botsBy.Locale[bot.Locale.Code5] = bot
	}
	return botsBy
}