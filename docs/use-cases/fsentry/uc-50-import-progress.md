# UC-50: Follow import progress

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application importing a large archive (a DeckBuilder game) behind a progress bar  
**Goal:** Learn how many files and bytes are written while the import runs  
**Preconditions:** Same as [UC-40](uc-40-import.md) or [UC-48](uc-48-import-folder.md)

## Main scenario (happy path)

1. The caller invokes `ImportWithProgress(r, progress, path…)` or `ImportFolderWithProgress(r, name, progress, path…)`.
2. The archive is checked as in UC-40 / UC-48.
3. Before the first file, `progress` gets `Files` 0 and the totals.
4. After each file is written and synced, `progress` gets the new counts.
5. The last call has `Files == FilesTotal` and `Bytes == BytesTotal`.

## Alternative scenarios and errors

* **1a. `progress` is nil:** same as `Import` / `ImportFolder`.
* **2a. Archive refused (`ErrExist`, `ErrBadArchive`, bad path):** `progress` is never called.
* **4a. A file fails mid-way:** rollback as in UC-40; no further `progress` call.

## Postconditions

* `progress` was called `FilesTotal + 1` times on success.
* It ran under the write lock; it did not call the same `*DB`.
