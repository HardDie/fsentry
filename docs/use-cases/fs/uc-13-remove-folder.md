# UC-13: Remove a folder

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** RemoveFolder, Drop  
**Goal:** Recursively delete a directory and its contents  
**Preconditions:** Path is a directory. Callers never use `os.RemoveAll` / `os.Remove` for folders.

## Main scenario (happy path)

1. The caller invokes `RemoveFolder(path)`.
2. The helper `Stat`s: missing → `ErrNotExist`; not a dir → `ErrNotDirectory`.
3. It then `RemoveAll`s the tree (children included).

## Alternative scenarios and errors

* **2a. Missing:** `ErrNotExist` (`os.RemoveAll` alone would return nil; we do not).
* **2b. Path is a file:** `ErrNotDirectory`.
* **2c. Permission:** `ErrPermission`.

## Postconditions

* On success, `path` does not exist. Errors are sentinels from UC-03.
