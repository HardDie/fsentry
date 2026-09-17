# UC-04: Write data to a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Higher fsentry layers (create/update entry JSON, create/update binary)  
**Goal:** Write every byte of a buffer to an already-created file, or fail with a known sentinel  
**Preconditions:** `file` is an open writeable `*os.File` from `CreateFile` (or later update). Callers never use `os.File.Write` for payload.

## Main scenario (happy path)

1. The caller invokes `Write(file, data)`.
2. The helper loops `file.Write` until all bytes are written (short writes are retried).
3. It does not create the file, seek, sync, or close.

## Alternative scenarios and errors

* **1a. `data` is empty or nil:** success, no syscall required.
* **2a. Disk full or quota:** `ErrNoSpace`.
* **2b. Read-only filesystem:** `ErrReadOnly`.
* **2c. Permission / closed / unknown:** `ErrPermission` or `ErrInternal` as mapped. Closed files typically become `ErrInternal` (EBADF).

## Postconditions

* On success, the bytes are in the file at the current offset. On failure, the error is a sentinel from UC-03. `Write` does not allocate on the empty-buffer path.
