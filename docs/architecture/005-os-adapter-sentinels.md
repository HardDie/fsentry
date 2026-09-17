# 5. OS adapter maps errors to sentinels; `wrap` is zero-alloc

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

Create, open, flock, and unlink return platform `Errno` values (`ENOENT` vs `ERROR_PATH_NOT_FOUND`). Higher layers (and later `*DB`) must not import `syscall` to classify them. Wrapping with `fmt.Errorf("%w: %w", sentinel, err)` or `errors.Join` allocated (5 allocs, then 1). Benchmarks on the adapter’s hot path need a 0-alloc classify.

## Considered options

1. **Return `*os.PathError` unchanged** and let callers use `errors.Is(err, os.ErrNotExist)`.
2. **Sentinel + chained OS error** (`fmt.Errorf` / `errors.Join`) so `%+v` still shows errno.
3. **Classify, then return the sentinel only.** `mapError` / `mapOS` (unix vs windows files). Higher code `errors.Is` the fixed list. No `syscall.Errno` outside `internal/fs`.

## Decision

Use option 3.

Helpers in `internal/fs` are the only place that calls `os.OpenFile`, `Mkdir`, `Write`/`Read`/`Close`/`Sync`/`Truncate`/`WriteAt`/`ReadAt`, `Rename`, `Unlink`/`RemoveAll`, `ReadDir`, `Stat`, `flock` / `LockFileEx`. `RemoveFile` uses `syscall.Unlink` (not `os.Remove`) so a directory is `ErrIsDirectory` rather than a recursive delete. `OpenWrite` is `O_WRONLY|O_TRUNC` without `O_CREATE`.

Sentinels: `ErrExist`, `ErrNotExist`, `ErrPermission`, `ErrNotDirectory`, `ErrIsDirectory`, `ErrNoSpace`, `ErrReadOnly`, `ErrLock`, `ErrBusy`, `ErrInternal`. Windows `mapOS` uses `golang.org/x/sys/windows` constants cast to `syscall.Errno`.

`wrap` ignores the original error and returns the sentinel so classify stays 0 alloc. Unit tests use `testing.AllocsPerRun` for `wrap` / `mapError`; there are no `Benchmark*` for those unexported functions.

## Consequences

### Positive

* The rest of the module is portable in the `errors.Is` sense.
* Classify does not allocate.

### Negative and risks

* Lost errno in the returned value; debugging a novel OS error looks like `ErrInternal`.
* Public `fsentry` sentinels will alias or re-export this list later; names are `Err*`, not old `ErrorNotExist`, unless a compatibility shim is requested.

### Neutral

* `internal/fs` is not a general filesystem library; it exists so this store does not scatter `os` calls.
