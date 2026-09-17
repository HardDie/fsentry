# UC-18: Non-blocking try-lock

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** `internal/lock` (before steal or wait)  
**Goal:** Learn whether the exclusive lock is free without blocking  
**Preconditions:** `file` came from `OpenLock`.

## Main scenario (happy path)

1. The caller invokes `TryLock(file)`.
2. The lock is free. The helper takes it (`flock` `LOCK_EX|LOCK_NB` / `LockFileEx` with `FAIL_IMMEDIATELY`) and returns nil.

## Alternative scenarios and errors

* **2a. Another handle holds the lock:** `ErrBusy` (no wait).
* **2b. Deadlock / other lock error:** `ErrLock`.
* **2c. Permission / closed handle:** `ErrPermission` / `ErrInternal`.

## Postconditions

* Success means this handle owns the exclusive lock. `ErrBusy` is not `ErrLock`; waiters use it to inspect the stamp and decide steal vs retry. Windows locks 1 byte at offset 8 so the 8-byte stamp stays readable.
