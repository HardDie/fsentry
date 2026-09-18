# Getting started

This page opens a store and writes one [folder](Folders), one [entry](Entries), and one [binary](Binaries). Related: [paths and IDs](Paths-and-IDs), [errors](Errors), [locking](Locking), [testing](Testing).

## Install

Requires **Go 1.27 or newer**.

```bash
go get github.com/HardDie/fsentry
```

Module path: `github.com/HardDie/fsentry`. Licensed under [GPL-3.0](https://github.com/HardDie/fsentry/blob/master/LICENSE).

## Open a store

`New` only remembers the root path and options. It does **not** create directories or take the lock. Call `Init` before any other method. `Init` is idempotent: a missing root is created; an existing directory is reused.

```go
package main

import (
	"log"

	"github.com/HardDie/fsentry"
)

func main() {
	db := fsentry.New("data", fsentry.WithPretty())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
}
```

| Option | Default | When to use |
|---|---|---|
| `WithPretty()` | compact JSON | Human-readable `.info.json` and entry files (tab indent) |
| `WithLogger(log)` | discard | Unexpected sync/close; never logs payload bytes |
| `WithNoLockFile()` | lock **on** | Tests and benchmarks only — see [Testing](Testing) |
| `WithLockFile()` | already the default | Turn locking back on after another option |
| `WithLockTimeout(d)` | 10 minutes | Steal a stale lock; `d <= 0` keeps the default — see [Locking](Locking) |

Empty root: `New("")` succeeds; `Init` returns [`ErrBadPath`](Errors). If the path exists and is a file, `Init` returns `ErrNotDirectory`.

## First objects

Names are what users type. On disk they become IDs (`"My Notes"` → `my_notes`). Parent folders are extra `path` arguments (IDs or names that map to the same IDs). Empty `path` is the store root.

```go
package main

import (
	"fmt"
	"log"

	"github.com/HardDie/fsentry"
)

type Notebook struct {
	Color string `json:"color"`
}

type Note struct {
	Body string `json:"body"`
}

func main() {
	db := fsentry.New("data", fsentry.WithPretty())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	nb, err := db.CreateFolder("My Notes", Notebook{Color: "blue"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(nb.ID, nb.Name) // my_notes  My Notes

	note, err := db.CreateEntry("Welcome", Note{Body: "hello"}, "My Notes")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(note.ID) // welcome

	if err := db.CreateBinary("cover", []byte{0x89, 0x50, 0x4e, 0x47}, nb.ID); err != nil {
		log.Fatal(err)
	}

	list, err := db.List("my_notes")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list.Entries, list.Binaries) // [welcome] [cover]
}
```

`CreateFolder` / `CreateEntry` infer `T` from the value. `GetFolder` / `GetEntry` usually need an explicit type:

```go
got, err := db.GetFolder[Notebook]("My Notes")
ent, err := db.GetEntry[Note]("Welcome", "my_notes")
```

Untyped `nil` as payload is JSON `null` and needs an explicit type parameter: `db.CreateFolder[any]("Empty", nil)`.

## List the root

```go
root, err := db.List()
if err != nil {
	log.Fatal(err)
}
fmt.Println(root.Folders) // [my_notes]
```

`List` returns **IDs**, not display names. Folders without a readable `.info.json` appear in `CorruptedFolder`. See [On-disk format](On-Disk-Format).

## Tear down

`Drop` unlocks, closes the lock file, and deletes the entire root. Missing root is not an error.

```go
if err := db.Drop(); err != nil {
	log.Fatal(err)
}
```

Do not call `Drop` on a production data directory unless you intend to wipe it.

## Next

- [Folders](Folders) — update, move, nested folders, remove
- [Entries](Entries) — duplicate and typed payloads
- [Binaries](Binaries) — `GetBinary` with a sized buffer
- [Errors](Errors) — `errors.Is(err, fsentry.ErrExist)`
- [Export and import](Export-and-Import) — zip backup of a store or folder
- [Validate](Validate) — report on-disk corruption
