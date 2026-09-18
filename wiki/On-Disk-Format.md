# On-disk format

This is the public layout. Existing trees that follow it (for example DeckBuilder data directories) should keep working. Related: [paths and IDs](Paths-and-IDs), [locking](Locking).

## Layout

```text
<root>/                          # Init()
  .fsentry.lock                  # advisory lock; hidden from List
  settings.json                  # Entry
  my_notes/                      # Folder
    .info.json
    welcome.json                 # Entry
    cover.bin                    # Binary
    recipes/                     # Nested folder
      .info.json
```

| Object | Files |
|---|---|
| [Folder](Folders) | directory `<id>/` + `<id>/.info.json` |
| [Entry](Entries) | `<id>.json` |
| [Binary](Binaries) | `<id>.bin` (raw bytes, no sidecar) |

Hidden names (leading `.`) are not user objects except reserved `.info.json`. `List` also skips `.fsentry.lock`. Directories `0755`, files `0666` masked by umask.

## JSON envelope

Folder `.info.json` and entry files share this shape. `WithPretty()` indents with a tab; default is compact JSON plus a trailing newline.

```json
{
	"id": "my_notes",
	"name": "\"My Notes\"",
	"createdAt": "2026-01-02T03:04:05Z",
	"updatedAt": "2026-01-02T03:04:05Z",
	"data": { "color": "blue" }
}
```

- `id` — filesystem ID.
- `name` — `QuotedString`: a JSON string whose contents are a quoted Go string (double-encoded). That matches existing DeckBuilder files. In Go you still see `"My Notes"` on `FolderInfo.Name` / `Entry.Name`.
- `createdAt` / `updatedAt` — UTC RFC3339. On create they are the same instant.
- `data` — any JSON (`object`, `array`, `null`).

Writes use create-exclusive (`O_CREATE|O_EXCL`) or replace via temp file + sync + rename so a crash does not leave a half JSON file. Create folder: `Mkdir` then `.info.json`; if info fails, the empty directory is removed.

## Corrupted folders

A directory that looks like a folder but has missing or unreadable `.info.json` is listed in `List.CorruptedFolder`. `GetFolder` / `RemoveFolder` return [`ErrFolderCorrupted`](Errors) rather than treating it as a normal folder.

## Compatibility

Do not invent a parallel layout (different sidecar names, SQLite next to the tree, and so on) if you need to read existing stores. Changing this contract is an architecture decision, not a silent refactor.
