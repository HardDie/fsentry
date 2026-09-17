# fsentry Development Guidelines

fsentry is a **Go library** that treats a directory tree as a small database. Callers store hierarchical data as folders and files on disk. There is no SQL, no network server, and no DeckBuilder/TTS domain logic here.

This repo is a **clean rewrite** of [HardDie/fsentry](https://github.com/HardDie/fsentry). That library is unfinished (duplicate package trees, `List` is a stub, `package fsentry` lives in `main.go`). Do **not** copy that layout. Take the **on-disk format and public operations** that [HardDie/DeckBuilder](https://github.com/HardDie/DeckBuilder) already depends on, then implement them once, with tests and godoc.

**Reference implementations** (local clones):

- `/Users/oleg/Project/go/fsentry` — previous library (behavior and tests, not package layout)
- `/Users/oleg/Project/go/DeckBuilder` — first real consumer (`internal/db/*` talks only to `IFSEntry`)

Prefer matching DeckBuilder’s usage over inventing a second API.

---

## Product scope (this library)

**Must have (first vertical slice)**

- Open a store at a root directory (`New` + `Init`). Missing root is created; existing root is reused.
- Three object kinds at any folder path: **folder**, **entry** (`.json`), **binary** (`.bin`).
- Create / get / list / update / move / remove for all three. Folders also **duplicate** (deep copy).
- Path is a chain of folder IDs under the root (`"games", gameID, collectionID`).
- Stable on-disk layout so an existing DeckBuilder data directory keeps working.
- Sentinel errors that callers can `errors.Is` (`ErrExist`, `ErrNotExist`, `ErrBadName`, `ErrBadPath`, …).
- Inter-process lock file on the store (off in unit/integration tests unless the test is about locking).
- Benchmarks with `testing.B` / `ReportAllocs`; hot paths aim for **zero extra allocations**.
- Entry payloads are **generic methods on `*DB`** (`CreateEntry[T]`, `GetEntry[T]`, …), not `any`. Requires **Go 1.27+**.

**Port from old fsentry / DeckBuilder (same slice, not a follow-up)**

- Name → ID sanitization (`NameToID`) so file names are portable across macOS, Linux, and Windows.
- Folder sidecar `.info.json` (original name, timestamps, optional JSON payload).
- Entry JSON envelope (`id`, `name`, `createdAt`, `updatedAt`, `data`).
- `QuotedString` encoding for the `name` field on disk (DeckBuilder data already uses it).
- In-process mutex **and** an advisory lock file so two processes do not race the same tree.
- `Drop` deletes the whole root (tests and DeckBuilder `core.Drop`).
- `UpdateFolderNameWithoutTimestamp` (DeckBuilder import/rename path).

**Out of scope unless explicitly requested**

- Games, collections, decks, cards, images-as-domain, HTTP, GUI.
- SQL, SQLite, or any other backend. The filesystem **is** the database.
- Replication, a daemon, or a distributed lock service.
- Encryption, compression, or content-addressed blobs.
- Changing the on-disk format in a way that cannot read existing DeckBuilder trees.

---

## Mental model

Think **filesystem as nested documents**, not a POSIX wrapper.

| Object | On disk | Role |
|---|---|---|
| **Folder** | directory named by **ID** + file `.info.json` inside it | Hierarchy node. May contain folders, entries, binaries. Payload lives in `.info.json` `data`. |
| **Entry** | file `<id>.json` | One JSON document (Go struct in `data`). |
| **Binary** | file `<id>.bin` | Opaque bytes (images, audio, anything). No metadata file. |

A folder without a readable `.info.json` is **corrupted** (list it separately; `GetFolder` fails). Hidden names (leading `.`) are not user objects except the reserved `.info.json`.

**DeckBuilder layout** (illustrates nesting; this library does not know these names):

```text
<root>/                          # Init()
  settings.json                  # Entry
  games/                         # Folder
    .info.json
    my_game/                     # Folder (game)
      .info.json                 # description, image URL, …
      image.bin                  # Binary
      my_collection/             # Folder
        .info.json
        image.bin
        my_deck/                 # Folder
          .info.json
          image.bin
          cards/                 # Folder; payload is a JSON map of cards
            .info.json
            1.bin                # card image, name = card id
```

---

## On-disk contract

This is the library’s public format. Agents must not invent a parallel layout.

**Root.** Caller passes an absolute or relative directory. All operations stay under that root after `filepath.Clean`. Reject empty path segments, `.`, `..`, separators inside a name, and any resolved path that escapes the root (`ErrBadPath`).

**ID from name.** `NameToID(name)`:

1. Lowercase.
2. Spaces → `_`.
3. Drop every rune that is not a Unicode letter, digit, or `_`.
4. Truncate to 200 characters.
5. Empty result, or a Windows reserved device name (`con`, `prn`, `aux`, `nul`, `com0`–`com9`, `lpt0`–`lpt9`), is `ErrBadName`.

Create/get/move take the **user name** (or an existing ID that already is a valid ID). The directory or file name on disk is always the ID. `GetFolder("My Game")` and `GetFolder("my_game")` address the same folder.

**Folder `.info.json`** (pretty-print when `WithPretty()`):

```json
{
  "id": "my_game",
  "name": "\"My Game\"",
  "createdAt": "2026-01-02T03:04:05Z",
  "updatedAt": "2026-01-02T03:04:05Z",
  "data": { }
}
```

`name` is a `QuotedString`: JSON string whose contents are a quoted Go string (double-encoded). That matches existing files. `data` is any JSON (object, array, `null`). Timestamps are UTC RFC3339. On create, `createdAt` and `updatedAt` are the same instant (not a null `updatedAt`).

**Entry `<id>.json`:** same envelope as folder info (`id`, `name`, `createdAt`, `updatedAt`, `data`). The file name is `id + ".json"`.

**Binary `<id>.bin`:** raw bytes, no envelope.

**List** of a path returns IDs (not display names):

- `Folders` — child directories that look like folders
- `Entries` — `*.json` files except `.info.json`, IDs without suffix
- `Binaries` — `*.bin` files, IDs without suffix
- `CorruptedFolder` — directories that exist but `.info.json` is missing or unreadable

Skip other files. Old fsentry omitted binaries from `List`; this rewrite includes them (additive; DeckBuilder does not need the field). Skip the lock file (`.fsentry.lock`).

**Writes.** Create uses `O_CREATE|O_EXCL` (exist → `ErrExist`). Update truncates an existing file (missing → `ErrNotExist`). Prefer write-to-temp-in-same-dir + `Sync` + `Rename` for JSON so a crash does not leave a half file. Create folder: `Mkdir` the ID directory, then create `.info.json`; if the info file fails, remove the empty directory. `RemoveFolder` is recursive (`RemoveAll`). Duplicate folder is a deep copy, then rewrite the destination `.info.json` with the new name/id and fresh timestamps.

**Lock file.** Production stores take an advisory exclusive lock on `<root>/.fsentry.lock` for the duration of each public operation (`flock` / `LockFileEx`, not a busy-wait “write pid and hope”). `Init` creates the file. `Drop` unlocks, closes, then removes the root. Hidden; `List` ignores it. This is **on by default**. Unit tests, integration tests, and benchmarks pass `WithNoLockFile()` so they stay fast, isolated, and race-detector-friendly. The exception is tests whose job is locking: those omit the option (or use `WithLockFile()`) and prove two `*DB` handles on the same root cannot enter a write at the same time. Do not spawn extra processes unless a lock test cannot be done with two handles in one process.

**Permissions.** Directories `0755`, files `0666` masked by umask (same idea as old code). Do not chmod through ACLs unless a Windows-only test proves we need `go-acl` again.

---

## Public Go API (this port)

Module: `github.com/HardDie/fsentry`. **One public package:** `fsentry`. Callers import only that. Do not split types into `pkg/fsentry` + constructor in another package (old layout).

```go
db := fsentry.New("path/to/root", fsentry.WithPretty())
if err := db.Init(); err != nil { ... }

ent, err := db.CreateEntry("settings", Settings{Theme: "dark"})
got, err := db.GetEntry[Settings]("settings")
```

**Options**

| Option | Default | Role |
|---|---|---|
| `WithPretty()` | compact JSON | indent with a tab |
| `WithLogger(Logger)` | discard | unexpected sync/close |
| `WithNoLockFile()` | lock **on** | tests and benchmarks only |

No global state. Production callers do not pass `WithNoLockFile()`.

**Concrete type `*DB`.** Go **1.27** allows type parameters on methods. Entry operations are **methods on `*DB`**, not package-level functions. Do not keep a parallel `data any` entry API.

**No interface for generic methods.** Go 1.27 still forbids type parameters on interface methods, and a generic method cannot satisfy an interface. Do not export `IFSEntry` (or any interface) that pretends to include `CreateEntry[T]`. Callers take `*DB`. Mock with a real temp-dir store in tests.

```text
(*DB) Init() error
(*DB) Drop() error
(*DB) List(path ...string) (List, error)

(*DB) CreateFolder(name string, data any, path ...string) (FolderInfo, error)
(*DB) GetFolder(name string, path ...string) (FolderInfo, error)
(*DB) MoveFolder(oldName, newName string, path ...string) (FolderInfo, error)
(*DB) UpdateFolder(name string, data any, path ...string) (FolderInfo, error)
(*DB) RemoveFolder(name string, path ...string) error
(*DB) DuplicateFolder(srcName, dstName string, path ...string) (FolderInfo, error)
(*DB) UpdateFolderNameWithoutTimestamp(oldName, newName string, path ...string) (FolderInfo, error)

(*DB) CreateEntry[T any](name string, data T, path ...string) (Entry[T], error)
(*DB) GetEntry[T any](name string, path ...string) (Entry[T], error)
(*DB) MoveEntry[T any](oldName, newName string, path ...string) (Entry[T], error)
(*DB) UpdateEntry[T any](name string, data T, path ...string) (Entry[T], error)
(*DB) DuplicateEntry[T any](srcName, dstName string, path ...string) (Entry[T], error)
(*DB) RemoveEntry(name string, path ...string) error

(*DB) CreateBinary(name string, data []byte, path ...string) error
(*DB) GetBinary(name string, buf []byte, path ...string) ([]byte, error)
(*DB) MoveBinary(oldName, newName string, path ...string) error
(*DB) UpdateBinary(name string, data []byte, path ...string) error
(*DB) RemoveBinary(name string, path ...string) error
```

```go
type Entry[T any] struct {
    ID, Name           string
    CreatedAt, UpdatedAt time.Time
    Data               T
}
```

Inference: `db.CreateEntry("e", myStruct)` does not need `[T]`. `GetEntry` usually needs `[T]`: `db.GetEntry[Settings]("settings")`. Escape hatch: `db.GetEntry[json.RawMessage]("settings")`. Folder `data` stays `any` for now (DeckBuilder still `json.Unmarshal`s folder payloads); do not genericize folders unless asked.

Language: `go 1.27` in `go.mod` (generic methods). CI and local toolchain: latest **1.27.x** patch (currently 1.27.1). Do not set `go 1.22` and wait for a newer compiler.

**Return values.** Return structs **by value** (`Entry[T]`, `FolderInfo`, `List`), not pointers, unless an API must distinguish missing from zero (errors do that). `GetBinary` reads into caller `buf` and returns that slice (or a grown one); passing a sized buffer is the zero-alloc path. `buf == nil` is allowed and allocates.

**Semantics**

- `path` is the parent folder chain from the root (IDs or names that map to IDs). Empty `path` is the root.
- Entry/folder `data` is marshaled with `encoding/json` into the envelope `data` field. Use a pointer only if the caller needs `null` vs omit; a zero `T` marshals as that value.
- Create returns the envelope so callers do not need a second Get.
- Move renames on disk (new ID) and updates `id`/`name` inside JSON; bumps `updatedAt` except `UpdateFolderNameWithoutTimestamp`.
- Update replaces `data` and bumps `updatedAt`; does not rename.

**Errors** (`errors.Is` on package sentinels; OS errors are classified then dropped so `wrap` is zero-alloc):

| Sentinel | When |
|---|---|
| `ErrBadName` | name/id sanitizes to empty or reserved |
| `ErrBadPath` | parent missing, not a directory, or path escapes root |
| `ErrExist` | create/move target already exists |
| `ErrNotExist` | get/update/remove/move source missing |
| `ErrNotFile` / `ErrNotDirectory` | wrong object kind at that path |
| `ErrFolderCorrupted` | folder dir exists, `.info.json` unreadable |
| `ErrPermission` | OS permission denied |
| `ErrLock` | lock file open/flock failed |
| `ErrInternal` | unexpected OS/JSON failure |

Do not panic on missing files. Do not return `os.ErrNotExist` as the only error; wrap so callers can keep using `fsentry.ErrNotExist` like DeckBuilder uses `fsentry_error.ErrorNotExist`. Keep aliases `ErrorNotExist` = `ErrNotExist` only if a compatibility shim is requested; default to `Err*` names and document the mapping for DeckBuilder.

**Concurrency (two layers)**

1. **In-process:** `sync.RWMutex` on `*DB` (write = Lock, Get/List = RLock). Required because a shared lock-file FD does not serialize goroutines.
2. **Inter-process:** exclusive advisory lock on `.fsentry.lock` for every public method (reads included; keep it simple). Default on.

Order: mutex first, then flock; reverse on the way out. `WithNoLockFile()` skips layer 2 only.

**Logger.** If set, log unexpected sync/close failures. Do not log payload bytes.

---

## Allocations and benchmarks

Goal: **zero allocations** on paths we control. `encoding/json` will still allocate for arbitrary `T` (maps, slices, strings). That does not excuse extra copies in this library.

**Do**

- Reuse per-`*DB` scratch buffers (path join, `NameToID`, JSON encode) under the mutex; `sync.Pool` only if the mutex-held buffer is not enough.
- `NameToID(dst []byte, name string) []byte` appends into `dst`; public helper may wrap it. Measure with `testing.AllocsPerRun` and keep the append-to-buffer path at **0 allocs**.
- Join paths without `filepath.Join` on the hot path if that allocates; write into the scratch buffer, still using OS separators and `Clean` rules (no `..` escape).
- `GetBinary(name, buf)`: `Read` into `buf`, grow with `append` only when short.
- Return `List` / `Entry[T]` / `FolderInfo` by value. Do not `new` them.
- Success path: no `fmt.Errorf`, no `string(+)` in a loop, no `[]byte(str)` / `string(bytes)` if the bytes are only for I/O.

**Do not**

- Call `json.Marshal` then copy the buffer again; encode into the scratch/pool buffer.
- Intern every ID in a map “for speed” (hidden allocs + memory growth).
- Disable the mutex in benchmarks; **do** use `WithNoLockFile()` so the bench is not measuring `flock`.

**Benchmark files:** `*_bench_test.go` next to the **exported** helpers (no build tag unless OS-specific). `b.ReportAllocs()`. Do not add `Benchmark*` for unexported functions (`wrap`, `mapError`, `mapOS`); keep `testing.AllocsPerRun` in unit tests for those.

| Benchmark | What |
|---|---|
| `CreateFile` / `CreateFolder` | report OS allocs |
| `Write` empty buffer | **0 allocs** |
| `Write` (reused buffer) | **0 allocs** beyond `os.File.Write` |
| `RenameFile` / `RenameFolder` | report OS allocs |
| `NameToID` into a reused buffer | 0 allocs |
| `Read` into a sized buffer | **0 allocs** |
| `OpenRead` / `OpenWrite` / `Close` / `Sync` / `Stat` / `ReadDir` / `RemoveFile` / `RemoveFolder` | report OS allocs |
| `CreateBinary` / `UpdateBinary` of a fixed `[]byte` | 0 extra besides OS |
| `GetEntry[struct{…}]` of a small fixed struct | report allocs; fight extras outside `json` |
| `List` of a small directory | report allocs (OS `Readdir` will allocate names) |
| `CreateEntry` small struct | report allocs (JSON marshal budget) |

```bash
make bench   # go test -bench=. -benchmem (Makefile uses '^$$' so make does not eat '$')
```

If a change raises allocs/op on `NameToID` or `GetBinary` (sized buf), treat it as a regression. List/JSON benches are recorded so we do not slide silently; do not fail CI on JSON allocs until we have a replacement encoder.

---

## Stack

| Layer | Choice |
|---|---|
| Language | Go **1.27+** (`go 1.27` in `go.mod`; generic methods on `*DB`) |
| Public API | package `fsentry` (`*DB` methods, including `CreateEntry[T]` / `GetEntry[T]`) |
| JSON | `encoding/json` into reused buffers |
| Lock file | `golang.org/x/sys` (`unix.Flock`, Windows `LockFileEx`) |
| Copy folder | `github.com/otiai10/copy` (already used by old fsentry / DeckBuilder) |
| Logging | optional `Logger` interface (`Debug/Info/Warn/Error`); no logrus |
| Tests | `go test -race`; `t.TempDir()`; `WithNoLockFile()` unless testing the lock |

No CGO. No Wails. No HTTP. `internal/` stays importable only from this module.

---

## Build & test

```bash
make help
make test          # go test -race ./...
make test-integration
make bench         # allocs/op; uses WithNoLockFile
make lint          # golangci-lint
make doc PKG=.     # go doc
make doc-all PKG=.
```

First slice: `go test -race ./...` must pass on darwin and linux with Go 1.27.x. Windows behavior (reserved names, ACL) is covered with unit tests for `NameToID` plus build-tagged integration if we keep Windows IO mapping. Do not compile this module with Go 1.26 or older (generic methods are a 1.27 language feature).

---

## Where things live

This file stays lean. **[README.md](README.md)** is the short user entry (what the library is, install, tiny example, status). This file is the agent/implementation spec. Architecture decisions go in **[docs/architecture](docs/architecture/INDEX.md)**. Use cases go in **[docs/use-cases](docs/use-cases/INDEX.md)** only after the code exists.

| If you need… | Read |
|---|---|
| What the library is, install, example | **[README.md](README.md)** |
| On-disk format, API, layout | this file |
| Why a decision was made | `docs/architecture/` |
| Consumer that must keep working | `../DeckBuilder/internal/db/` |
| Previous implementation (do not copy tree) | `../fsentry` |

### Intended tree

This is a **library**, not an application. No `cmd/` until someone asks for a CLI. No `package main`.

```text
.
├── Makefile
├── README.md
├── CURSOR.md
├── LICENSE
├── go.mod                      # module github.com/HardDie/fsentry; go 1.27
├── fsentry.go                  # New, options, *DB, Init/Drop/List
├── folder.go                   # folder methods
├── entry.go                    # (*DB) CreateEntry[T], GetEntry[T], …
├── binary.go                   # binary methods (GetBinary into buf)
├── types.go                    # List, Entry[T], FolderInfo
├── errors.go                   # sentinels + Wrap
├── quoted_string.go            # QuotedString for on-disk name
├── doc.go                      # package comment
├── fsentry_test.go             # black-box tests; WithNoLockFile
├── fsentry_bench_test.go       # ReportAllocs
├── lock_integration_test.go    # //go:build integration; lock file on
├── internal/
│   ├── fs/                     # OS create file/folder; map errors (unix + windows)
│   ├── lock/                   # advisory lock file (unix + windows)
│   ├── name/                   # NameToID into []byte
│   └── jsonutil/               # marshal envelope, pretty flag, scratch buf
├── examples/
│   └── basic/                  # same story as old README (folder + entry + binary)
├── docs/
│   ├── architecture/
│   └── use-cases/              # UC files only after the code exists
└── .github/workflows/
    └── test.yml
```

**Layout rules**

- Exported surface lives at module root. `internal/` is OS and helpers only.
- Do **not** recreate old `internal/service` + `internal/folder/service` + `internal/repository` triples. One implementation per object kind is enough.
- Tests for public behavior sit next to the public files (`fsentry_test.go`, or `folder_test.go` in package `fsentry_test`). Tests for `NameToID` sit in `internal/name`. Default `New` in tests: `WithNoLockFile()`.
- `examples/basic` must compile (`go test ./examples/...` or a make target). It is documentation, not a dump of DeckBuilder.
- Do not add `pkg/`. Do not add `internal/entity` that duplicates public types.

### Go package documentation (godoc)

Every package must have documentation `go doc` can print.

**Required**

- Package comment (`doc.go`): filesystem-as-database, three object kinds, not a generic file helper.
- Every exported type, func, method, const, var has a comment starting with the name.
- Document `errors.Is` sentinels, folder `path` chains, `WithNoLockFile`, and Go 1.27 generic methods (`Entry[T]` on `*DB`, not on an interface).
- `QuotedString` documents the double-encoded JSON and why it exists (compat).

**Do not**

- Leave `TODO` or empty `// Foo …` on exports.
- Put DeckBuilder domain names in godoc except as a one-line “used by” mention.
- Put install steps in godoc; those belong in README.

```bash
make doc PKG=.
make doc-all PKG=.
```

### Tests

Every package **must** have unit tests. Integration tests when talking to the real disk.

| Kind | Files | Build tag | Command |
|---|---|---|---|
| Unit | `*_test.go` next to code | none | `make test` |
| Benchmark | `*_bench_test.go` | none | `make bench` |
| Integration (disk) | `*_integration_test.go` | `//go:build integration` | `make test-integration` |

**Rules**

- Table-driven tests. Cover happy path and `ErrBadName`, `ErrBadPath`, `ErrExist`, `ErrNotExist`, reserved Windows names, unicode letters in names, pretty vs compact JSON, nested paths, duplicate folder, list corrupted folder, binary round-trip, typed `GetEntry[T]`.
- Use `t.TempDir()`; never write into the repo as `test/` (old tests did that).
- **`WithNoLockFile()` on every unit, integration, and benchmark store** unless the test is specifically for the lock file.
- Race detector on. Mutex tests: concurrent create of different names under one `*DB` with the lock file off.
- Lock tests (integration): two `*DB` values, same root, lock file **on**; one holds a write while the other must not observe a torn file. Optional second-process test later.
- `testing.AllocsPerRun` for `NameToID` (buffer) and `GetBinary` (sized buf): **0**.
- Integration tests skip nothing on CI: they only need a temp directory.
- Porting is incomplete without tests, godoc, benchmarks for the hot paths, and a use-case file for that behavior.

CI (every push and PR): Go **1.27.x**, `go test -race ./...` then `go test -tags=integration ./...`.

### Internal packages

Collapse by **object**, not by layer.

| Old fsentry | This library |
|---|---|
| `main.go` + `pkg/fsentry` + `pkg/fsentry_error` | root `fsentry` |
| `internal/service` + `internal/folder/service` + `internal/repository/folder` | `folder.go` + `internal/fs` |
| `internal/entity` | public `types.go` |
| `internal/utils.NameToID` | `internal/name` (append into `[]byte`) |
| `internal/io` + `internal/fs/storage` | `internal/fs` (one OS adapter) |
| process mutex only | mutex + `internal/lock` |

**Import direction:** public `fsentry` → `internal/fs`, `internal/lock`, `internal/name`, `internal/jsonutil`. Internals must not import the public API in a cycle; they may use stdlib plus `x/sys` and `copy`.

**`internal/fs` owns syscalls.** Other packages do not call `os.OpenFile`, `os.Mkdir`, `os.File.Write`/`Read`/`Close`/`Sync`, `os.Rename`, `os.Remove`/`RemoveAll`, `os.ReadDir`, or `os.Stat`.

| Helper | Behavior |
|---|---|
| `CreateFile` | `O_EXCL`, empty write-only handle |
| `OpenRead` / `OpenWrite` | existing file; write truncates, no `O_CREATE` |
| `Write` / `Read` | all bytes / until EOF into caller buf |
| `Close` / `Sync` | handle close / fsync |
| `CreateFolder` | `Mkdir` only (no parents) |
| `RenameFile` / `RenameFolder` | `os.Rename` (no copy) |
| `RemoveFile` | `unlink` only (not rmdir) |
| `RemoveFolder` | `Stat` then recursive `RemoveAll`; missing is `ErrNotExist` |
| `ReadDir` / `Stat` | one directory listing / `FileInfo` |

Helpers return only: `ErrExist`, `ErrNotExist`, `ErrPermission`, `ErrNotDirectory`, `ErrIsDirectory`, `ErrNoSpace`, `ErrReadOnly`, `ErrInternal`. The OS error is classified then dropped (`wrap` is zero-alloc). Higher code `errors.Is` those sentinels; it does not inspect `syscall.Errno`. `mapOS` is split `error_unix.go` / `error_windows.go`.

---

## Key facts to keep in mind

- **This is a library.** If you are about to add `main()`, HTTP, or a game type, stop.
- **Disk format is the API.** A user can zip the root and copy it. Do not store truth only in memory.
- **DeckBuilder is the compatibility test.** On-disk trees stay readable. Call sites change: `db.GetEntry[Settings]("settings")` instead of `GetEntry` + `json.Unmarshal`. Folders/binaries keep method names.
- **IDs are filesystem names; names are display strings.** Never put unsanitized user text in a path.
- **`.info.json` is not an entry.** List must not return it as `Entries`. Same for `.fsentry.lock`.
- **Cards in DeckBuilder are not entries.** They are a map inside the `cards` folder payload plus per-id binaries. This library still has no card type.
- **Entries are `Entry[T]` on `*DB` methods (Go 1.27).** No `data any` on the entry API. No `IFSEntry` interface: generic methods cannot be interface methods.
- **Finish `List`.** Returning `nil, nil` (old rewrite) is a bug.
- **Timestamps are UTC `time.Time` in the public struct.** Pointers were a leftover; do not export them.
- **Pretty JSON is optional** and must not change semantics.
- **Lock file on in apps, off in tests.** Two processes, one root, without the lock is unsupported.
- **Zero extra allocs on `NameToID` and buffered `GetBinary`.** Benchmark those. JSON will allocate; do not add a second copy on top.
- **Smallest correct implementation.** One mutex, one lock file, one FS adapter, three object files. No repository/service/entity sandwich.

---

## Agent working agreements

- Match Go style: small functions, `errors.Is`, `slog`-shaped logger interface, no init() side effects.
- Keep this file updated when the on-disk format, exported methods, or directory layout change.
- **Keep [README.md](README.md) short.** Update it in the same change when install, example, or status change. Do not dump this file into the README.
- **Use cases only after code exists.** Add `docs/use-cases/uc-NN-….md` from the template when a behavior lands; set the row to Done in `docs/use-cases/INDEX.md`. Do not invent UC files for unimplemented methods.
- **Godoc on every package.** Verify with `go doc -all` before calling a package done.
- **Tests on every package.** Unit always; integration when hitting disk; benchmarks for hot paths. CI must stay green. Tests use `WithNoLockFile()` except lock tests.
- **New core decisions get an ADR** in `docs/architecture/` (next number, update INDEX). Do not leave accepted decisions only in chat.
- Prefer finishing one object kind (folder or entry or binary) end-to-end over a large rewrite.
- Do not copy files from `../fsentry` wholesale. Read them, then write the new tree.
- Do not add DeckBuilder types to this module.
- If a change would break reading existing `.info.json` / `*.json` / `*.bin` trees, it needs an explicit request and an ADR.
