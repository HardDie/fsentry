# Paths and IDs

Every folder, entry, and binary has a **display name** (what you pass in) and a filesystem **ID** (what appears on disk). Parent location is a chain of path segments under the store root. Related: [on-disk format](On-Disk-Format), [errors](Errors).

## NameToID

```go
id := fsentry.NameToID("My Notes") // "my_notes"
```

Rules:

1. Lowercase.
2. Spaces become `_`.
3. Drop every rune that is not a Unicode letter, digit, or `_`.
4. Truncate to 200 characters.
5. Empty result, or a Windows reserved device name (`con`, `prn`, `aux`, `nul`, `com0`–`com9`, `lpt0`–`lpt9`), is invalid → [`ErrBadName`](Errors) on create/get/move (the helper itself returns `""`).

`GetFolder("My Game")` and `GetFolder("my_game")` address the same directory.

## Parent path

Methods look like:

```text
CreateFolder[T](name, data, path ...string)
CreateEntry[T](name, data, path ...string)
CreateBinary(name, data, path ...string)
List(path ...string)
```

`path` is the parent folder chain from the root. **Empty `path` is the store root.**

```go
db.CreateFolder("Games", struct{}{})
db.CreateFolder("My Game", meta, "Games")
db.CreateEntry("settings", cfg)                 // <root>/settings.json
db.CreateBinary("icon", bytes, "games", "my_game")
db.List("games", "my_game")
```

Segments must stay inside the root after `filepath.Clean`. Empty segments, `.`, `..`, and separators inside a name are [`ErrBadPath`](Errors). A missing parent folder is also `ErrBadPath` (not `ErrNotExist`).

## IDs in List

[`List`](Folders) returns IDs (`my_game`), not display names (`My Game`). Read `.info.json` / the entry envelope with `GetFolder` / `GetEntry` when you need the original name.
