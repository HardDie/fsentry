# UC-39: Export a store (or folder) as a zip archive

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application backing up or sharing a tree  
**Goal:** Write a zip of the directory at `path` (empty path is the store root)  
**Preconditions:** `Init` succeeded; destination writer is writable  

## Main scenario (happy path)

1. The caller invokes `Export(w)` or `Export(w, path…)`.
2. The store takes a read lock (mutex + lock file unless `WithNoLockFile`).
3. The directory is walked. `.fsentry.lock` is skipped. Files are stored with `/` names.
4. If `path` is non-empty, every zip name is prefixed with that folder's ID (`my_notes/.info.json`).
5. The zip writer is closed so the archive is complete.

## Alternative scenarios and errors

* **2a. Missing parent / missing folder:** `ErrBadPath` or `ErrNotExist` (same as `List`).
* **2b. `w == nil`:** `ErrInternal`.
* **3a. Walk / read fails:** the corresponding OS sentinel.

## Postconditions

* Disk is unchanged. The writer contains a zip that `Import` can read.
