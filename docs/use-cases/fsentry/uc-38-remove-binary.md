# UC-38: Remove a binary

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Unlink `<id>.bin`  
**Preconditions:** binary exists

## Main scenario (happy path)

1. The caller invokes `RemoveBinary(name, path…)`.
2. The file is unlinked.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing file:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. Path is a directory:** `ErrNotFile`.

## Postconditions

* The ID is gone from `List.Binaries`.
