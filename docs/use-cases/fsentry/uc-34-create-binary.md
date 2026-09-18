# UC-34: Create a binary

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Create `<id>.bin` with raw bytes (no envelope)  
**Preconditions:** `Init` succeeded

## Main scenario (happy path)

1. The caller invokes `CreateBinary(name, data, path…)`.
2. The file is created with `O_EXCL`. `data == nil` is an empty file.

## Alternative scenarios and errors

* **2a. Bad name:** `ErrBadName`.
* **2b. Parent missing:** `ErrBadPath`.
* **2c. Parent exists but is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **2d. File exists:** `ErrExist`.

## Postconditions

* `List.Binaries` includes the ID. Bytes on disk are exactly `data`.
