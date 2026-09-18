# UC-44: Duplicate a folder

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Deep-copy a folder tree to a new name with a fresh destination envelope  
**Preconditions:** source folder exists and is valid; destination ID does not exist

## Main scenario (happy path)

1. The caller invokes `DuplicateFolder[T](srcName, dstName, path…)`.
2. The store copies the source directory recursively (nested folders, entries, binaries). Nested envelopes keep their ids, names, and timestamps.
3. Destination `.info.json` is rewritten with the new id/name and `createdAt`/`updatedAt` set to now. Payload `data` is copied.
4. The caller receives `FolderInfo[T]`.

## Alternative scenarios and errors

* **2a. Bad names:** `ErrBadName`.
* **2b. Parent missing / invalid:** `ErrBadPath` / `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **2c. Source missing:** `ErrNotExist`.
* **2d. Source `.info.json` unreadable or id/name mismatch:** `ErrFolderCorrupted`.
* **2e. Destination exists:** `ErrExist`.
* **2f. Copy or info write fails:** the destination directory is removed.

## Postconditions

* Source is unchanged. Destination is a separate tree. Same name and ID (`"My Game"` / `"my_game"`) collide with an existing folder.
