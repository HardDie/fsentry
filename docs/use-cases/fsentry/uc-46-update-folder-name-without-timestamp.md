# UC-46: Rename a folder without bumping updatedAt

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application (DeckBuilder import/rename)  
**Goal:** Rename the directory and `id`/`name` in `.info.json` while keeping `createdAt` and `updatedAt`  
**Preconditions:** source folder exists

## Main scenario (happy path)

1. The caller invokes `UpdateFolderNameWithoutTimestamp[T](oldName, newName, path…)`.
2. The directory is renamed like `MoveFolder`.
3. `.info.json` is rewritten with the new id/name. Timestamps are copied from the source envelope.
4. The caller receives `FolderInfo[T]`.

## Alternative scenarios and errors

* **2a. Bad names:** `ErrBadName`.
* **2b. Parent missing / invalid:** `ErrBadPath` / `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **2c. Source missing:** `ErrNotExist`.
* **2d. Destination exists:** `ErrExist`.
* **2e. Info write fails:** rename is rolled back.

## Postconditions

* Old ID is gone. `createdAt` and `updatedAt` match the source folder. Children stay inside the moved tree.
