# UC-25: Move (rename) a folder

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Rename the directory and update `id`/`name` in `.info.json`  
**Preconditions:** source folder exists

## Main scenario (happy path)

1. The caller invokes `MoveFolder(oldName, newName, path…)`.
2. The directory is renamed, then `.info.json` is rewritten (new id/name, `updatedAt` now).
3. The caller receives `FolderInfo[T]` (`MoveFolder[T]`).

## Alternative scenarios and errors

* **2a. Bad names:** `ErrBadName`.
* **2b. Parent missing:** `ErrBadPath`.
* **2c. Source missing:** `ErrNotExist`.
* **2d. Destination exists:** `ErrExist`.
* **2e. Info write fails:** rename is rolled back.

## Postconditions

* Old ID is gone. `createdAt` is preserved. For a rename that also keeps `updatedAt`, see [UC-46](uc-46-update-folder-name-without-timestamp.md).
