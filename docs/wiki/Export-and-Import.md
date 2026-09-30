# Export and import

Backup or copy a store (or one [folder](Folders)) as a zip archive. Zip names always use `/`. The lock file is not included. Related: [on-disk format](On-Disk-Format), [errors](Errors), [validate](Validate).

Prerequisites: [Getting started](Getting-Started) (`New` + `Init`).

## Export the whole store

```go
var buf bytes.Buffer
if err := db.Export(&buf); err != nil {
	return err
}
// buf.Bytes() is a zip file
```

Write to a file:

```go
f, err := os.Create("backup.zip")
if err != nil {
	return err
}
defer f.Close()
if err := db.Export(f); err != nil {
	return err
}
```

Root export puts files at the zip root (`settings.json`, `my_notes/.info.json`, …). `.fsentry.lock` is omitted.

## Export one folder

Non-root `path` wraps entries in that folder's ID (same idea as DeckBuilder game export):

```go
var buf bytes.Buffer
if err := db.Export(&buf, "My Notes"); err != nil {
	return err
}
// zip contains my_notes/.info.json, my_notes/welcome.json, …
```

## Export one folder and its children

`ExportFolder` takes the folder name and the parent chain. DeckBuilder games live under `games`, so one game is:

```go
var buf bytes.Buffer
if err := db.ExportFolder(&buf, "My Game", "games"); err != nil {
	return err
}
// zip contains my_game/.info.json, my_game/rules.json, …
// sibling games and the rest of the store are not included
```

A folder at the store root omits the parent path: `db.ExportFolder(&buf, "My Notes")`.

## Import one folder into an existing store

`ImportFolder` extracts one folder archive under `path`. Other children of that directory stay. The archive's single top-level folder must already be an id (`my_game`, not `My Game`). A whole-store zip is `ErrBadArchive`.

`name` is the display name to assign. Empty keeps the archive id and its `.info.json`. A non-empty name rewrites the zip root to that id before anything is written, and updates `.info.json` without changing timestamps. Importing archive `my_game` as `"Other Title"` leaves an existing `my_game` folder in place.

If the destination id already exists, that folder is replaced. A failed import puts the previous folder back.

```go
id, err := db.ImportFolder(bytes.NewReader(buf.Bytes()), "", "games")
if err != nil {
	return err
}
// id == "my_game", files at games/my_game/…

id, err = db.ImportFolder(bytes.NewReader(buf.Bytes()), "Other Title", "games")
// id == "other_title"
```

## Import

```go
if err := db.Import(bytes.NewReader(buf.Bytes())); err != nil {
	return err
}
```

Import a folder zip under a parent:

```go
if err := db.Import(bytes.NewReader(buf.Bytes()), "Inbox"); err != nil {
	return err
}
// Inbox/my_notes/…
```

Existing files are not overwritten (`ErrExist`). Names that would leave the destination (`../secret`) are `ErrBadArchive`. A failed import removes files and directories **this call** created. After a successful import, [Validate](Validate) the destination if the zip may have been edited.

If the reader also implements `io.ReaderAt` and `Size() int64` (for example `*bytes.Reader`), the zip is opened without copying the whole archive into a second buffer.

## Errors

| Sentinel | When |
|---|---|
| `ErrBadArchive` | Not a zip, a path would extract outside the destination, or `ImportFolder` got an archive that is not a single folder |
| `ErrExist` | A file in the zip already exists at the destination |
| `ErrIsDirectory` | Zip file would replace an existing directory |
| `ErrBadPath` / `ErrNotExist` | Destination folder missing. `ExportFolder` returns `ErrNotExist` when the named folder is missing |
