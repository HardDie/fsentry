# 7. Advisory lock file, stamp, and steal after timeout

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

Two processes must not tear JSON in the same store. Old fsentry used an in-process mutex only. A pid file with busy-wait is easy to get wrong. `flock` / `LockFileEx` release when the process dies, but a hung process (or a filesystem that keeps the lock) can block waiters forever.

## Considered options

1. **In-process `sync.RWMutex` only.**
2. **Pid / lock file contents without OS advisory locks.**
3. **Exclusive OS lock on `<root>/.fsentry.lock` for every public method**, plus the mutex (the lock FD does not serialize goroutines). Waiters `TryLock`; if the 8-byte unix-nano stamp (or mtime if missing) is older than a timeout, **steal** by unlinking and locking a new inode. Default timeout 10 minutes, overridable at open / `WithLockTimeout`. Tests pass `WithNoLockFile()` unless the test is about locking.

## Decision

Use option 3.

`internal/fs` owns `OpenLock`, blocking `Lock`, `TryLock` (`ErrBusy` if held), `Unlock`. Windows locks 1 byte at offset 8 so waiters can read the stamp. `internal/lock` writes the stamp, polls ~100ms, and steals. Order later on `*DB`: mutex first, then flock; reverse on the way out. `Drop` unlocks, closes, then removes the root. `List` skips `.fsentry.lock`.

A crash usually drops the OS lock immediately; steal is for hang / stale inode. A live holder past the timeout can race the waiter (split brain). Unlink while the holder still has the file open may fail on Windows; then steal returns the unlink error.

Do not spawn extra processes unless a lock test cannot use two handles in one process.

## Consequences

### Positive

* Production stores are exclusive across processes without a lock daemon.
* Tests stay fast and race-detector-friendly with the lock off.

### Negative and risks

* Steal is unsafe if a healthy operation holds the lock longer than the timeout.
* Two `*DB` handles in one process still need the mutex.

### Neutral

* Lock is on for the duration of each public operation, including reads, to keep the rule simple.
