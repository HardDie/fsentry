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
* **2c. Folder missing:** `ErrNotExist`.
* **2d. Directory without readable `.info.json`:** `ErrFolderCorrupted`.

## Postconditions

* No files are written.
