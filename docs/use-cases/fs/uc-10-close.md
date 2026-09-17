# UC-10: Close a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Any helper that received `*os.File`  
**Goal:** Close the handle and map OS errors  
**Preconditions:** `file` came from `CreateFile`, `OpenRead`, or `OpenWrite`. Callers never use `os.File.Close`.

## Main scenario (happy path)

1. The caller invokes `Close(file)`.
2. The helper closes the descriptor. The file on disk remains.

## Alternative scenarios and errors

* **2a. Already closed / invalid:** `ErrInternal`.

## Postconditions

* On success, `file` must not be used again. Errors are sentinels from UC-03.
