# UC-24: List a directory

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Return IDs of folders, entries, binaries, and corrupted folders  
**Preconditions:** `Init` succeeded

## Main scenario (happy path)

1. The caller invokes `List(path…)`. Empty path is the root.
2. Child directories with readable `.info.json` are `Folders`. Directories without are `CorruptedFolder`.
3. `*.json` except `.info.json` → `Entries` (suffix stripped). `*.bin` → `Binaries`. `.fsentry.lock` is skipped.

## Alternative scenarios and errors

* **2a. Invalid path:** `ErrBadPath` / `ErrBadName`.
* **2b. Missing path:** `ErrBadPath` if `path` is non-empty; `ErrNotExist` if the root is missing.

## Postconditions

* List values are IDs, not display names.
