# UC-06: Rename a folder

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Higher fsentry layers (`MoveFolder`)  
**Goal:** Rename or move a directory (with its contents) in one syscall, or fail with a known sentinel  
**Preconditions:** Source directory exists. Destination parent exists. Callers never use `os.Rename` for folders.

## Main scenario (happy path)

1. The caller invokes `RenameFolder(oldpath, newpath)`.
2. The helper calls `os.Rename`. It does not copy the tree and does not create missing parents.
3. The directory and its children are available only at `newpath`.

## Alternative scenarios and errors

* **2a. Source missing:** `ErrNotExist`.
* **2b. Destination parent missing:** `ErrNotExist`.
* **2c. Destination is a non-empty directory:** `ErrExist` (`ENOTEMPTY` / `ERROR_DIR_NOT_EMPTY`).
* **2d. A path component is a file:** `ErrNotDirectory` (Windows may report `ErrNotExist`).
* **2e. Permission:** `ErrPermission`.
* **2f. Cross-device:** `ErrInternal` (no copy fallback).

## Postconditions

* On success, `oldpath` does not exist and `newpath` is the same directory tree. Errors are sentinels from UC-03.
