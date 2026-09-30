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
7. If a folder with the destination id already exists, it is moved aside, the archive is extracted, and the aside copy is removed. Other children of `path` are not modified.
8. The method returns the destination id.

## Alternative scenarios and errors

* **1a. `r == nil` or not a zip:** `ErrBadArchive`.
* **1b. `name` sanitizes to empty:** `ErrBadName`. Nothing is written.
* **2a. Missing destination folder:** `ErrBadPath` / `ErrNotExist`.
* **2b. Non-empty `path` is not a valid folder:** `ErrFolderCorrupted` (see [UC-43](uc-43-ensure-path-folders.md)).
* **5a. Archive is a whole-store export (several roots):** `ErrBadArchive`. Nothing is written.
* **7a. A file occupies the destination id:** `ErrExist`.
* **7b. Extract fails after the previous folder was moved aside:** files this call created are removed and the previous folder is renamed back.

## Postconditions

* On success, `path/<id>` is the archived folder (replaced when that id already existed). Siblings are unchanged. On failure, `path` matches the pre-import tree.
