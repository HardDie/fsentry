# 3. Generic methods on `*DB`; no `IFSEntry`

* **Status:** Accepted
* **Date:** 2026-09-17
* **Updated:** 2026-09-18
* **Authors:** @oleg

---

## Context

Old fsentry (and DeckBuilder) used `data any` plus `json.Unmarshal` at the call site, and an `IFSEntry` interface for the store. Go 1.27 allows type parameters on methods. Interface methods still cannot have their own type parameters, and a generic method cannot satisfy an interface method.

Folder payloads were left as `any` in the first cut so DeckBuilder could keep `json.Unmarshal`. That split is no longer wanted: folders should match entries.

## Considered options

1. **Keep `data any` and `IFSEntry`** for drop-in DeckBuilder source compatibility.
2. **Package-level functions** `CreateEntry[T](db *DB, …)` so an interface of non-generic methods could exist.
3. **Generic methods on the concrete `*DB`** for both entries and folders (`CreateEntry[T]`, `GetEntry[T]`, `CreateFolder[T]`, `GetFolder[T]`, …). No store interface that pretends to include those methods.

## Decision

Use option 3. `go.mod` is `go 1.27`. CI and local toolchain are the latest 1.27.x patch. Do not compile with 1.26 or older.

Callers take `*DB`. Tests mock with a real temp-dir store. Escape hatch: `GetEntry[json.RawMessage]` / `GetFolder[json.RawMessage]`. Return `Entry[T]`, `FolderInfo[T]`, and `List` by value. Untyped `nil` as payload needs an explicit type (`CreateFolder[any]("x", nil)` → JSON `null`).

`RemoveFolder` and `List` stay non-generic (no payload).

## Consequences

### Positive

* Call sites become `db.GetEntry[Settings]("settings")` and `db.GetFolder[Game]("my_game")` instead of unmarshal after get.
* No fake interface that cannot actually describe the API.
* Folder and entry payloads use the same pattern.

### Negative and risks

* DeckBuilder must change types (`*DB` instead of `IFSEntry`, typed get). On-disk format does not change.
* `CreateFolder("x", nil)` does not compile.

### Neutral

* `WithPretty` and other options stay functional options on `New`, not generics.
