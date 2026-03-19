# Code Review: `bots-go-framework/bots-fw`

**Reviewer:** Senior Backend Engineer (AI-assisted)
**Module:** `github.com/bots-go-framework/bots-fw`
**Go Version:** 1.24.3
**Review Date:** 2026-03-19
**Current Tag:** v0.71.34

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Architecture Overview](#2-architecture-overview)
3. [Package-by-Package Review](#3-package-by-package-review)
   - 3.1 [Root Package (`botsframework`)](#31-root-package-botsframework)
   - 3.2 [`botinput` — Input Abstractions](#32-botinput--input-abstractions)
   - 3.3 [`botmsg` — Output Message Types](#33-botmsg--output-message-types)
   - 3.4 [`botsfwconst` — Platform Constants](#34-botsfwconst--platform-constants)
   - 3.5 [`botsdal` — Data Access Layer](#35-botsdal--data-access-layer)
   - 3.6 [`botsfw` — Core Framework](#36-botsfw--core-framework)
   - 3.7 [`botswebhook` — Webhook Driver & Router](#37-botswebhook--webhook-driver--router)
   - 3.8 [`mocks` — Generated Mocks](#38-mocks--generated-mocks)
4. [Cross-Cutting Concerns](#4-cross-cutting-concerns)
5. [Bugs](#5-bugs)
6. [Security Observations](#6-security-observations)
7. [Test Coverage Assessment](#7-test-coverage-assessment)
8. [Technical Debt Inventory](#8-technical-debt-inventory)
9. [Recommendations (Prioritised)](#9-recommendations-prioritised)

---

## 1. Executive Summary

`bots-go-framework/bots-fw` is a Go framework for building multi-platform chatbots (Telegram, Facebook
Messenger, Viber, WhatsApp). It provides a clean, interface-driven abstraction for input parsing, routing,
context management, and response dispatch.

The architecture is sound and the interface design is generally good. However, the codebase carries
**substantial technical debt**: large volumes of commented-out code, dozens of unresolved TODOs, several
critical unimplemented stubs, at least one routing bug, and weak test coverage in the most complex areas.

**Severity summary:**

| Severity | Count |
|---|---|
| 🔴 Critical (bugs, broken code paths) | 4 |
| 🟠 High (design/correctness concerns) | 7 |
| 🟡 Medium (debt, quality, maintainability) | 12 |
| 🔵 Low (style, naming, polish) | 8 |

---

## 2. Architecture Overview

```
botsframework (root)          ← Module entry point (empty)
├── botinput/                  ← Input interfaces (Entry, User, Message, etc.)
├── botmsg/                    ← Output message types (text, callback, attachments)
├── botsfwconst/               ← Platform type constants
├── botsdal/                   ← DAL helpers (keys, CRUD wrappers)
├── botsfw/                    ← Core: context, settings, routing interfaces
├── botswebhook/               ← Concrete driver + router implementation
└── mocks/                     ← Generated gomock mocks
    ├── mock_botinput/
    ├── mock_botmsg/
    ├── mock_botsfw/
    └── mock_botsfwmodels/
```

**Request flow:**
1. HTTP POST hits a platform-specific handler (implemented outside this module).
2. `botswebhook.webhookDriver.HandleWebhook()` — receives request, calls `GetBotContextAndInputs()`.
3. For each input message, `processWebhookInput()` creates a `WebhookContext`.
4. If the user has no app-user record, a transaction creates one.
5. `webhooksRouter.Dispatch()` matches the input to a registered `Command` and invokes its action.
6. The responder sends the response message back to the platform.

---

## 3. Package-by-Package Review

### 3.1 Root Package (`botsframework`)

**Files:** `package.go`, `package_test.go`

`package.go` contains only a comment: *"Main code for the package is in the `botsfw` directory."*
The module root package exists purely as the Go module name and provides no exports.

**Issues:**

- 🔵 `package_test.go` consists of a single test that calls `t.Log("Package test")` — it tests nothing.
  Delete it or turn it into a real integration smoke test.

---

### 3.2 `botinput` — Input Abstractions

**Purpose:** Defines all input-side interfaces: `Entry`, `InputMessage`, `User`, `Chat`, message
sub-types, payment types, and the `Type` enum.

**Strengths:**
- Well-decomposed interface hierarchy (`Actor → Sender → User`).
- `Type` enum with `go:generate stringer` — generated `type_string.go` includes a compile-time guard
  against stale generation (the blank array index trick).
- Comprehensive coverage of all major bot input types including payments (`SuccessfulPayment`,
  `RefundedPayment`, `PreCheckoutQuery`).

**Issues:**

- 🟡 **`bot_user.go` is entirely commented out.** The file exists with ~60 lines of commented code for
  a `BotUser` interface and `New()` constructor with functional options. This is dead code and should be
  removed. If the design was superseded by `botsdal.BotUser`, document that decision.

- 🟡 **Unimplemented interface stubs.** `StickerMessage`, `VoiceMessage`, `PhotoMessage`, and
  `AudioMessage` all have `// TODO: Define ... message interface` as their only body. They are not
  useful as-is. Either flesh them out or replace with marker interfaces that at least embed `Message`.

- 🟡 **`LogRequest()` on `InputMessage` interface** has an in-code debate (`// TODO: should not be
  part of Input?`). This concern is valid — logging is a cross-cutting concern that doesn't belong on a
  domain interface. Consider a separate `RequestLogger` interface.

- 🔵 **`Actor.Platform()` method** has a TODO questioning whether it should be removed. It introduces
  a circular concern (actors knowing their platform) and should be resolved.

- 🔵 **`payment` interface is unexported** while `SuccessfulPayment`, `RefundedPayment`, and
  `PreCheckoutQuery` depend on it. Since the `payment` base interface is package-private, implementors
  outside the package must redundantly re-declare those methods. Consider exporting `Payment`.

- 🔵 **Commented-out `SuccessfulPayment` struct** at the bottom of `input_interfaces.go` (lines
  ~220-233) — dead code that adds visual noise. Remove it.

---

### 3.3 `botmsg` — Output Message Types

**Purpose:** Defines bot output message types: `BotMessage` interface, text messages, callback answers,
attachments, and the `Type` / `Format` / `AttachmentType` enums.

**Strengths:**
- `AnswerCallbackQuery.Validate()` is thorough: checks required fields, length limits, and URL validity.
- Compile-time interface assertion `var _ BotMessage = AnswerCallbackQuery{}` is good practice.
- JSON tags are consistently applied.

**Issues:**

- 🔴 **`MessageFromBot` design is contradictory.** The struct embeds `TextMessageFromBot` *and* has a
  `BotMessage BotMessage` field. The in-code comment says: *"TODO: This feels wrong and need to be
  refactored! Use BotMessage instead"*. This creates ambiguity: callers do not have a clear, single
  way to set message content. New code creating `MessageFromBot` may set `m.Text` (via the embedded
  struct) while other code uses `m.BotMessage`. Resolving this is important for correctness.

  ```go
  // Current (confusing):
  type MessageFromBot struct {
      TextMessageFromBot          // embedded: sets m.Text, m.Format, ...
      BotMessage BotMessage       // also a message field
      ...
  }

  // Preferred:
  type MessageFromBot struct {
      BotMessage BotMessage
      ResponseChannel BotAPISendMessageChannel
      Analytics analytics.Message
  }
  ```

- 🟡 **`AttachmentType` has a `//goland:noinspection GoUnusedConst` suppress annotation.** If
  `AttachmentTypeNone` is truly unused externally, it should either be used or the design reviewed.
  Suppressing the lint warning hides a potential dead-code smell.

- 🔵 **`NoMessageToSend` sentinel constant** (`"<NO_MESSAGE_TO_SEND>"`) is a string sentinel. A
  dedicated type (e.g., `NoOpMessage{}` implementing `BotMessage`) would be safer and more idiomatic.

---

### 3.4 `botsfwconst` — Platform Constants

**File:** `botsfwconst/paltform.go`

**Issues:**

- 🔵 **Filename typo:** `paltform.go` should be `platform.go`. This is a minor but persistent
  embarrassment in the repository.

- 🟡 **`Platform` is defined as a `string` type with 4 constants** (Telegram, Viber, FBM, WhatsApp).
  WhatsApp support is declared but does not appear to be implemented anywhere in the framework. Consider
  marking it as `// PlatformWhatsApp is reserved, not yet implemented`.

---

### 3.5 `botsdal` — Data Access Layer

**Purpose:** Provides key construction helpers and CRUD wrappers for bot entities: platforms, bots,
bot chats, and platform (bot) users.

**Strengths:**
- Key hierarchy (`platform → bot → botChat` / `platform → botUser`) is clean and hierarchical.
- `CreatePlatformUserRecord` checks for a `Validate()` method on the data object before writing —
  a nice defensive touch.

**Issues:**

- 🔴 **`app_user_store.go` is 100% commented out** (~65 lines). The entire `appUserStore` type, its
  constructor, and all three CRUD methods are commented out. There is no functional code in this file.
  Remove it or restore/replace it.

- 🔴 **`facade_user.go` partial implementation.** `AppUserDal` interface is defined but its
  implementation `appUserDal` and all four methods are commented out with `panic("implement me")`. This
  means `AppUserDal` is an interface with zero concrete implementations in this package. Users of the
  framework must implement it from scratch with no reference implementation.

- 🟠 **`dal_bot_user.go` has ~80 lines of commented-out `botUserStore`** implementation. This was
  apparently a refactoring in progress. The old code should be removed; it creates confusion about the
  current intended design.

- 🟡 **`dal_bot_user_test.go` `TestGetBotUser` passes a nil `ReadwriteTransaction`** to
  `GetPlatformUser()`. This causes a panic (which the test expects), but the panic path through
  `NewPlatformUserKey()` is triggered by an empty platform, not a nil transaction. The test is actually
  testing that an empty `platform` argument panics, but its name suggests it tests `GetBotUser`. The
  test should be renamed and its intent clarified.

- 🟡 **`dal_bot_chat_test.go` `TestNewBotChatStore`** calls `panic("temporary disabled")` directly
  inside the test body. This test always panics and is caught by the `defer/recover` block, making it
  pass trivially while testing nothing. It should be deleted or fixed.

- 🔵 **Collection name constants** (`"botChats"`, `"bots"`, `"botUsers"`, `"botPlatforms"`) are
  unexported package-level `const` strings. If users of the framework need to reference these names (for
  indexes, queries, etc.), they cannot. Consider exporting them or providing accessor functions.

---

### 3.6 `botsfw` — Core Framework

**Purpose:** Central package. Defines `WebhookContext` (the main request context), `BotSettings`,
`BotProfile`, `BotContextProvider`, the `Router` interface, analytics integration, and key error types.

This is the most complex package and has the most issues.

#### `webhook_context_base.go` — `WebhookContextBase`

**Strengths:**
- Context enrichment (locale, chatData, platformUser) is lazy-loaded and cached.
- `NewWebhookContextBase` validates `args.HttpRequest != nil` at construction time.
- `loadChatEntityBase` applies a 1-second timeout on the DB read — defensive and correct.

**Issues:**

- 🔴 **`IsInTransaction` and `NonTransactionalContext` both `panic("not implemented")`** while being
  declared on the `WebhookContextBase` struct (not an interface). Any caller reaching these methods at
  runtime will crash. Either implement them or remove them from the base and document they must be
  provided by platform-specific subclasses.

- 🔴 **`SaveBotUser` is a stub that returns a hardcoded error:**
  ```go
  func (whcb *WebhookContextBase) SaveBotUser(ctx context.Context) error {
      return whcb.db.RunReadwriteTransaction(ctx, func(...) error {
          return errors.New("func SaveBotUser is not implemented yet")
      })
  }
  ```
  This method starts a database transaction and then immediately returns an error. Any call to
  `SaveBotUser` will always fail. This must be fixed before any code paths that save bot user changes
  can work.

- 🔴 **`isAccessGranted` is hardcoded to `true`** when creating a new bot chat from scratch:
  ```go
  // line ~474
  true, // isAccessGranted - TODO: Implement!!!
  ```
  This means every newly created chat is considered access-granted regardless of business logic. This
  is a security/authorization bug.

- 🟠 **`AppUserID()` can trigger side effects silently.** If `appUserID` is empty it calls
  `getPlatformUserRecord()` which performs a DB read. If that fails with a non-not-found error it
  **panics**. Panics in accessor methods make the code hard to reason about.
  ```go
  // In AppUserID():
  if err = whcb.getPlatformUserRecord(whcb.db); err != nil {
      if !dal.IsNotFound(err) {
          panic(fmt.Errorf("failed to get bot user entity: %w", err))
      }
  }
  ```
  Consider returning an error or using the `(value, error)` return pattern.

- 🟠 **`ChatData()` panics in multiple error paths.** Five separate `panic(...)` calls in
  `ChatData()`. Since `ChatData()` is called frequently and has no error return, callers have no way to
  handle these gracefully.

- 🟠 **`isLoadingChatData` / `isLoadingPlatformUserData` boolean flags** guard against re-entrant
  loading. The code itself comments *"TODO: This smells bad. Needs refactoring?"*. These flags are not
  goroutine-safe. If `WebhookContextBase` is ever accessed concurrently (e.g., in a goroutine spawned
  during request handling), this is a data race.

- 🟡 **`EnvLocal` and `EnvProduction` are `var`, not `const`:**
  ```go
  var EnvLocal = "local"
  var EnvProduction = "production"
  ```
  Mutable package-level strings functioning as constants should be `const`. Any package could mutate
  them, silently breaking environment detection logic.

- 🟡 **`SetChatID` is exported** (comment says *"TODO: Should it be private?"*). It provides direct
  mutation of internal state and should be unexported or protected.

- 🟡 **`WebhookContextBase.createPlatformUserRecord`** passes `botUserID` as both `botUserID` and
  `appUserID`:
  ```go
  whcb.recordsFieldsSetter.SetBotUserFields(..., botUserID, botUserID)
  //                                               ^ botUserID  ^ appUserID
  ```
  Setting `appUserID = botUserID` before an app user has actually been created is likely incorrect.

- 🔵 **Multiple methods marked `// TODO: remove`** (`Chat()`, `GetRecipient()`, `GetTime()`,
  `InputType()`). These are present on `WebhookContextBase` but all delegate to `whcb.input`. If they
  are to be removed they should be, or the intent should be documented.

#### `webhook_context.go` — `WebhookContext` interface

- 🟡 **`WebhookContext` is a very large interface** (~35 methods). The in-code comment acknowledges
  this: *"TODO: Make interface much smaller?"*. Large interfaces are harder to mock, implement, and
  understand. It should be split into focused sub-interfaces (e.g., `WebhookIO`, `WebhookI18n`,
  `WebhookUserContext`).

- 🟡 **`WebhookInlineQueryContext`** (defined in the same file) is an empty interface — it adds no
  contract. Remove it or define its methods.

- 🟡 **`ExecutionContext`** only provides `Context() context.Context`. `WebhookContext` already
  embeds `Context()`. Its added value is unclear and matches the TODO comment. Remove or document.

#### `webhook_context_dummy.go`

- 🟡 **`whContextDummy`** exists *"only to check what is NOT implemented by WebhookContextBase"*. It
  panics on every method. This is effectively a compile-time check disguised as a runtime type. A
  better approach: create a `_test.go` file with `var _ WebhookContext = (*WebhookContextBase)(nil)`
  and resolve the gaps one at a time.

#### `bot_profile.go`

- 🟡 **`NewBotProfile` does not validate `router`** (checks only `id`, `newBotChatData`,
  `newBotUserData`). A nil router will not panic at construction time but will panic later when
  `Profile.Router()` is called.

- 🔵 **`NewBotProfile` silently adds the default locale** to `supportedLocales` if missing. This is
  useful behavior but it is undocumented.

#### `settings.go`

- 🟠 **`NewBotSettings` reads `os.Getenv`** at construction time to fill in missing `token` and
  `gaToken`. This tightly couples configuration to environment variables in a non-obvious way.
  Constructors should not have side effects. Consider a separate `BotSettingsFromEnv()` factory or
  a `BotSettings.Validate()` method that callers call explicitly.

- 🟡 **`BotSettingsBy` TODO comment** says it should probably use `map[string]*BotSettings` instead
  of `map[string]BotSettings`. Given that bots are looked up by code/ID on every request, using
  pointer maps avoids redundant copies. This should be resolved.

#### `context_auth.go`

- 🟡 **`SetAccessGranted`** calls `botsdal.GetPlatformUser` outside of a transaction, and then reads
  the record again inside the transaction. This is a classic TOCTOU (time-of-check/time-of-use)
  pattern that can cause stale reads. Refactor to perform the full read-modify-write inside the
  transaction only.

#### `context_new.go`

- 🔵 **`WebhookNewContext`** is a struct containing `BotContext` and `botinput.InputMessage` with
  the comment *"TODO: needs to be checked & described"*. It has no methods and is not used in any
  visible code path. Determine its purpose and document it, or delete it.

---

### 3.7 `botswebhook` — Webhook Driver & Router

**Purpose:** Concrete implementation of the `WebhookDriver` (HTTP request handling) and `Router`
(command matching and dispatch) interfaces defined in `botsfw`.

#### `driver.go`

**Strengths:**
- Panic recovery in `processWebhookInput` with stack trace logging and user notification is robust.
- Per-input error handling distinguishes callback queries (which must return HTTP 200) from other types.
- Environment guard (`isRunningLocally` vs. production) prevents accidental cross-environment calls.

**Issues:**

- 🟠 **`invalidContextOrInputs` does not write an HTTP response when `err != nil` and the error is not
  `ErrAuthFailed`**. In that path the function returns `true` causing `HandleWebhook` to return
  without sending any response to the bot platform. The platform will likely retry, creating an
  amplification loop.
  ```go
  if err != nil {
      var errAuthFailed botsfw.ErrAuthFailed
      if errors.As(err, &errAuthFailed) {
          http.Error(w, ..., http.StatusForbidden)
      }
      // ← Missing: http.Error(w, ..., http.StatusInternalServerError) for other errors
      return true
  }
  ```

- 🟠 **`reportPanicToAnalytics` logs "temporary disabled" but still enqueues the analytics event:**
  ```go
  func (webhookDriver) reportPanicToAnalytics(...) {
      log.Warningf(c, "reportPanicToAnalytics() is temporary disabled")
      err := fmt.Errorf("%s", messageText)
      msg := analytics.NewErrorMessage(err)
      whc.Analytics().Enqueue(msg) // ← still executes!
  }
  ```
  The "disabled" comment is misleading. If the intent is to re-enable it, remove the warning. If the
  intent is to disable it, remove the `Enqueue` call.

- 🟡 **`isRunningLocally`** hardcodes ngrok domain suffixes (`.ngrok.io`, `.ngrok.dev`, etc.). The
  TODO marks this as needing customisation. Until then, any other tunnelling solution (e.g., localtunnel,
  cloudflare tunnel) will be misclassified as production.

- 🟡 **`AnalyticsSettings.GaTrackingID`** is defined but its TODO says *"Refactor to list of analytics
  providers"* — implying it is not yet wired up. The field appears to be unused in the driver body
  (only `Enabled func(r *http.Request) bool` is checked).

#### `router.go`

**Strengths:**
- Comprehensive command matching: by code, exact text, translated text, command prefix, awaiting-reply
  state, and custom matcher functions.
- Registration-time panics for mis-configured commands (missing action, duplicate codes) catch bugs early.
- Inline query and chosen-inline-result routing are correctly separated from text routing.

**Issues:**

- 🔴 **`LocationAction` registration bug in `RegisterCommands`.**
  For commands that have `InputTypes` set, `locationAdded` is declared but **never set to `true`** in
  the `TypeLocation` switch case. Moreover, the post-loop check uses `locationAdded` (positive) instead
  of `!locationAdded` (negation), which is the pattern used for all other action types:
  ```go
  // All other actions correctly use !<type>Added:
  if command.TextAction != nil && !textAdded { addCommand(TypeText, command) }
  if command.CallbackAction != nil && !callbackAdded { addCommand(TypeCallbackQuery, command) }
  // ...

  // Location is broken — uses positive (always false since locationAdded is never set true):
  if command.LocationAction != nil && locationAdded {   // ← BUG: should be !locationAdded
      addCommand(botinput.TypeLocation, command)
  }
  ```
  A command with `LocationAction` set but `TypeLocation` not in `InputTypes` will silently never be
  registered for location inputs.

- 🟡 **`DispatchInlineQuery` method on `webhooksRouter` always panics:**
  ```go
  func (whRouter *webhooksRouter) DispatchInlineQuery(responder botsfw.WebhookResponder) {
      panic(fmt.Errorf("not implemented, responder: %+v", responder))
  }
  ```
  This is declared as a method but not in any interface, making it an orphaned stub.

- 🟡 **`changeLocaleIfLangPassed`** contains hardcoded language normalisations:
  ```go
  case "en-EN":
      lang = "en-US"
  case "fa-FA":
      lang = "fa-IR"
  ```
  These are data patches for specific broken inputs. The correct fix is to validate and normalise locale
  codes at the input source. Framework-level locale fixups are brittle and hard to discover.

- 🟡 **Telegram-specific string matching in framework router:**
  ```go
  case strings.Contains(errText, "message is not modified"):
  case strings.Contains(errText, "message to edit not found"):
  ```
  These are Telegram API error strings inside the platform-agnostic router. The TODO acknowledges this.
  These checks must be moved to the Telegram platform adapter.

- 🟡 **`logInputDetails` sends a user-facing message** ("Unknown Type=%d") unconditionally after
  logging. This mixes logging with response generation. A user receives a confusing error message for
  any unhandled input type, even for types the application deliberately ignores.

- 🟡 **`matchCallbackCommands` logs at `Errorf` level for no-match** — this produces error noise for
  every unrecognised callback data string. `Warningf` or `Debugf` is more appropriate since no-match
  is a recoverable condition.

- 🔵 **`AddCommands` and `AddCommandsGroupedByType` are `Deprecated`** in their own doc comments but
  remain exported and tested. If `RegisterCommands` is the canonical API, the deprecated variants
  should be removed in a future breaking release (v1.0), or at a minimum link to the replacement in
  their godoc.

---

### 3.8 `mocks` — Generated Mocks

Generated with `go.uber.org/mock`. All mocks follow the standard gomock pattern correctly.

**Issues:**

- 🔵 **No `go:generate` directives found in the source interfaces** that produced these mocks.
  Without `//go:generate` annotations on the source types, regenerating mocks requires tribal knowledge
  of the exact `mockgen` commands. The generated file headers contain the commands — but they should
  also be on the source interfaces.

- 🔵 **`mock_botsfwmodels` is in the `mocks/` directory** but the model definitions are in the
  external `bots-fw-store` module. Mocks for an external module's interfaces inside this repo are
  acceptable but should be clearly documented.

---

## 4. Cross-Cutting Concerns

### Error Handling

The codebase uses **three different error handling patterns inconsistently:**

1. **Panic** — used for programming errors (nil arguments at construction) — acceptable.
2. **Panic** — used for runtime errors in request handling (e.g., `ChatData()`, `AppUserID()`) — **not
   acceptable** in a framework that recovers panics itself.
3. **Return `error`** — the Go idiomatic pattern — used in some places.

The mixed use of panic and error returns in `WebhookContextBase` makes it impossible for callers to
write safe, defensive code. The rule should be:
- Panics only in constructors / registration-time (`init`, `NewX`, `Register`).
- Error returns in all request-handling paths.

### Large Amounts of Commented-Out Code

The following files contain significant commented-out code representing dead or unfinished features:

| File | Estimated dead lines |
|---|---|
| `botsdal/app_user_store.go` | ~65 lines (100% of file) |
| `botsdal/dal_bot_user.go` | ~80 lines |
| `botsdal/facade_user.go` | ~30 lines |
| `botinput/bot_user.go` | ~60 lines (100% of file) |
| `botsfw/context_auth.go` | ~50 lines |
| `botsfw/webhook_context_base.go` | ~100 lines |
| `botsfw/webhook_context.go` | ~15 lines |
| `botmsg/message_from_bot.go` | ~5 lines |

**Total: ~400+ lines of dead code.** This severely hurts readability. If the code is no longer needed,
delete it — that is what version control is for.

### TODOs and FIXMEs

There are **50+ TODO comments** in production code (not counting tests or mocks). Several are design-
critical (security, missing features) and many have existed for multiple version bumps. Selected
high-priority ones:

| Location | TODO |
|---|---|
| `webhook_context_base.go:474` | `isAccessGranted - TODO: Implement!!!` |
| `botmsg/message_from_bot.go:63` | `TODO: This feels wrong and need to be refactored!` |
| `botswebhook/router.go:749` | `TODO: Edited messages are not supported` |
| `botswebhook/router.go:809` | Telegram-specific error check — should be in TG adapter |
| `botsfw/webhook_context.go:24` | `TODO: Make interface much smaller?` |

These should be converted into tracked GitHub issues, not left as code comments.

### Naming Inconsistencies

- `botChat` vs `chat` vs `Chat` — used interchangeably as variable names.
- `BotContext` vs `BotContextProvider` — `BotContext` is a struct (value type), confusingly similar to
  the interface `BotContextProvider`. Consider `BotRuntimeContext` for the struct.
- `whc`, `whcb`, `c`, `ctx` — context variable names are inconsistent across the codebase.
- Method `GetBotCode()` vs field access `BotSettings.Code` — the getter and direct field access are
  mixed. Since `BotSettings` is a public struct, consistent field access would be simpler.

---

## 5. Bugs

### BUG-1: `LocationAction` never registered via `RegisterCommands` (🔴 Critical)

**File:** `botswebhook/router.go`
**Impact:** Commands that set `LocationAction` but do not include `TypeLocation` in `InputTypes` are
silently not registered for location input. Location commands will never be dispatched.

```go
// Fix: set locationAdded = true in the switch, and use !locationAdded in the post-loop guard
case botinput.TypeLocation:
    if command.LocationAction == nil && command.Action == nil {
        panic(...)
    }
    locationAdded = true  // ← add this
// ...
if command.LocationAction != nil && !locationAdded {  // ← change && locationAdded to && !locationAdded
    addCommand(botinput.TypeLocation, command)
}
```

### BUG-2: `isAccessGranted` hardcoded `true` for new chats (🔴 Critical)

**File:** `botsfw/webhook_context_base.go`, line ~474
**Impact:** Every newly created bot chat is marked as access-granted. Any access control logic relying
on `IsAccessGranted()` is bypassed for first-time users.

### BUG-3: `invalidContextOrInputs` leaves response blank on non-auth errors (🔴 Critical)

**File:** `botswebhook/driver.go`
**Impact:** When `GetBotContextAndInputs` returns a non-`ErrAuthFailed` error, the HTTP response is
never written. The bot platform receives a connection close or timeout, may log it as an error, and
will likely retry the webhook delivery, potentially causing repeated failures.

### BUG-4: `SaveBotUser` always returns an error (🔴 Critical)

**File:** `botsfw/webhook_context_base.go`
**Impact:** Any code that calls `whc.SaveBotUser(ctx)` will fail with "func SaveBotUser is not
implemented yet". Changes to bot user data cannot be persisted through this method.

---

## 6. Security Observations

- 🔵 **`BotSettings.Token` is stored as a plain string** in the struct. If this struct is ever
  serialised (e.g., for logging, caching, or JSON marshalling via `AppContext`), the token would leak.
  Consider a custom `String()` method that returns a redacted representation.

- 🟡 **`BotSettings.PaymentToken` and `PaymentTestToken`** are stored in plain text in memory. Same
  concern as above.

- 🟡 **`AnswerCallbackQuery.Validate()` validates the URL** with `url.Parse()`. `url.Parse()` is
  permissive (it accepts relative URLs, bare schemes, etc.). If the URL is used to open a web page in
  the user's Telegram client, the validation should be stricter (require `https://` scheme).

---

## 7. Test Coverage Assessment

| Package | Tests | Quality |
|---|---|---|
| Root package | `TestPackage` — trivial | ❌ No value |
| `botinput` | None | ❌ |
| `botmsg` | None | ❌ |
| `botsfwconst` | None | ❌ |
| `botsdal` | `TestNewBotChatStore` (broken), `TestGetBotUser` (shallow), `TestCreateBotUserRecord` (shallow) | ❌ |
| `botsfw` | `TestCommand_DefaultTitle`, `TestCommand_TitleByKey`, translator tests, `TestNewWebhookContextBase` | ⚠️ Partial |
| `botswebhook` | Router construction/registration tests, settings tests | ⚠️ Partial |

**Critical gaps:**
- Zero tests for `WebhookContextBase` (the most complex class).
- Zero tests for `webhooksRouter.Dispatch()` — the entire routing logic is untested.
- Zero tests for `webhookDriver.processWebhookInput()` — the full request pipeline is untested.
- Zero tests for `context_auth.go` — `SetAccessGranted` is untested.
- `botsdal.TestNewBotChatStore` always panics with `panic("temporary disabled")` and thus tests nothing.

The test suite passes (`ok` for all packages) but this masks a near-zero coverage of business logic.

---

## 8. Technical Debt Inventory

| Item | Location | Effort |
|---|---|---|
| Remove ~400 lines of commented-out code | Multiple files | Low |
| Fix `locationAdded` routing bug | `botswebhook/router.go` | Low |
| Fix `isAccessGranted = true` hardcode | `webhook_context_base.go` | Medium |
| Implement `SaveBotUser` | `webhook_context_base.go` | Medium |
| Fix `invalidContextOrInputs` missing response | `botswebhook/driver.go` | Low |
| Implement `IsInTransaction`/`NonTransactionalContext` or remove | `webhook_context_base.go` | Medium |
| Fix `reportPanicToAnalytics` misleading log | `botswebhook/driver.go` | Low |
| Refactor `MessageFromBot` (embedded vs. field) | `botmsg/message_from_bot.go` | High |
| Shrink `WebhookContext` interface | `botsfw/webhook_context.go` | High |
| Move Telegram-specific error strings out of router | `botswebhook/router.go` | Medium |
| Replace `isLoadingChatData` flags with proper sync | `webhook_context_base.go` | Medium |
| Make `EnvLocal`, `EnvProduction` constants | `webhook_context_base.go` | Low |
| Rename `botsfwconst/paltform.go` | `botsfwconst/` | Low |
| Add `go:generate` directives to interface files | Multiple | Low |
| Convert TODO comments to GitHub issues | Everywhere | Low |

---

## 9. Recommendations (Prioritised)

### P0 — Fix immediately (correctness/security)

1. **Fix `isAccessGranted = true` hardcode** — new users should not get access granted automatically.
2. **Fix the `locationAdded` routing bug** in `RegisterCommands`.
3. **Implement or clearly stub `SaveBotUser`** — return an explicit `ErrNotImplemented` sentinel
   (exported, so callers can check with `errors.Is`), not a raw `errors.New` string.
4. **Fix `invalidContextOrInputs`** to write an HTTP error response for all non-auth errors.

### P1 — Fix before next major feature (design correctness)

5. **Refactor `MessageFromBot`** to remove the dual representation (embedded struct + field).
6. **Replace panics in `ChatData()` and `AppUserID()`** with error returns or well-documented panic
   contracts.
7. **Delete all commented-out code** — commit the decision. Use git history to recover if needed.
8. **Implement or remove `IsInTransaction` and `NonTransactionalContext`.**
9. **Move Telegram-specific error strings** to the Telegram adapter package.
10. **Fix `reportPanicToAnalytics`** — remove misleading "disabled" warning.

### P2 — Address in upcoming sprints (quality)

11. **Shrink `WebhookContext` interface** — decompose into 3-4 focused sub-interfaces.
12. **Change `EnvLocal` / `EnvProduction` to `const`.**
13. **Fix `NewBotSettings` env-var side effect** — document or extract into a separate factory.
14. **Add `//go:generate` directives** to all interface files for mock generation.
15. **Rename `botsfwconst/paltform.go`** → `platform.go`.
16. **Write tests for `Dispatch()`** using table-driven tests and the existing mock infrastructure.
17. **Remove or rewrite broken tests** in `botsdal` (the `panic("temporary disabled")` test).

### P3 — Longer-term improvements

18. **Export payment base interface** in `botinput`.
19. **Define body for stub message interfaces** (`StickerMessage`, `VoiceMessage`, etc.).
20. **Add locale normalisation** at input parsing level rather than patching in the router.
21. **Consider thread-safety** of `WebhookContextBase` fields (`isLoadingChatData`, locale, etc.) if
    concurrent access is ever introduced.
22. **Evaluate `BotSettingsBy` pointer maps** — using `map[string]*BotSettings` avoids struct copies
    on every lookup.
