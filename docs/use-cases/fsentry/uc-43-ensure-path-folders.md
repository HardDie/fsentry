# UC-43: Require valid folders on a nested path

**Module:** `fsentry`  
**Status:** Done  
**Actors:** any public method with a parent `path` (`CreateFolder`, `GetEntry`, `List`, `Export`, …)  
**Goal:** Do not operate inside a folder whose ancestors are missing, not directories, or corrupted  
**Preconditions:** `Init` succeeded

## Main scenario (happy path)

1. The caller passes a parent chain such as `CreateFolder("Cards", data, "My Game", "My Collection")`.
2. The store walks from the root. The root is a directory (it has no `.info.json`).
3. Each path segment is a directory whose `.info.json` is readable and whose envelope `id` and `NameToID(name)` match the disk ID.
4. The requested operation runs in that directory.

## Alternative scenarios and errors

* **2a. Empty / `.` / `..` / separator in a segment, or path leaves the root:** `ErrBadPath`.
* **2b. A segment name is not a valid ID:** `ErrBadName`.
* **2c. A segment directory is missing or a parent is missing:** `ErrBadPath`.
* **2d. A segment exists but is not a directory:** `ErrNotDirectory`.
* **2e. A segment directory has missing or unreadable `.info.json`, or `id` / `name` do not match the disk ID:** `ErrFolderCorrupted`.
* **2f. Empty `path` (store root):** only the root directory is required; no `.info.json`.

## Postconditions

* Nested access never treats a corrupted or mismatched ancestor as a normal parent. `List()` of the root still reports corrupted *children* in `CorruptedFolder`. `Validate()` of the root still walks them.
