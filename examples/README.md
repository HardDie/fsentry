# Examples

Each subdirectory is a small `package main` you can run from the module root:

```bash
go run ./examples/basic
make examples   # compile all of them
```

They use a temporary directory (printed on stdout) so they do not write into this repo. Inspect that path if you want to see `.info.json`, `*.json`, and `*.bin` on disk.

Requires **Go 1.27+**.

## Order

Start at the top. Later folders assume you already know `New` + `Init`.

| Folder | What it shows |
|---|---|
| [basic](basic/) | Open a store, create one folder, one JSON entry, one binary, list the root |
| [nested](nested/) | Parent `path` chains (`CreateFolder` / `CreateEntry` under another folder) |
| [typed](typed/) | Struct payloads, `GetFolder[T]` / `GetEntry[T]`, `json.RawMessage`, untyped `nil` |
| [errors](errors/) | `errors.Is` with `ErrBadName`, `ErrExist`, `ErrNotExist`, `ErrBadPath` |
| [options](options/) | `WithPretty`, `WithNoLockFile` vs `WithLockFile`, `WithLockTimeout` |
| [mutate](mutate/) | Update, move, `DuplicateEntry`, remove |
| [binaries](binaries/) | `.bin` create / get (sized buffer) / update / move / remove |
| [export-import](export-import/) | Zip the store or one folder, import into another root |
| [validate](validate/) | `Validate` on a healthy tree, then after a broken folder |
| [notebook](notebook/) | Nested “app” tree: notebooks → notes → cover image, then export + validate |

## Notes

- Production code omits `WithNoLockFile()`. These programs pass it so they stay easy to run next to tests.
- Display names become IDs (`"My Notes"` → `my_notes`). Get/move accept either form.
- Godoc also has `Example*` tests in the library package (`go test -run Example`).
