# UC-32: Duplicate an entry

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Copy an entry to a new name with fresh timestamps  
**Preconditions:** source exists, destination does not

## Main scenario (happy path)

1. The caller invokes `DuplicateEntry[T](srcName, dstName, path…)`.
2. A new file is created (`O_EXCL`) with copied `data`, new id/name, `createdAt`/`updatedAt` now.

## Alternative scenarios and errors

* **2a. Source missing:** `ErrNotExist`.
* **2b. Destination exists:** `ErrExist`.
* **2c. Bad names / missing parent:** `ErrBadName` / `ErrBadPath`.

## Postconditions

* Source is unchanged. Destination is a separate object.
