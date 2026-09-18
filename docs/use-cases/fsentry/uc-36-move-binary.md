# UC-36: Move (rename) a binary

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Rename `<id>.bin`  
**Preconditions:** source exists

## Main scenario (happy path)

1. The caller invokes `MoveBinary(oldName, newName, path…)`.
2. The file is renamed. Contents are unchanged.

## Alternative scenarios and errors

* **2a. Bad names / missing parent / missing source:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. Destination exists:** `ErrExist`.

## Postconditions

* Old ID is gone. Bytes are at the new ID.
