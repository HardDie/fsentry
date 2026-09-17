# UC-19: Steal a stale lock

**Module:** `internal/lock`  
**Status:** Done  
**Actors:** `*DB` public methods (after the in-process mutex)  
**Goal:** Do not wait forever if the previous process crashed (or hung) with the lock held  
**Preconditions:** Two `Open` handles on `<root>/.fsentry.lock`. Timeout is `DefaultTimeout` (10 minutes) unless `Open` was given a positive duration (`WithLockTimeout` on `*DB` Init).

## Main scenario (happy path)

1. Process A `Lock`s: OS exclusive lock, then 8-byte unix-nano stamp at offset 0.
2. Process A exits without `Unlock` (crash). The OS drops the lock when the handle closes.
3. Process B `Lock`s: `TryLock` succeeds; B writes a new stamp.

## Alternative scenarios and errors

* **2a. A is still alive, stamp younger than timeout:** B retries `TryLock` every 100ms until A `Unlock`s.
* **2b. A still holds the OS lock, stamp (or mtime if no stamp) older than timeout:** B closes its handle, `RemoveFile`s the path (new inode), `OpenLock`s, and `TryLock`s again. A live holder past the timeout races B (split brain).
* **2c. Unlink fails (file still open, typical on Windows):** reopen and return the unlink error (`ErrPermission` / `ErrBusy` / `ErrInternal`).
* **2d. Open / stamp I/O fails:** mapped sentinels from `internal/fs`.

## Postconditions

* After a crash, the next `Lock` succeeds as soon as the OS has released the handle (usually immediately). After a hang longer than timeout, waiters steal. The timeout is a const default and is chosen when the lock file is opened (DB Init).
