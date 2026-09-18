# UC-23: Get a folder

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Read `.info.json` for a folder  
**Preconditions:** `Init` succeeded

## Main scenario (happy path)

1. The caller invokes `GetFolder[T](name, path…)`.
2. Name and ID address the same directory.
3. Envelope fields are returned as `FolderInfo[T]` (`QuotedString` decoded, `data` unmarshaled as `T`).

## Alternative scenarios and errors

* **2a. Bad name:** `ErrBadName`.
* **2b. Parent missing:** `ErrBadPath`.
* **2c. Parent exists but is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **2d. Folder missing:** `ErrNotExist`.
* **2e. Directory without readable `.info.json` or id/name mismatch:** `ErrFolderCorrupted`.

## Postconditions

* No files are written.
