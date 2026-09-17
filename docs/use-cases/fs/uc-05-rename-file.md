# UC-05: Rename a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Higher fsentry layers (`MoveEntry`, `MoveBinary`)  
**Goal:** Rename or move a file in one syscall, or fail with a known sentinel  
**Preconditions:** Source file exists. Destination parent exists. Callers never use `os.Rename` for files.

## Main scenario (happy path)

1. The caller invokes `RenameFile(oldpath, newpath)`.
2. The helper calls `os.Rename`. It does not copy bytes and does not create missing parents.
3. The file is available only at `newpath`.

## Alternative scenarios and errors

* **2a. Source missing:** `ErrNotExist`.
* **2b. Destination parent missing:** `ErrNotExist`.
* **2c. A path component is a file:** `ErrNotDirectory` (Windows may report `ErrNotExist`).
* **2d. Permission:** `ErrPermission`.
* **2e. Cross-device:** `ErrInternal` (no copy fallback).
* **POSIX replace:** renaming onto an existing file may replace it (not `ErrExist`).

## Postconditions

* On success, `oldpath` does not exist and `newpath` holds the same inode/content. Errors are sentinels from UC-03.
