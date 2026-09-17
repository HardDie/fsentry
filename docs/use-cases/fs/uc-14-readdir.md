# UC-14: Read a directory listing

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** public `List`  
**Goal:** Return the names in one directory (not recursive)  
**Preconditions:** Path is a directory. Callers never use `os.ReadDir` / `os.Open`+`Readdir`.

## Main scenario (happy path)

1. The caller invokes `ReadDir(path)`.
2. The helper returns `[]os.DirEntry` (files and subdirs). Classification into folders/entries/binaries is not done here.

## Alternative scenarios and errors

* **2a. Missing:** `ErrNotExist`.
* **2b. Path is a file:** `ErrNotDirectory`.
* **2c. Permission:** `ErrPermission`.

## Postconditions

* On success, entries are the immediate children. Errors are sentinels from UC-03.
