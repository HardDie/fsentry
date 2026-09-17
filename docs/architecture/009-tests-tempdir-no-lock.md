# 9. Tests: race detector, `t.TempDir()`, lock off unless testing the lock

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

Old fsentry tests wrote into a `test/` directory in the repo. Lock files and `flock` fight the race detector and slow unit tests. Close benches that open `b.N` FDs hit `EMFILE`.

## Considered options

1. **Keep repo-relative `test/` fixtures** like the old module.
2. **Always enable the lock file** so tests match production.
3. **`t.TempDir()` everywhere.** `go test -race` on unit tests. Default stores use `WithNoLockFile()`. Lock tests omit that option (or use `WithLockFile()`) and use two handles on one root. Integration tests that need the real disk use `//go:build integration`. Close/lock benches open one FD (or open+close per iteration), not `b.N` leaked FDs.

## Decision

Use option 3.

Every package has unit tests. Table-driven cases cover sentinels that already exist in `internal/fs`. Platform `mapOS` tests are build-tagged. CI: Go 1.27.x, `go test -race ./...`, then integration tags when those files exist.

## Consequences

### Positive

* No junk directories in git.
* Race detector stays usable on the default test path.

### Negative and risks

* Tests with the lock off cannot catch inter-process tearing; that is the job of lock-specific tests.

### Neutral

* Two-process tests are optional later; two `OpenLock` / two `*DB` in one process is the default way to prove exclusion.
