# UC-29: Get an entry

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Read `<id>.json` and unmarshal `data` as `T`  
**Preconditions:** entry exists

## Main scenario (happy path)

1. The caller invokes `GetEntry[T](name, path…)`.
2. Name and ID address the same file.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing file:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. Path is a directory:** `ErrNotFile`.
* **2c. Envelope JSON is invalid:** `ErrInternal`.
* **2d. Payload does not match `T`:** `ErrInternal`.

## Postconditions

* No files are written.
