# UC-09: Read from a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** GetEntry, GetFolder, GetBinary  
**Goal:** Read from the current offset until EOF into a caller buffer  
**Preconditions:** `file` is open for reading. Callers never use `os.File.Read` / `io.ReadAll` for payloads.

## Main scenario (happy path)

1. The caller invokes `Read(file, buf)`.
2. The helper reads until EOF. `io.EOF` is success.
3. If `cap(buf)` is large enough, no extra allocation. Otherwise the buffer is grown.

## Alternative scenarios and errors

* **1a. Empty file / empty remaining:** success, returned slice length 0.
* **2a. Closed / I/O error:** `ErrInternal` (or `ErrPermission` / `ErrNotExist` if the OS reports those).

## Postconditions

* Returned slice holds the bytes read. Offset is at EOF. Errors are sentinels from UC-03.
