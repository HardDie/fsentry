# 6. `internal/fs` files follow package `os`

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

The OS adapter started as one file per helper (`create.go`, `open.go`, `read.go`, …). That is unlike the standard library and made the package look like a list of wrappers rather than a small `os` façade. Stamp encoding for the lock lease also sat next to `OpenLock`.

## Considered options

1. **Keep one exported function per file.**
2. **Group like `os`:** `file.go` (bytes and names), `dir.go` (directories), `lock.go` + `lock_unix.go` / `lock_windows.go`, `error.go` + `error_unix.go` / `error_windows.go`.
3. **Single `fs.go`** plus build-tagged lock/error files.

## Decision

Use option 2.

Lease stamp bytes and steal policy live in `internal/lock` (`stamp.go`, `lock.go`), using `WriteAt` / `ReadAt` / `Truncate` from `internal/fs`. Package `fs` does not import `io/fs` for `FileMode`; it uses `os.FileMode`.

Build tags are `unix` / `windows` (not a separate darwin file when the syscall is the same).

## Consequences

### Positive

* Contributors look for `CreateFile` in `file.go`, not a 20-file list.
* Lock protocol is not mixed into the syscall wrappers.

### Negative and risks

* Test files may still be named after an older helper (`create_test.go`); they stay in package `fs` and remain valid.

### Neutral

* This does not change helper names or sentinels.
