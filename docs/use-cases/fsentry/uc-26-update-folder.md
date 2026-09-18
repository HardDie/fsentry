# UC-26: Update folder payload

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Replace `data` in `.info.json` without renaming  
**Preconditions:** folder exists

## Main scenario (happy path)

1. The caller invokes `UpdateFolder(name, data, path…)`. `T` is inferred from `data`.
2. `.info.json` is rewritten via temp + rename. `updatedAt` is now; `createdAt` is unchanged.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing folder:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. Unreadable info:** `ErrFolderCorrupted`.

## Postconditions

* Directory name is unchanged.
