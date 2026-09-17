# UC-02: Create a folder

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Higher fsentry layers (`CreateFolder` on `*DB`, `Init`)  
**Goal:** Create a single directory at an exact path, or fail with a sentinel from the known list  
**Preconditions:** The parent directory exists and is writable. Callers never use `os.Mkdir` / `os.MkdirAll` for this.

## Main scenario (happy path)

1. The caller invokes `CreateFolder(path)`.
2. The helper calls `os.Mkdir` with mode `0755` (umask applied). It does not create missing parents (`MkdirAll` is not used).
3. The directory exists and is empty.

## Alternative scenarios and errors

* **2a. Directory (or a file at that name) already exists:** `ErrExist`.
* **2b. Parent directory missing:** `ErrNotExist`.
* **2c. A path component is a file:** `ErrNotDirectory` (Windows may report `ErrNotExist`).
* **2d. Parent not writable:** `ErrPermission`.
* **2e. Disk full or quota:** `ErrNoSpace`.
* **2f. Read-only filesystem:** `ErrReadOnly`.
* **2g. Anything else:** `ErrInternal`. The original OS error is wrapped; callers still `errors.Is` the sentinel only.

## Postconditions

* On success, `path` is a directory. Intermediate directories are not created. On failure, the error is one of the sentinels in UC-03.
