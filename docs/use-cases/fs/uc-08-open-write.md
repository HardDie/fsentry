# UC-08: Open a file for writing (truncate)

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** UpdateEntry, UpdateFolder, UpdateBinary  
**Goal:** Open an existing file with `O_WRONLY|O_TRUNC`, or fail with a known sentinel  
**Preconditions:** Path exists as a file. Callers never use `os.OpenFile` for updates.

## Main scenario (happy path)

1. The caller invokes `OpenWrite(path)`.
2. The helper opens with `O_WRONLY|O_TRUNC` and **does not** set `O_CREATE` (missing → `ErrNotExist`).
3. Existing content is discarded. The caller `Write`s then `Close`s (and `Sync` when doing atomic replace).

## Alternative scenarios and errors

* **2a. Missing:** `ErrNotExist`.
* **2b. Path is a directory:** `ErrIsDirectory` (Windows may report `ErrPermission`).
* **2c. Permission / read-only:** `ErrPermission` / `ErrReadOnly`.

## Postconditions

* On success, the file length is 0 and the handle is writable. Errors are sentinels from UC-03.
