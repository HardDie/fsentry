# 8. Benchmarks on exported helpers; `wrap` returns the sentinel only

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

The store should not add copies on paths it controls (`NameToID`, sized `GetBinary`, `Write`/`Read` into caller buffers). `encoding/json` will still allocate. Early benches on unexported `mapError` / `wrap` measured classify, then `wrap` was changed until allocs hit zero. `fmt.Errorf` was 5 allocs; `errors.Join` was 1; returning the sentinel only is 0. Create/rename still allocate inside the standard library (`os.File`, path syscalls); those benches **report** OS allocs rather than claiming 0.

## Considered options

1. **Benchmark every internal function**, including `wrap`.
2. **No alloc discipline** until the public `*DB` exists.
3. **`Benchmark*` only next to exported helpers** (`b.ReportAllocs()`). Unexported classify stays at 0 via `AllocsPerRun` in unit tests. Do not fail CI on JSON allocs until there is another encoder. Benchmarks use `WithNoLockFile()` so they are not measuring `flock`. Do not disable the in-process mutex in benches.

## Decision

Use option 3.

Treat a rise in allocs/op on `NameToID` or sized `GetBinary` as a regression. Makefile `bench` uses `-run '^$$'` so Make does not eat `$`.

## Consequences

### Positive

* Alloc work focuses on API and reusable buffers, not on wrapping errno.
* OS-heavy helpers have recorded benches without fake “0 alloc” claims.

### Negative and risks

* JSON and `Readdir` will always show allocs; those numbers are for drift, not a gate yet.

### Neutral

* Per-`*DB` scratch buffers (when `*DB` lands) beat a `sync.Pool` until the mutex-held buffer is not enough.
