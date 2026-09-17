# UC-17: Exclusive lock and unlock

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** every public `*DB` method (after the in-process mutex)  
**Goal:** Take and release an exclusive advisory lock so two processes cannot run store operations at once  
**Preconditions:** `file` came from `OpenLock`. Two handles to the same path (two processes or two `OpenLock` calls) are distinct FDs.

## Main scenario (happy path)

1. The caller invokes `Lock(file)` (`flock` `LOCK_EX` / `LockFileEx` exclusive). It blocks until the lock is free.
2. The caller runs the store operation.
3. The caller invokes `Unlock(file)`.

## Alternative scenarios and errors

* **1a. Same process, same FD:** a second `Lock` on the same handle is a no-op or succeeds (OS-defined); goroutines are **not** serialized by this lock. Use `sync.RWMutex` as well.
* **1b. Second handle while first holds the lock:** `Lock` blocks until `Unlock` (or close) on the first handle.
* **1c. Deadlock / lock error:** `ErrLock`.
* **1d. Permission / closed handle:** `ErrPermission` / `ErrInternal`.

## Postconditions

* Exclusive inter-process exclusion for the locked region (whole-file advisory lock on unix; 1 byte at offset 8 on Windows so the stamp is readable). Not a mutex for goroutines sharing one `*os.File`. Errors are sentinels; `EDEADLK` / blocking `ERROR_LOCK_VIOLATION` map to `ErrLock`. Waiters that must not block use `TryLock` (UC-18) and stale steal (UC-19).
