# UC-48: Import one folder into an existing store

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application restoring one hierarchy node (a DeckBuilder game) into a store that already has other folders  
**Goal:** Extract a single-folder zip under `path`, optionally under a new name, without touching siblings  
**Preconditions:** `Init` succeeded; destination directory exists; the zip has exactly one top-level folder whose name is already a valid id  

## Main scenario (happy path)

1. The caller invokes `ImportFolder(r, name, path…)`. Empty `path` is the store root. Empty `name` keeps the archive's root id.
2. The store takes a write lock.
3. The zip is opened (`ReaderAt`+`Size` when available, otherwise `io.ReadAll`).
4. Names are normalized to `/`, cleaned, and rejected if they would leave the destination (`ErrBadArchive`). `.fsentry.lock` entries are skipped.
5. Every entry must sit under the same top-level folder, that folder name must satisfy `NameToID(name) == name`, and at least one file must live under it. A file at the zip root, a second top-level folder, or a root that is not an id is `ErrBadArchive`.
6. When `name` is non-empty, every zip path is rewritten from the archive id to `NameToID(name)` before any disk write. The folder `.info.json` `id` and `name` are updated; `createdAt` and `updatedAt` stay as in the archive.
7. If a folder or file with the destination id already exists, the call returns `ErrExist` before any write. Other children of `path` are not modified.
8. On success, the method returns the destination id.

## Alternative scenarios and errors

* **1a. `r == nil` or not a zip:** `ErrBadArchive`.
* **1b. `name` sanitizes to empty:** `ErrBadName`. Nothing is written.
* **2a. Missing destination folder:** `ErrBadPath` / `ErrNotExist`.
* **2b. Non-empty `path` is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **5a. Archive is a whole-store export (several roots):** `ErrBadArchive`. Nothing is written.
* **7a. Destination id already exists (folder or file):** `ErrExist`. The existing tree is unchanged.
* **7b. Extract fails after some creates:** those creates are removed (rollback).

## Postconditions

* On success, `path/<id>` is the archived folder. Siblings are unchanged. On failure, including `ErrExist`, `path` matches the pre-import tree.
