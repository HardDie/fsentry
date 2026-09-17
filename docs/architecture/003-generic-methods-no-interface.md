# 3. Generic methods on `*DB`; no `IFSEntry`

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

Old fsentry (and DeckBuilder) used `data any` plus `json.Unmarshal` at the call site, and an `IFSEntry` interface for the store. Go 1.27 allows type parameters on methods. Interface methods still cannot have their own type parameters, and a generic method cannot satisfy an interface method.

## Considered options

1. **Keep `data any` and `IFSEntry`** for drop-in DeckBuilder source compatibility.
2. **Package-level functions** `CreateEntry[T](db *DB, …)` so an interface of non-generic methods could exist.
3. **Generic methods on the concrete `*DB`** (`CreateEntry[T]`, `GetEntry[T]`, …). No store interface that pretends to include those methods. Folder payloads stay `any` until someone asks to genericize them.

## Decision

Use option 3. `go.mod` is `go 1.27`. CI and local toolchain are the latest 1.27.x patch. Do not compile with 1.26 or older.

Callers take `*DB`. Tests mock with a real temp-dir store. Escape hatch: `GetEntry[json.RawMessage]`. Return `Entry[T]`, `FolderInfo`, and `List` by value.

## Consequences

### Positive

* Call sites become `db.GetEntry[Settings]("settings")` instead of unmarshal after get.
* No fake interface that cannot actually describe the API.

### Negative and risks

* DeckBuilder must change types (`*DB` instead of `IFSEntry`). On-disk format does not change.
* Folder `data any` and entry `T` are inconsistent until a later decision.

### Neutral

* `WithPretty` and other options stay functional options on `New`, not generics.
