# fsentry wiki

**fsentry** is a Go library that treats a directory tree as a small database. Callers store [folders](Folders), JSON [entries](Entries), and opaque [binaries](Binaries) under one root. There is no SQL, no network server, and no application domain types in this module.

Start with the [README](https://github.com/HardDie/fsentry#readme) for install, then this wiki for worked examples.

| Page | What you will do |
|---|---|
| [Getting started](Getting-Started) | Install, open a store with `New` + `Init`, create a folder, entry, and binary, then `Drop` |
| [Folders](Folders) | Create / get / list / update / move / remove folders and nested payloads |
| [Entries](Entries) | Typed JSON documents, duplicate, update, move |
| [Binaries](Binaries) | Raw `.bin` files and reading into a caller buffer |
| [Paths and IDs](Paths-and-IDs) | `NameToID`, parent path chains, what is a bad name |
| [On-disk format](On-Disk-Format) | `.info.json`, envelopes, `QuotedString`, lock file |
| [Errors](Errors) | Sentinel errors and `errors.Is` |
| [Locking](Locking) | In-process mutex, `.fsentry.lock`, steal after timeout |
| [Testing](Testing) | `t.TempDir()`, `WithNoLockFile()`, race tests and benches |
| [Export and import](Export-and-Import) | Zip backup of a store or folder |

## Mental model

Think nested documents on disk, not a thin POSIX wrapper.

```text
<root>/                    # Init()
  settings.json            # Entry at the root
  my_notes/                # Folder (ID from "My Notes")
    .info.json
    welcome.json           # Entry
    cover.bin              # Binary
    recipes/               # Nested folder
      .info.json
```

Production: lock file **on** (default). Tests: [Testing](Testing). Concurrency details: [Locking](Locking).

## Requirements

- **Go 1.27+** — folder and entry methods are generic on `*DB` (`CreateEntry[T]`, `GetFolder[T]`, …).
- Import only `github.com/HardDie/fsentry`. Do not depend on `internal/`.

## Publishing these pages

The markdown in the `wiki/` directory of the source repository is the source for this GitHub wiki (`https://github.com/HardDie/fsentry/wiki`). Copy `*.md` (including `_Sidebar.md` and `_Footer.md`) into the wiki git remote:

```bash
git clone https://github.com/HardDie/fsentry.wiki.git
cp /path/to/fsentry/wiki/*.md fsentry.wiki/
cd fsentry.wiki
git add .
git commit -m "Sync wiki from repo wiki/"
git push
```

GitHub wiki links in these files are **page slugs without `.md`**, for example `[Folders](Folders)` → `https://github.com/HardDie/fsentry/wiki/Folders`.
