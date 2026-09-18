# UC-31: Update entry payload

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Replace `data` without renaming  
**Preconditions:** entry exists

## Main scenario (happy path)

1. The caller invokes `UpdateEntry(name, data, path…)`.
2. The file is rewritten via temp + rename. `updatedAt` is now.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing file:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.

## Postconditions

* File name is unchanged.
