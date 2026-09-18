# fsentry

[![Go Reference](https://pkg.go.dev/badge/github.com/HardDie/fsentry.svg)](https://pkg.go.dev/github.com/HardDie/fsentry)

A Go library that treats a **directory tree as a small database**.

You store nested **folders**, JSON **entries**, and opaque **binaries** on disk under one root. There is no SQL server, no daemon, and no application domain types. A zip of the root is a backup.

Requires **Go 1.27+** (generic methods on `*DB`).

## Install

```bash
go get github.com/HardDie/fsentry
```

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/HardDie/fsentry"
)

type Meta struct {
	Kind string `json:"kind"`
}

func main() {
	db := fsentry.New("data", fsentry.WithPretty())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	folder, err := db.CreateFolder("My Notes", Meta{Kind: "notebook"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(folder.ID) // my_notes

	entry, err := db.CreateEntry("Welcome", map[string]string{"body": "hello"}, folder.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(entry.Name) // Welcome

	if err := db.CreateBinary("cover", []byte("PNG..."), folder.ID); err != nil {
		log.Fatal(err)
	}
}
```

`New` only holds the path. Call `Init` before other methods (it creates the root if needed). Production stores take an advisory lock on `.fsentry.lock`; tests should pass `fsentry.WithNoLockFile()`.

## Wiki

Guides and copy-paste examples live on the [GitHub wiki](https://github.com/HardDie/fsentry/wiki). Source for those pages is in [`wiki/`](wiki/) in this repo.

| Page | Topic |
|---|---|
| [Home](https://github.com/HardDie/fsentry/wiki) | Map of the wiki |
| [Getting started](https://github.com/HardDie/fsentry/wiki/Getting-Started) | Install, `New` / `Init` / `Drop`, first store |
| [Folders](https://github.com/HardDie/fsentry/wiki/Folders) | Create, get, list, update, move, remove |
| [Entries](https://github.com/HardDie/fsentry/wiki/Entries) | JSON documents and `DuplicateEntry` |
| [Binaries](https://github.com/HardDie/fsentry/wiki/Binaries) | `.bin` files and `GetBinary` buffers |
| [Paths and IDs](https://github.com/HardDie/fsentry/wiki/Paths-and-IDs) | `NameToID`, parent `path` chains |
| [On-disk format](https://github.com/HardDie/fsentry/wiki/On-Disk-Format) | Directories, envelopes, lock file |
| [Errors](https://github.com/HardDie/fsentry/wiki/Errors) | `errors.Is` sentinels |
| [Locking](https://github.com/HardDie/fsentry/wiki/Locking) | Mutex + lock file, steal after timeout |
| [Testing](https://github.com/HardDie/fsentry/wiki/Testing) | Temp dirs, race detector, benches |
| [Export and import](https://github.com/HardDie/fsentry/wiki/Export-and-Import) | Zip backup of a store or folder |

## Object kinds

| Kind | On disk | Role |
|---|---|---|
| Folder | directory named by ID + `.info.json` | Nesting node; optional JSON payload |
| Entry | `<id>.json` | One JSON document in an envelope |
| Binary | `<id>.bin` | Raw bytes (images, audio, anything) |

Display names become portable IDs (`"My Notes"` → `my_notes`). `GetFolder("My Notes")` and `GetFolder("my_notes")` hit the same folder. See [Paths and IDs](https://github.com/HardDie/fsentry/wiki/Paths-and-IDs).

## Development

```bash
make help
make test
make test-integration
make bench
```

Architecture notes: [`docs/architecture`](docs/architecture/INDEX.md). Implementation spec for contributors: [`CURSOR.md`](CURSOR.md).

## License

See the repository license file when present. Module path: [`github.com/HardDie/fsentry`](https://github.com/HardDie/fsentry).
