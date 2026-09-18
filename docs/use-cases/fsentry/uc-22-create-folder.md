# UC-22: Create a folder

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Create an ID directory and `.info.json` under a parent path  
**Preconditions:** `Init` succeeded

## Main scenario (happy path)

1. The caller invokes `CreateFolder(name, data, path…)`. `T` is inferred from `data` (`CreateFolder[any](name, nil)` for JSON `null`).
2. `NameToID(name)` is the directory name. Parent `path` is resolved from the root.
3. `Mkdir` then exclusive create of `.info.json` (QuotedString name, UTC timestamps, JSON `data`).
4. The caller receives `FolderInfo[T]` by value.

## Alternative scenarios and errors

* **2a. Bad name:** `ErrBadName`.
* **2b. Parent missing / not a directory / escapes root:** `ErrBadPath`.
* **2c. Directory exists:** `ErrExist`.
* **2d. Info create fails:** the new directory is removed.

## Postconditions

* A folder object exists. Same name and ID (`"My Game"` / `"my_game"`) collide.
