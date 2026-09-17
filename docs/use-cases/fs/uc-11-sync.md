# UC-11: Sync a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Atomic JSON write (temp → write → Sync → RenameFile)  
**Goal:** Flush data and metadata to stable storage  
**Preconditions:** `file` is open for writing. Callers never use `os.File.Sync`.

## Main scenario (happy path)

1. The caller invokes `Sync(file)` after `Write`.
2. The helper calls `fsync` / equivalent.

## Alternative scenarios and errors

* **2a. Closed handle:** `ErrInternal`.
* **2b. I/O / read-only:** `ErrInternal` / `ErrReadOnly`.

## Postconditions

* On success, written bytes are on durable storage (as far as the OS guarantees). Errors are sentinels from UC-03.
