# UC-16: Open a lock file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** `Init` / `*DB` (path `<root>/.fsentry.lock`)  
**Goal:** Open a file for advisory locking, creating it if missing  
**Preconditions:** Parent directory exists. Callers never use `os.OpenFile` for the lock file.

## Main scenario (happy path)

1. The caller invokes `OpenLock(path)`.
2. The helper opens with `O_RDWR|O_CREATE` (no truncate, no `O_EXCL`).
3. It returns a handle. The caller `Lock`s, `Unlock`s, then `Close`s.

## Alternative scenarios and errors

* **2a. Parent missing:** `ErrNotExist`.
* **2b. Permission:** `ErrPermission`.

## Postconditions

* On success, `path` exists. Closing without `Unlock` still drops the OS lock (handle close). Errors are sentinels from UC-03 (plus `ErrLock` for lock syscalls).
