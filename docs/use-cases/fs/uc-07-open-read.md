# UC-07: Open a file for reading

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** GetEntry, GetFolder `.info.json`, GetBinary  
**Goal:** Open an existing file with `O_RDONLY`, or fail with a known sentinel  
**Preconditions:** Path exists as a file. Callers never use `os.Open` / `os.OpenFile` for reads.

## Main scenario (happy path)

1. The caller invokes `OpenRead(path)`.
2. The helper opens with `O_RDONLY` and does not create the file.
3. It returns `*os.File` at offset 0. The caller `Close`s it.

## Alternative scenarios and errors

* **2a. Missing:** `ErrNotExist`.
* **2b. Parent component is a file:** `ErrNotDirectory` (Windows may report `ErrNotExist`).
* **2c. Permission:** `ErrPermission`.

## Postconditions

* On success, the handle is readable. Errors are sentinels from UC-03.
