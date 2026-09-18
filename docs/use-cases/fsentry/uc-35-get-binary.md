# UC-35: Get a binary

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Read `<id>.bin` into a caller buffer  
**Preconditions:** binary exists

## Main scenario (happy path)

1. The caller invokes `GetBinary(name, buf, path…)`.
2. Bytes are read into `buf` (grown if short). `buf == nil` allocates.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing file:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. Path is a directory:** `ErrNotFile`.

## Postconditions

* Returned slice holds the file bytes. A sized `buf` is the low-alloc path.
