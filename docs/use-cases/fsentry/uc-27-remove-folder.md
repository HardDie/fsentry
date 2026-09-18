# UC-27: Remove a folder

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Recursively delete a folder object  
**Preconditions:** folder exists with readable `.info.json`

## Main scenario (happy path)

1. The caller invokes `RemoveFolder(name, path…)`.
2. After verifying `.info.json`, `RemoveAll` deletes the directory.

## Alternative scenarios and errors

* **2a. Bad name / missing parent:** `ErrBadName` / `ErrBadPath`.
* **2b. Missing folder:** `ErrNotExist`.
* **2c. Directory without `.info.json`:** `ErrFolderCorrupted` (not deleted).

## Postconditions

* The ID is gone from `List`. Children are deleted with the folder.
