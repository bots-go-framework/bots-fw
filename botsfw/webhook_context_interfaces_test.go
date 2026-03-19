package botsfw

// Compile-time checks: WebhookContextBase satisfies sub-interfaces where it provides all methods.
// Note: Some methods like IsNewerThen, NewEditMessage, UpdateLastProcessed, Responder, AppUserData
// are NOT on WebhookContextBase — they must be provided by embedding types (e.g. whContextDummy).
var _ WebhookRequestContext = (*WebhookContextBase)(nil)
var _ WebhookInputContext = (*WebhookContextBase)(nil)
var _ WebhookTelemetry = (*WebhookContextBase)(nil)
var _ WebhookI18n = (*WebhookContextBase)(nil)

// whContextDummy (which embeds *WebhookContextBase) satisfies the full WebhookContext
var _ WebhookContext = (*whContextDummy)(nil)

// TestWebhookContext (from webhook_context_test.go) also satisfies the full WebhookContext
// Already checked in webhook_context_test.go via: var _ WebhookContext = TestWebhookContext{}
