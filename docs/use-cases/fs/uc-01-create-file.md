# UC-01: Create a file

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** Higher fsentry layers (`CreateEntry`, `CreateBinary`, folder `.info.json`)  
**Goal:** Create a new empty file at an exact path, or fail with a sentinel from the known list  
**Preconditions:** The parent directory exists and is writable. Callers never use `os.OpenFile` for create.

## Main scenario (happy path)

1. The caller invokes `CreateFile(path)`.
2. The helper opens the path with `O_WRONLY|O_CREATE|O_EXCL` and mode `0666` (umask applied).
3. It returns an open write-only `*os.File`. The file length is 0. The caller closes the handle.
4. Missing parents are not created. No bytes are written.

## Alternative scenarios and errors

* **2a. File already exists:** `ErrExist`.
* **2b. Parent directory missing:** `ErrNotExist`.
* **2c. A path component is a file:** `ErrNotDirectory` (Windows may report `ErrNotExist`).
* **2d. `path` is an existing directory:** `ErrIsDirectory` (Windows may report `ErrExist` or `ErrPermission`).
* **2e. Directory not writable:** `ErrPermission`.
* **2f. Disk full or quota:** `ErrNoSpace`.
* **2g. Read-only filesystem:** `ErrReadOnly`.
* **2h. Anything else:** `ErrInternal`. The original OS error is wrapped; callers still `errors.Is` the sentinel only.

## Postconditions

* On success, `path` exists as an empty regular file. On failure, no new file is left and the error is one of the sentinels in UC-03.
