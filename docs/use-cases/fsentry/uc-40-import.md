# UC-40: Import a zip archive into a store

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application restoring or receiving a tree  
**Goal:** Extract a zip into the directory at `path` (empty path is the store root) without leaving the root  
**Preconditions:** `Init` succeeded; destination directory exists  

## Main scenario (happy path)

1. The caller invokes `Import(r)` or `Import(r, path…)`.
2. The store takes a write lock.
3. The zip is opened (`ReaderAt`+`Size` when available, otherwise `io.ReadAll`).
4. Names are normalized to `/`, cleaned, and rejected if they would leave the destination (`ErrBadArchive`). `.fsentry.lock` entries are skipped.
5. Existing files at extract paths fail with `ErrExist` before any write.
6. Files and missing directories are created. JSON and binaries land as in the archive.

## Alternative scenarios and errors

* **1a. `r == nil` or not a zip:** `ErrBadArchive`.
* **2a. Missing destination folder:** `ErrBadPath` / `ErrNotExist`.
* **2b. Non-empty `path` is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **5a. Target path is a directory where a file is expected:** `ErrIsDirectory`.
* **6a. Extract fails after some creates:** those creates are removed (rollback).

## Postconditions

* On success, dest contains the archive tree. On failure, dest matches the pre-import files for names this call created.
