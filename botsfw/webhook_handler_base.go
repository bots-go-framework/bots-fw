package botsfw

// WebhookHandlerBase provides base implementation for a bot handler
type WebhookHandlerBase struct {
	WebhookDriver
	BotHost
	BotPlatform
	//RecordsMaker        botsfwmodels.BotRecordsMaker
	RecordsFieldsSetter BotRecordsFieldsSetter
	TranslatorProvider  TranslatorProvider
	//DataAccess          botsfwdal.DataAccess
}

// Register driver
func (bh *WebhookHandlerBase) Register(d WebhookDriver, h BotHost) {
	if d == nil {
		panic("WebhookDriver == nil")
	}
	if h == nil {
		panic("BotHost == nil")
	}
	bh.WebhookDriver = d
	bh.BotHost = h
}
