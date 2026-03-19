# Proposal: Decompose `WebhookContext` Interface

**Status:** Draft  
**Related:** `CODE_REVIEW.md` — Section 3.6, item "WebhookContext is a very large interface"

---

## 1. Problem Statement

`WebhookContext` currently declares **36 methods** (including those from embedded interfaces):

```go
type WebhookContext interface {
    Environment() string
    BotInputProvider                                              // → Input()
    BotPlatform() BotPlatform
    Request() *http.Request
    Context() context.Context
    SetContext(c context.Context)
    ExecutionContext() ExecutionContext
    AppContext() AppContext
    BotContext() BotContext
    MustBotChatID() string
    GetBotCode() string
    GetBotSettings() *BotSettings
    DB() dal.DB
    ChatData() botsfwmodels.BotChatData
    GetBotUser() (botsdal.BotUser, error)
    GetBotUserForUpdate(ctx context.Context, tx dal.ReadwriteTransaction) (botsdal.BotUser, error)
    GetBotUserID() string
    IsInGroup() (bool, error)
    CommandText(title, icon string) string
    SetLocale(code5 string) error
    NewMessage(text string) botmsg.MessageFromBot
    NewMessageByCode(messageCode string, a ...interface{}) botmsg.MessageFromBot
    NewEditMessage(text string, format botmsg.Format) (botmsg.MessageFromBot, error)
    UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error
    AppUserID() string
    SetUser(id string, data botsfwmodels.AppUserData)
    AppUserData() (botsfwmodels.AppUserData, error)
    BotState                                                      // → IsNewerThen()
    SaveBotChat() error
    RecordsFieldsSetter() BotRecordsFieldsSetter
    i18n.SingleLocaleTranslator                                   // → Locale, Translate, TranslateNoWarning, TranslateWithMap
    GetTranslator(locale string) i18n.SingleLocaleTranslator
    Responder() WebhookResponder
    Analytics() WebhookAnalytics
}
```

This causes several concrete problems:

| Problem | Impact |
|---|---|
| **Implementing** the interface requires 36 methods | Platform adapters must stub out everything even if the platform only supports a subset |
| **Mocking** the interface is heavyweight | Tests that only care about one concern must stub 36 methods |
| **Helper functions** must accept `WebhookContext` even when they only use 2–3 methods | Tight coupling, harder to unit-test helpers in isolation |
| **Discoverability** is poor | Developers cannot tell at a glance which methods relate to routing vs. user data vs. i18n |
| **Incremental implementation** is impossible | A new platform adapter cannot pass `go build` until every method is implemented |

---

## 2. Callers — What They Actually Need

Analysing all non-test call sites of `WebhookContext` in the codebase:

| Caller | Methods actually used |
|---|---|
| `Command.DefaultTitle` / `TitleByKey` | `Translate()`, `CommandText()` |
| `CommandMatcher` | `ChatData()`, `Input()`, `Context()` |
| `SetAccessGranted` | `Context()`, `GetBotCode()`, `ChatData()`, `Input()`, `SaveBotChat()`, `DB()`, `BotPlatform()`, `BotContext()` |
| `matchMessageCommands` | `Context()`, `ChatData()`, `TranslateNoWarning()`, `DefaultTitle()` |
| `matchCallbackCommands` | `Context()`, `Input().LogRequest()` |
| `changeLocaleIfLangPassed` | `Context()`, `ChatData()`, `Locale()`, `SetLocale()` |
| `logInputDetails` | `Context()`, `Input()`, `Responder()`, `NewMessage()` |
| `processCommandResponse` | `Context()`, `Input()`, `Responder()`, `Analytics()` |
| `processCommandResponseError` | `Context()`, `Input()`, `Responder()`, `NewMessage()`, `Analytics()` |
| `reportPanicToAnalytics` | `Analytics()` |
| `router.Dispatch` | `Context()`, `Input()`, `ChatData()`, `Locale()`, `SetLocale()`, `Responder()`, `NewMessage()`, `TranslateNoWarning()`, `Analytics()` |

No single helper uses all 36 methods. The maximum used by one function is ~10.

---

## 3. Proposed Sub-Interfaces

The interface is decomposed into **6 focused sub-interfaces** plus a **compose-all** `WebhookContext`.
The boundaries follow the Go proverb: *"the bigger the interface, the weaker the abstraction."*

---

### 3.1 `WebhookRequestContext` — Identity & Infrastructure

Everything a function needs to know *where* and *in which environment* it is running.
This is the foundation: almost every helper that does logging or DB work needs it.

```go
// WebhookRequestContext provides identity and infrastructure access for the current request.
type WebhookRequestContext interface {
    // Context returns the Go context for this request.
    Context() context.Context

    // SetContext replaces the request context (e.g. after adding values or a deadline).
    SetContext(c context.Context)

    // Request returns the raw HTTP request.
    Request() *http.Request

    // Environment returns the deployment environment (e.g. "local", "production").
    Environment() string

    // BotPlatform returns the platform this request arrived on (Telegram, Viber, FBM, …).
    BotPlatform() BotPlatform

    // BotContext returns settings and host information for the current bot.
    BotContext() BotContext

    // GetBotCode is a convenience shortcut for BotContext().BotSettings.Code.
    GetBotCode() string

    // GetBotSettings is a convenience shortcut for BotContext().BotSettings.
    GetBotSettings() *BotSettings

    // DB returns the database handle assigned to this bot.
    DB() dal.DB

    // AppContext returns application-level context (i18n, DAL, etc.).
    AppContext() AppContext
}
```

---

### 3.2 `WebhookInputContext` — Incoming Message

Everything about the *message that arrived*.

```go
// WebhookInputContext provides access to the incoming message from the user.
type WebhookInputContext interface {
    // Input returns the parsed incoming message.
    Input() botinput.InputMessage

    // GetBotUserID returns the platform-specific user ID of the sender as a string.
    GetBotUserID() string

    // MustBotChatID returns the chat ID or panics if it cannot be determined.
    MustBotChatID() string

    // IsInGroup reports whether the message was received in a group chat.
    IsInGroup() (bool, error)
}
```

---

### 3.3 `WebhookUserData` — User & Chat Persistence

Everything related to loading and saving the persistent state of the current user and chat.
This is the most data-access-heavy sub-interface and is naturally needed together.

```go
// WebhookUserData provides read/write access to the persistent state of the current
// bot user, app user, and chat.
type WebhookUserData interface {
    // ChatData returns the current bot chat's persistent data.
    // Returns nil for input types that have no associated chat (e.g. InlineQuery).
    ChatData() botsfwmodels.BotChatData

    // SaveBotChat persists the current chat data to the database.
    SaveBotChat() error

    // GetBotUser returns the current platform user record.
    GetBotUser() (botsdal.BotUser, error)

    // GetBotUserForUpdate returns the platform user record inside a write transaction.
    GetBotUserForUpdate(ctx context.Context, tx dal.ReadwriteTransaction) (botsdal.BotUser, error)

    // AppUserID returns the application-layer user ID linked to this bot user.
    AppUserID() string

    // SetUser caches the resolved app user ID and data into the context.
    // Called by the driver after app-user creation.
    SetUser(id string, data botsfwmodels.AppUserData)

    // AppUserData loads and returns the app user's persistent data.
    AppUserData() (botsfwmodels.AppUserData, error)

    // RecordsFieldsSetter returns the helper used to populate new bot/chat/user records.
    RecordsFieldsSetter() BotRecordsFieldsSetter

    // UpdateLastProcessed records the message sequence number / timestamp on the chat entity.
    UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error

    // IsNewerThen reports whether the current message is newer than the chat entity's
    // last-processed sequence number (used to detect and discard duplicate deliveries).
    IsNewerThen(chatEntity botsfwmodels.BotChatData) bool
}
```

> **Note:** `IsNewerThen` (currently on embedded `BotState`) is moved directly into this interface
> because it is always used alongside `ChatData()`. The standalone `BotState` interface can be kept
> for backwards compatibility but marked deprecated.

---

### 3.4 `WebhookI18n` — Internationalisation

Translation, locale management, and command-text formatting.

```go
// WebhookI18n provides localisation support for the current request.
type WebhookI18n interface {
    // Embeds the standard single-locale translator (Locale, Translate, TranslateNoWarning, TranslateWithMap).
    i18n.SingleLocaleTranslator

    // SetLocale switches the active locale for this request.
    SetLocale(code5 string) error

    // GetTranslator returns a translator pinned to the given locale code.
    // Useful when rendering content in a locale other than the active one.
    GetTranslator(locale string) i18n.SingleLocaleTranslator

    // CommandText formats a command title and icon into a display string,
    // translating the title key if needed.
    CommandText(title, icon string) string
}
```

---

### 3.5 `WebhookMessaging` — Message Construction & Dispatch

Creating outgoing messages and sending them.

```go
// WebhookMessaging provides helpers to construct and send messages back to the user.
type WebhookMessaging interface {
    // NewMessage creates a plain-text MessageFromBot.
    NewMessage(text string) botmsg.MessageFromBot

    // NewMessageByCode creates a MessageFromBot from an i18n key, formatting it with args.
    NewMessageByCode(messageCode string, a ...interface{}) botmsg.MessageFromBot

    // NewEditMessage creates a MessageFromBot that edits the previously sent message.
    NewEditMessage(text string, format botmsg.Format) (botmsg.MessageFromBot, error)

    // Responder returns the WebhookResponder used to deliver messages to the platform.
    Responder() WebhookResponder
}
```

---

### 3.6 `WebhookTelemetry` — Analytics

```go
// WebhookTelemetry provides access to the analytics pipeline.
type WebhookTelemetry interface {
    Analytics() WebhookAnalytics
}
```

---

### 3.7 `WebhookContext` — Full Composition (unchanged external API)

`WebhookContext` remains the single type used by all command action functions. It becomes a pure
composition of the six sub-interfaces with **no new methods of its own**:

```go
// WebhookContext is the full request context passed to every command action handler.
// It is a composition of focused sub-interfaces. Prefer accepting the narrowest
// sub-interface that covers your function's actual needs.
type WebhookContext interface {
    WebhookRequestContext
    WebhookInputContext
    WebhookUserData
    WebhookI18n
    WebhookMessaging
    WebhookTelemetry
}
```

`ExecutionContext` is **removed** from the composition (it wraps `Context()` which is already on
`WebhookRequestContext`, and adds no independent value). For backwards compatibility it can be kept
as a deprecated alias:

```go
// Deprecated: use WebhookRequestContext directly.
type ExecutionContext = WebhookRequestContext
```

---

## 4. Refactoring Internal Helpers

Internal functions in `botsfw` and `botswebhook` that currently accept `WebhookContext` can be
narrowed to accept only what they use. This makes them easier to test in isolation.

### Before → After examples

```go
// BEFORE (botswebhook/driver.go)
func (webhookDriver) reportPanicToAnalytics(
    c context.Context, whc botsfw.WebhookContext, messageText string)

// AFTER — only needs analytics
func (webhookDriver) reportPanicToAnalytics(
    c context.Context, analytics botsfw.WebhookTelemetry, messageText string)
```

```go
// BEFORE (botswebhook/router.go)
func changeLocaleIfLangPassed(
    whc botsfw.WebhookContext, callbackUrl *url.URL) (botmsg.MessageFromBot, error)

// AFTER — only needs i18n + user data
func changeLocaleIfLangPassed(
    whc interface {
        botsfw.WebhookI18n
        botsfw.WebhookUserData
    },
    callbackUrl *url.URL,
) (botmsg.MessageFromBot, error)
```

```go
// BEFORE (botsfw/commands.go)
func (c Command) TitleByKey(key string, whc WebhookContext) string

// AFTER — only needs i18n
func (c Command) TitleByKey(key string, tr WebhookI18n) string
```

```go
// BEFORE (botswebhook/router.go)
func (webhookDriver) logInputDetails(whc botsfw.WebhookContext, isKnownType bool)

// AFTER — needs input + messaging
func (webhookDriver) logInputDetails(
    whc interface {
        botsfw.WebhookInputContext
        botsfw.WebhookMessaging
    },
    isKnownType bool,
)
```

For cases where several sub-interfaces are needed repeatedly, a lightweight named composition alias
can be introduced in the router's internal file (not exported):

```go
// routerContext is the subset of WebhookContext used by the router's internal helpers.
type routerContext interface {
    botsfw.WebhookRequestContext
    botsfw.WebhookInputContext
    botsfw.WebhookUserData
    botsfw.WebhookI18n
    botsfw.WebhookMessaging
    botsfw.WebhookTelemetry
}
```

---

## 5. Impact on Platform Adapters

Platform adapters (e.g., `bots-fw-telegram`) embed `*WebhookContextBase`. After the change:

- `WebhookContextBase` continues to implement all six sub-interfaces (and therefore `WebhookContext`).
- Platform adapters only need to implement methods **not** provided by the base — primarily
  `NewEditMessage`, `IsNewerThen`, `UpdateLastProcessed`, and `Responder`.
- Any adapter that currently stubs out methods it does not support (with `panic("not implemented")`)
  can now be limited to the sub-interface it actually satisfies.

**Platform adapters can now declare partial compliance:**

```go
// TelegramWebhookContext implements WebhookContext.
// Only NewEditMessage, IsNewerThen, UpdateLastProcessed, and Responder require
// platform-specific implementation; all other methods are inherited from WebhookContextBase.
var _ botsfw.WebhookContext = (*TelegramWebhookContext)(nil)
```

---

## 6. Impact on Tests

### Current mocking burden

```go
// Every test helper that receives WebhookContext must set up ~36 EXPECT() calls
// even when only one method is relevant.
whcMock.EXPECT().Context().Return(ctx).AnyTimes()
whcMock.EXPECT().Analytics().Return(analytics).AnyTimes()
// ... 34 more
```

### After decomposition

```go
// Test for reportPanicToAnalytics only needs a WebhookTelemetry mock
ctrl := gomock.NewController(t)
telemetry := mock_botsfw.NewMockWebhookTelemetry(ctrl)
telemetry.EXPECT().Analytics().Return(mockAnalytics)
reportPanicToAnalytics(ctx, telemetry, "some panic")
```

Each sub-interface will have its own small generated mock in `mocks/mock_botsfw/`, produced by adding
`//go:generate` directives to each interface declaration.

---

## 7. Migration Plan

This is a **backwards-compatible** change for users of `WebhookContext` — the composed interface
retains all existing methods. Migration can be done incrementally:

### Phase 1 — Define sub-interfaces (no breaking changes)

1. Add the six sub-interface declarations to `botsfw/webhook_context.go`.
2. Change `WebhookContext` to embed them instead of listing methods directly.
3. Run `go build ./...` and `go test ./...` — everything must pass unchanged.
4. Add `//go:generate` directives and regenerate mocks.

### Phase 2 — Narrow internal helpers (no breaking changes)

5. Change `Command.TitleByKey` / `DefaultTitle` to accept `WebhookI18n`.
6. Change `reportPanicToAnalytics` to accept `WebhookTelemetry`.
7. Change `changeLocaleIfLangPassed` to accept the narrower composition.
8. Change `logInputDetails` to accept `WebhookInputContext` + `WebhookMessaging`.
9. Run tests after each change.

### Phase 3 — Update documentation and remove dead abstractions

10. Remove or deprecate the empty `WebhookInlineQueryContext` interface.
11. Deprecate `ExecutionContext` (it wraps only `Context()`).
12. Deprecate the standalone `BotState` interface (merged into `WebhookUserData`).
13. Update the README and any usage examples.

### Phase 4 — Propagate to platform adapters (coordination with downstream)

14. Update `bots-fw-telegram` and any other platform adapter to use narrower types in their own
    internal helpers.
15. Increment the minor version (this is additive, not breaking).

---

## 8. Summary of Benefits

| Concern | Before | After |
|---|---|---|
| Interface size | 36 methods | 6 sub-interfaces, 3–9 methods each |
| Mocking a single concern | 36 stubs | 3–9 stubs |
| Helper function coupling | Accepts `WebhookContext` (all 36) | Accepts the 2–5 methods it needs |
| Platform adapter burden | Must implement all 36 | Must implement only platform-specific subset |
| Discoverability | One large block | Self-documenting groupings |
| Breaking change | N/A | None — `WebhookContext` is a superset |

---

## 9. Open Questions

1. **`SetUser` visibility.** `SetUser` is called only by `driver.go` after creating an app user.
   It is arguably internal plumbing, not part of the public contract. Should it move to an
   unexported interface or be passed through a constructor argument instead?

2. **`MustBotChatID` vs. `BotChatID`.** `MustBotChatID` panics; `BotChatID` (which returns an error)
   is not on the interface at all. Panicking convenience methods on interfaces are a smell. Consider
   replacing with a non-panicking `BotChatID() (string, error)` and removing `MustBotChatID`, leaving
   a package-level helper `MustBotChatID(ctx WebhookInputContext)` for callers that genuinely want the
   panic behaviour.

3. **`WebhookUserData.RecordsFieldsSetter`.** This returns a framework-internal helper type. It is
   unlikely to be useful to application-layer command handlers. Consider moving it to
   `WebhookRequestContext` (since it is bot-configuration-level) or making it part of
   `WebhookContextBase`'s concrete API only.

4. **`CommandText` placement.** `CommandText` is formatting glue between i18n and command metadata.
   It could live on `WebhookI18n` (as proposed) or on a dedicated `WebhookCommandContext`. The
   simpler option is `WebhookI18n` since it is used primarily in i18n-adjacent code.
