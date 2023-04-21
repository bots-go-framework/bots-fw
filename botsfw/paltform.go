package botsfw

type Platform string

const (
	PlatformTelegram          Platform = "telegram"
	PlatformViber             Platform = "viber"
	PlatformFacebookMessenger Platform = "fbm"
	PlatformWhatsApp          Platform = "whatsapp"
)
