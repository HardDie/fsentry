# UC-41: Validate a store tree

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application after import, or after possible manual edits  
**Goal:** Report every on-disk defect under `path` without changing files  
**Preconditions:** `Init` succeeded  

## Main scenario (happy path)

1. The caller invokes `Validate()` or `Validate(path…)`.
2. The store takes a read lock.
3. The directory is walked (root has no `.info.json`; nested directories must). `.fsentry.lock` is skipped.
4. Each folder, entry, and binary is checked: valid ID on disk, readable JSON envelope, envelope `id` / `name` match the path.
5. Stray files (`*.tmp`, unknown names, extra `.info.json` at root) are reported.
6. The caller receives a slice of `Problem`. Empty/nil means the tree is sound.

## Alternative scenarios and errors

* **2a. Missing dest:** `ErrBadPath` / `ErrNotExist` (same as `List`).
* **2b. Non-empty `path` is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)). Root `Validate()` still walks corrupted children.
* **4a. Missing/unreadable `.info.json`:** `ProblemInfo` (walk continues).
* **4b. Bad entry JSON:** `ProblemJSON`.
* **4c. Envelope `id` ≠ disk ID:** `ProblemID`.
* **4d. `NameToID(name)` ≠ disk ID:** `ProblemName`.
* **4e. Disk name is not an ID:** `ProblemBadName`.
* **5a. `*.tmp`:** `ProblemTemp`. Other files: `ProblemUnexpected`.

## Postconditions

* Disk is unchanged. Validation does not repair.
