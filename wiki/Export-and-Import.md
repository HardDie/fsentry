# Export and import

Backup or copy a store (or one [folder](Folders)) as a zip archive. Zip names always use `/`. The lock file is not included. Related: [on-disk format](On-Disk-Format), [errors](Errors).

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

Existing files are not overwritten (`ErrExist`). Names that would leave the destination (`../secret`) are `ErrBadArchive`. A failed import removes files and directories **this call** created.

If the reader also implements `io.ReaderAt` and `Size() int64` (for example `*bytes.Reader`), the zip is opened without copying the whole archive into a second buffer.

## Errors

| Sentinel | When |
|---|---|
| `ErrBadArchive` | Not a zip, or a path would extract outside the destination |
| `ErrExist` | A file in the zip already exists at the destination |
| `ErrIsDirectory` | Zip file would replace an existing directory |
| `ErrBadPath` / `ErrNotExist` | Destination folder missing |
