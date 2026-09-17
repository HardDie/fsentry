# UC-15: Stat a path

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** List (file vs dir), RemoveFolder, existence checks  
**Goal:** Return `os.FileInfo` for a path (follows the last symlink)  
**Preconditions:** Callers never use `os.Stat` / `os.Lstat`.

## Main scenario (happy path)

1. The caller invokes `Stat(path)`.
2. The helper returns size, mode, and `IsDir`.

## Alternative scenarios and errors

* **2a. Missing:** `ErrNotExist`.
* **2b. Permission:** `ErrPermission`.
* **2c. Parent component is a file:** `ErrNotDirectory` (Windows may report `ErrNotExist`).

## Postconditions

* On success, `FileInfo` describes the path. Errors are sentinels from UC-03.
