# UC-47: Export one folder and its children

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application sharing one hierarchy node (a DeckBuilder game)  
**Goal:** Write a zip of one folder and everything under it, without the rest of the store  
**Preconditions:** `Init` succeeded; the folder exists and is valid; destination writer is writable  

## Main scenario (happy path)

1. The caller invokes `ExportFolder(w, name, path…)`. `name` is the folder. `path` is the parent chain (empty path is the store root).
2. The store takes a read lock (mutex + lock file unless `WithNoLockFile`).
3. The folder directory is walked. `.fsentry.lock` is skipped. Files are stored with `/` names, prefixed by that folder's ID (`my_game/.info.json`).
4. The zip writer is closed so the archive is complete.

## Alternative scenarios and errors

* **1a. `w == nil`:** `ErrInternal`.
* **2a. Missing parent:** `ErrBadPath` (same as `GetFolder`).
* **2b. `name` sanitizes to empty:** `ErrBadName`.
* **2c. Folder missing:** `ErrNotExist`.
* **2d. Folder directory is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **3a. Walk / read fails:** the corresponding OS sentinel.

## Postconditions

* Disk is unchanged. The writer contains a zip that `ImportFolder` can read. Siblings of the folder are absent from the archive.
