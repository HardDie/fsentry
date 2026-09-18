# UC-37: Update a binary

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Replace bytes of an existing `<id>.bin`  
**Preconditions:** binary exists

## Main scenario (happy path)

1. The caller invokes `UpdateBinary(name, data, path…)`.
2. The file is truncated and overwritten. `data == nil` leaves an empty file.

## Alternative scenarios and errors

* **2a. Bad name / missing parent / missing file:** `ErrBadName` / `ErrBadPath` / `ErrNotExist`.

## Postconditions

* File name is unchanged. Size matches `len(data)`.
