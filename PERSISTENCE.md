# Persistence architecture and migration

This is a code-composition migration. **Do not clean or migrate existing bot
records.** The default DALgo adapters deliberately keep the historical keys and
record data unchanged:

- Platform users: `botPlatforms/{platform}/botUsers/{platformUser}`
- Bot chats: `botPlatforms/{platform}/bots/{bot}/botChats/{chat}`
- Telegram callback chat instances:
  `botPlatforms/telegram/chatInstances/{chatInstance}`
- WhatsApp data: `waSubjects`, `waChatData`, `waTemplates`, and
  `waTemplateNames`, with the same IDs as before.

Existing records therefore load normally. There is no data migration phase; an
existing linked platform user is reused, while an old record without an
application-user link is repaired when it next receives a webhook.

## Boundaries

`bots-fw-store/botsfwstore` defines `StateStore`, a small use-case interface.
The core framework only asks it to:

1. resolve/create the framework-owned platform-user, application-user link, and
   chat state (`EnsureLinked`);
2. load/update a platform user; and
3. load an application user or save chat data.

`WebhookContext` has no `DB()` method and does not expose DALgo records or
transactions. Command handlers should use application facades. A façade may use
whatever repository or transaction model the application needs, but it is not a
framework context concern.

`bots-fw-store-dalgo` is the optional default implementation. It owns the DALgo
transaction that persists the platform-user link and chat; the router runs only
after that method returns. This intentionally avoids a general framework unit of
work and prevents message sends or other external effects from occurring inside
a retryable database callback.

## Application identity link

An application supplies `botsfwstoredalgo.AppUserStore` to the DALgo adapter.
This port has two deliberate phases:

1. `PrepareAppUser` runs before the retryable transaction. It may idempotently
   provision an external identity such as Firebase Auth, but must not persist
   the application's DALgo user record.
2. `EnsureAppUser` receives the adapter-owned transaction and persists only the
   prepared application user. The adapter commits that record, the platform
   link, and the chat together.

Both phases must converge on the same deterministic application-user ID for a
platform identity. The adapter re-reads current state in its transaction and
fails with `botsfwstore.ErrIdentityConflict` if concurrent work disagrees,
instead of silently orphaning or relinking a user.

Firebase and Firestore do not share a transaction: a failed database commit may
leave an already-provisioned Firebase identity. That outcome is intentionally
recoverable because preparation is idempotent and uses the same deterministic
ID on retry. Existing database records are reused unchanged.

## Composition

At startup, compose the framework store explicitly:

```go
stateStore := botsfwstoredalgo.NewStateStore(db, appUserStore)

settings := botsfw.NewBotSettings(
    botsfwconst.PlatformTelegram, env, profile, code, id, token, gaToken,
    locale, stateStore,
)

chatInstances := telegramdalgo.NewChatInstanceStore(db, telegram.NewChatInstanceData)
handler := telegram.NewTelegramWebhookHandler(
    botContextProvider, translatorProvider, chatInstances,
)
```

Hosts that resolve DALgo from request or tenant context use
`NewStateStoreWithProvider` and `NewChatInstanceStoreWithProvider` instead.

The analogous WhatsApp ports (`SubjectStore`, `ChatDataStore`, and
`TemplateCatalog`) are implemented by `bots-fw-whatsapp-dalgo`.

## Breaking API changes

- `BotSettings` now receives `botsfwstore.StateStore`, rather than database and
  application-user getter callbacks.
- `AppContext` is presentation/i18n only; it has no persistence identity API.
- `WebhookContext.DB()` and transaction-oriented platform-user methods are
  removed. Use `GetBotUser` and `SetBotUserAccessGranted` instead.
- Telegram construction additionally requires a `ChatInstanceStore`.
- DALgo code moved to optional adapter modules:
  `bots-fw-store-dalgo`, `bots-fw-telegram-dalgo`, and
  `bots-fw-whatsapp-dalgo`.

Release these modules as `v0.x` in dependency order: store contract, DALgo
store adapter, core, platform adapters, then platform DALgo adapters. No module
needs a `v1` release for this migration.
