# UC-12: Remove a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** RemoveEntry, RemoveBinary, rollback after a failed create  
**Goal:** Unlink a file without removing directories  
**Preconditions:** Path is a file. Callers never use `os.Remove` for files.

## Main scenario (happy path)

1. The caller invokes `RemoveFile(path)`.
2. The helper uses `unlink` (not `rmdir`). The file is gone.

## Alternative scenarios and errors

* **2a. Missing:** `ErrNotExist`.
* **2b. Path is a directory:** `ErrIsDirectory` (some OS report `ErrPermission`).
* **2c. Permission:** `ErrPermission`.

## Postconditions

* On success, `path` does not exist. Directories are not deleted. Errors are sentinels from UC-03.
