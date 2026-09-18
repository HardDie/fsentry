# UC-28: Create an entry

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application  
**Goal:** Create `<id>.json` with the standard envelope  
**Preconditions:** `Init` succeeded

## Main scenario (happy path)

1. The caller invokes `CreateEntry(name, data, path…)`. `T` is inferred from `data`.
2. The file is created with `O_EXCL` at `parent/id.json`.
3. The caller receives `Entry[T]`.

## Alternative scenarios and errors

* **2a. Bad name:** `ErrBadName`.
* **2b. Parent missing:** `ErrBadPath`.
* **2c. File exists:** `ErrExist`.

## Postconditions

* `List` includes the ID in `Entries`. Folder of the same name can coexist (`dir` vs `dir.json`).
