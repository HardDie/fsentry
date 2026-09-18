# UC-30: Move (rename) an entry

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Rename `<id>.json` and update `id`/`name` in the envelope  
**Preconditions:** source entry exists

## Main scenario (happy path)

1. The caller invokes `MoveEntry[T](oldName, newName, path…)`.
2. The file is renamed, then rewritten (new id/name, `updatedAt` now).

## Alternative scenarios and errors

* **2a. Bad names / missing parent / missing source:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. Destination exists:** `ErrExist`.
* **2c. Rewrite fails:** rename is rolled back.

## Postconditions

* Old ID is gone. `createdAt` is preserved.
