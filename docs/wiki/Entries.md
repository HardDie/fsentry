# Entries

An entry is one JSON document: file `<id>.json` with the same envelope as folder info (`id`, `name`, `createdAt`, `updatedAt`, `data`). It is not a directory. Related: [folders](Folders), [binaries](Binaries), [on-disk format](On-Disk-Format).

Prerequisites: [Getting started](Getting-Started).

## Create and get

```go
type Settings struct {
	Theme string `json:"theme"`
}

ent, err := db.CreateEntry("settings", Settings{Theme: "dark"})
if err != nil {
	return err
}

got, err := db.GetEntry[Settings]("settings")
if err != nil {
	return err
}
fmt.Println(got.Data.Theme)
```

Create at a nested folder by passing the parent chain:

```go
_, err = db.CreateEntry("Welcome", map[string]string{"body": "hi"}, "My Notes")
```

That file is `<root>/my_notes/welcome.json`.

## Update

Replaces `data` and bumps `updatedAt`. Does not rename.

```go
ent, err = db.UpdateEntry("settings", Settings{Theme: "light"})
```

## Move (rename)

Renames `<old>.json` → `<new>.json` and updates `id` / `name` in the envelope.

```go
ent, err = db.MoveEntry[Settings]("settings", "prefs")
```

## Duplicate

Copies the file to a new name with **fresh** `createdAt` and `updatedAt`. Payload is the same.

```go
copy, err := db.DuplicateEntry[Settings]("prefs", "prefs backup")
// file prefs_backup.json
```

Create of the destination uses `O_EXCL`; an existing ID is [`ErrExist`](Errors).

## Remove

```go
err = db.RemoveEntry("prefs backup")
```

## Type parameters

`CreateEntry` / `UpdateEntry` infer `T`. `GetEntry`, `MoveEntry`, and `DuplicateEntry` usually need `[T]`.

```go
ent, err := db.GetEntry[Settings]("prefs")
raw, err := db.GetEntry[json.RawMessage]("prefs")
```

Untyped `nil`: `db.CreateEntry[any]("empty", nil)` (JSON `null` in `data`).
