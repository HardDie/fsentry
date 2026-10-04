# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

1. What this is
 1. Go library `github.com/HardDie/fsentry`.
 2. Treats a directory tree as a small database.
 3. A library only: no `main`, no `cmd/`, no HTTP, no domain types.
 4. Requires Go 1.27+ (generic methods on `*DB`).
 5. Clean rewrite of the old fsentry at `../fsentry`.
 6. Do not copy the old package layout.
 7. Compatibility target: `../DeckBuilder/internal/db/`.
 8. Existing DeckBuilder data trees must stay readable.

2. Source of truth
 1. `CURSOR.md` is the full implementation spec.
 2. It covers on-disk format, public API, errors, locking, alloc rules.
 3. Read it before changing behavior or format.
 4. Update it when format, exported methods, or layout change.
 5. ADRs live in `docs/architecture/` (see `INDEX.md`).
 6. A new core decision gets the next ADR number.
 7. Wiki sources live in `docs/wiki/`.
 8. Use cases go in `docs/use-cases/` only after the code exists.
 9. Keep `README.md` short.

3. Commands
 1. `make test` runs unit tests with `-race`.
 2. `make test-integration` adds `-tags=integration`.
 3. `make bench` runs benchmarks only, with allocs/op.
 4. `make lint` runs golangci-lint.
 5. `make examples` runs every `examples/*/main.go`.
 6. `make ci` runs unit, integration, vet, and fmt.

```bash
go test -race -run TestCreateFolderPretty .
go test -race -run TestLock ./internal/lock
go test -tags=integration -run TestFullPublicFlow ./integration
go test -bench=BenchmarkGetBinary -benchmem -run '^$' .
```

4. CI
 1. Workflow: `.github/workflows/test.yml`.
 2. Linux and macOS: race unit tests, then integration.
 3. Windows: unit tests without race, then integration.
 4. Ubuntu also checks gofmt, vet, `go mod tidy` diff, examples.
 5. golangci-lint runs as a separate job.

5. Layout
 1. One public package `fsentry` at module root.
 2. One file per object kind: `folder.go`, `entry.go`, `binary.go`.
 3. `archive.go` holds Export/Import and ExportFolder/ImportFolder.
 4. `validate.go` walks the tree and reports `Problem`s.
 5. `internal/fs` is the only package that calls `os` file syscalls.
 6. `internal/lock` owns the advisory lock file, stamp, and steal.
 7. `internal/name` owns `NameToID` (append into `[]byte`).
 8. `internal/jsonutil` owns envelope marshaling.
 9. Internals never import the root package.
 10. No `pkg/`, no service/repository/entity layers.

6. Request flow
 1. Every public method wraps its body in `db.withLock` (`init.go`).
 2. `withLock` takes the `sync.RWMutex`, then the lock file.
 3. It releases in reverse order.
 4. `WithNoLockFile()` skips only the lock file.
 5. `db.ensurePath` walks the parent chain from the root.
 6. Each parent must be a dir with a valid `.info.json`.
 7. Missing parent is `*BadPathError` (`errors.Is` `ErrBadPath`).
 8. Invalid parent is `ErrFolderCorrupted`.
 9. Names become disk IDs via `NameToID` (`"My Notes"` → `my_notes`).
 10. `db.clock()` gives UTC time; tests may override `db.now`.
 11. Reads take the lock file shared (`RLock`); writes take it exclusive.
 12. Goroutines on one `*DB` count as one reader (ADR 012).
 13. A reader stamps without fsync; a writer still fsyncs.

7. On-disk format
 1. Folder: directory named by ID plus `.info.json`.
 2. Entry: `<id>.json`.
 3. Binary: `<id>.bin`, raw bytes.
 4. Envelope fields: `id`, `name`, `createdAt`, `updatedAt`, `data`.
 5. `name` is a `QuotedString` (double-encoded JSON string).
 6. Lock file is `<root>/.fsentry.lock`.
 7. `List` and `Export` skip `.info.json` and the lock file.
 8. A format change needs an explicit request and an ADR.

8. Errors
 1. `internal/fs` maps OS errors to its own sentinels.
 2. The OS error is dropped so `wrap` stays zero-alloc.
 3. Callers use `errors.Is` on `fsentry.Err*` sentinels.
 4. Do not inspect `syscall.Errno` outside `internal/fs`.

9. Tests
 1. Every store in tests uses `t.TempDir()` and `WithNoLockFile()`.
 2. Only lock tests leave the lock file on.
 3. Lock tests use two `*DB` handles on one root, one process.
 4. Prefer table-driven tests.
 5. Integration tests live in `integration/` with `//go:build integration`.
 6. Every exported func needs an `Example*` in `example_test.go`.
 7. Each example needs an `Output:` block.
 8. Every exported symbol needs a godoc comment starting with its name.

10. Allocations
 1. `NameToID` into a reused buffer must stay at 0 allocs.
 2. `GetBinary` with a sized buffer must stay at 0 allocs.
 3. A higher allocs/op on those two is a regression.
 4. Benchmarks go in `*_bench_test.go` with `b.ReportAllocs()`.
 5. Benchmark exported helpers only.
 6. Check unexported helpers with `testing.AllocsPerRun`.
 7. No `fmt.Errorf` or extra buffer copies on success paths.

11. Spec drift
 1. CURSOR.md lists `otiai10/copy` for folder copy.
 2. The code uses its own `copyDir` in `folder.go`.
 3. `go.mod` depends only on `golang.org/x/sys`.
