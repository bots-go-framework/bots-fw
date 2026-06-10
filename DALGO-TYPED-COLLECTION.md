# dalgo typed `Collection[K, T]` — applicability to bots-fw

**Status:** **adopted for point CRUD.** dalgo `v0.61.0` added the read accessor
`dal.GetRecordWithIDIntoData` and `v0.62.0` the write twin
`dal.InsertRecordWithDataAndID`; `GetBotChat`, `GetPlatformUser`, and
`CreatePlatformUserRecord` are all migrated onto them. Struct/composite ids
remain future work.
**Date:** 2026-06-10

> **Update (v0.61.0):** the dalgo-side unlock landed as
> `dal.GetRecordWithIDIntoData[K, D](ctx, s, key, id, data)` — it decodes into a
> caller-supplied value, so interface `D` (e.g. `BotChatData`) works. The reads
> below are migrated; the design analysis is retained for context and for the
> still-open items.

## TL;DR

dalgo `v0.59+` ships a typed convenience layer — `dal.Collection[K, T]` (id type
`K`, record type `T`) plus the `record.GetWithID` helper — that turns point CRUD
into one-liners returning typed values. We evaluated adopting it in `botsdal`.

**We did not adopt it.** The blocker is not the version (we are now on `v0.59.1`);
it is a **design mismatch**: `botsdal`'s data-access functions are deliberately
**generic over interface-typed model data supplied by a factory**, while
`Collection[K, T]` is built for **concrete, allocatable `T`** and returns a
different shape than the `record.DataWithID[K, D]` the framework standardises on.

This document records why, and what would have to change — on the **bots-fw side**
and/or the **dalgo side** — to make adoption worthwhile. It is feasible
eventually; it needs cleanup/refactoring first.

## What the new layer offers

```go
// dalgo v0.59+
type Collection[K comparable, T any] interface {
    GetData(ctx, s ReadSession, id K) (T, error)
    GetRecord(ctx, s ReadSession, id K) (Record, error)
    All/Count/Exists/First ...
    Insert / InsertWithID / InsertRecord
    SetByID / SetRecord
    UpdateByID / UpdateByKey
    DeleteByID / DeleteByKey
    In(parent *Key) Collection[K, T]
}
func CollectionOf[K comparable, T CollectionNamer](opts ...CollectionOption) Collection[K, T]
func CollectionAt[K comparable, T any](name string, opts ...CollectionOption) Collection[K, T]

// package record (kept out of dal to avoid a dal -> record import cycle)
func GetWithID[K comparable, T any](ctx, c dal.Collection[K, T], s dal.ReadSession, id K) (record.WithID[K], error)
```

Internally each read does `new(T)` / `NewRecordWithData(key, &value)` and lets the
adapter decode into it; writes wrap `&value`.

## Why it does not fit `botsdal` today

### 1. Model data is an interface + factory, not a concrete `T`

The framework is generic over app-supplied models:

```go
// bots-fw-store/botsfwmodels
type BotChatData interface { ... }
type PlatformUserData interface { ... }
```

`botsdal` reads/writes through those interfaces, with the concrete type provided
by a factory at the call site:

```go
func GetBotChat(ctx, tx, platformID, botID, chatID string,
    newData func() botsfwmodels.BotChatData,
) (record.DataWithID[string, botsfwmodels.BotChatData], error) {
    key := NewBotChatKey(platformID, botID, chatID)
    data := newData()                          // concrete instance behind the interface
    chat := record.NewDataWithID(chatID, key, data)
    return chat, tx.Get(ctx, chat.Record)
}
```

`Collection[K, T].GetData` does `new(T)` internally. If `T` is an **interface**
(`BotChatData`), `new(T)` is a `*BotChatData` — a pointer to an interface — which
adapters cannot decode into. The factory pattern exists precisely because the
framework never holds a concrete type to allocate.

### 2. The return shape is `DataWithID[K, D]`, not `T` / `WithID[K]`

`botsdal` standardises on `record.DataWithID[string, D]` — it bundles the typed
**id**, the typed **data** (`.Data`), and the **`dal.Record`** (`.Record`).
Callers depend on all three downstream, e.g.:

```go
whcb.platformUser, _ = botsdal.GetPlatformUser(...)   // record.DataWithID[string, PlatformUserData]
... tx.Set(ctx, whcb.platformUser.Record)             // .Record reused for writes
... whcb.platformUser.ID, whcb.platformUser.Data      // .ID and typed .Data reused
```

The new layer returns either a plain `T` (`GetData`) or `record.WithID[K]`
(`GetWithID`). `WithID[K]` carries `ID`/`Key`/`Record` but **no typed `.Data`** —
strictly less than the `DataWithID[string, D]` the framework already builds by
hand. So even ignoring (1), adopting it would be a downgrade in the return type.

### 3. (Resolved) version

Historically `go.mod` named `v0.41.15`, but module resolution already selected
`v0.58.2` transitively. We bumped the direct requirement to **`v0.59.1`**; build,
`go vet`, and all tests stay green. So version is no longer a blocker — the
design mismatch above is.

## What would make adoption worthwhile

Two independent tracks; both probably needed.

### A. dalgo side — make the typed layer compose with `DataWithID`

The richest shape (`DataWithID[K, D]` = typed id **and** typed data **and** the
record) is what real apps want. The convenience layer should offer it, and should
support **caller-supplied data** so interface `D` + factory works.

Because `record.DataWithID` lives in package `record` (which imports `dal`), these
**must be free functions in `record`**, mirroring `GetWithID` — a method on
`dal.Collection` returning `record.DataWithID` would create a `dal -> record`
import cycle.

Proposed additions (package `record`):

```go
// Concrete D: decodes into new(D); completes the GetData/GetRecord/GetWithID family.
func GetDataWithID[K comparable, D any](
    ctx context.Context, c dal.Collection[K, D], s dal.ReadSession, id K,
) (DataWithID[K, D], error)

// Interface/factory D: caller supplies the instance to decode into — this is the
// one that unblocks bots-fw, because new(T) cannot allocate an interface.
func ReadDataWithID[K comparable, D any](
    ctx context.Context, s dal.ReadSession, key *dal.Key, id K, data D,
) (DataWithID[K, D], error)
```

`ReadDataWithID` is essentially what `GetBotChat`/`GetPlatformUser` already do by
hand, upstreamed and reusable. (It does not even need `Collection` — it is a
record-level helper over a `*dal.Key` + a caller value.)

Open design question for dalgo (raised separately): **should `Collection[K, T]`
itself gain a data-target/factory variant** (e.g. `GetRecordInto(ctx, s, id, data
any)`), or should the typed-data accessors stay entirely in package `record` as
free functions? Recommendation: keep them in `record` as free functions — it
keeps the `dal.Collection` interface small and respects the cycle constraint.

### B. bots-fw side — cleanup to expose the fit

Even with (A), `botsdal` would benefit from:

- Deciding whether the public DAL functions become **generic over the concrete
  data type** (`GetBotChat[D botsfwmodels.BotChatData]`) so `K, D` are inferable,
  or stay interface+factory and use `ReadDataWithID` from (A).
- Consolidating the per-entity key builders (`NewPlatformKey` / `NewBotKey` /
  `NewBotChatKey` / `NewPlatformUserKey`) with `Collection.In(parent)` nesting if
  the functions move onto `Collection` handles.
- Removing the large commented-out blocks in `app_user_store.go`,
  `dal_bot_user.go`, `facade_user.go` first, so the refactor target is clear.

## Recommendation / status

1. ~~**Next (dalgo):** add a caller-supplied/interface-data accessor.~~ **Done in
   dalgo `v0.61.0`** as `dal.GetRecordWithIDIntoData` (the type
   `record.ReadDataWithID` was renamed during review to `GetRecordWithIDIntoData`,
   and `record.DataWithID` is now an alias for `dal.RecordWithDataAndID`).
2. ~~**Later (bots-fw):** migrate `GetBotChat` / `GetPlatformUser`.~~ **Done** —
   both now delegate to `dal.GetRecordWithIDIntoData(ctx, tx, key, id, data)`,
   returning the same `record.DataWithID[string, D]` shape as before (it is now an
   alias of `dal.RecordWithDataAndID`).
3. ~~`CreatePlatformUserRecord` write migration.~~ **Done** — dalgo `v0.62.0`
   added the write twin `dal.InsertRecordWithDataAndID`, and
   `CreatePlatformUserRecord` now delegates to it.
4. **Still open:**
   - struct/composite ids in the typed `id K` slot remain deferred in dalgo.
   - the large commented-out blocks in `app_user_store.go` / `dal_bot_user.go` /
     `facade_user.go` are still candidates for cleanup.
