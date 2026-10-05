# UC-37: Update a binary

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Replace bytes of an existing `<id>.bin`  
**Preconditions:** binary exists

## Main scenario (happy path)

1. The caller invokes `UpdateBinary(name, data, path…)`.
2. The bytes go to `<id>.bin.tmp`, are fsynced, and the temp file is renamed over `<id>.bin`. `data == nil` leaves an empty file.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing file:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.
* **2b. The temp file cannot be written, synced, or renamed:** the error is returned and `<id>.bin` keeps its old bytes.

## Postconditions

* File name is unchanged. Size matches `len(data)`.
* After a crash the binary holds its old bytes or its new ones, never part of them.
