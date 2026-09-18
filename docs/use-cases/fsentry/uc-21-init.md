# UC-21: Init the store root

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application after `New`  
**Goal:** Ensure the root directory exists and the lock file is opened when locking is on  
**Preconditions:** `*DB` from `New`

## Main scenario (happy path)

1. The caller invokes `Init()`.
2. Missing root is created (`CreateFolderAll`). An existing directory is reused.
3. If the lock file is enabled, `internal/lock.Open` creates/opens `<root>/.fsentry.lock`.

## Alternative scenarios and errors

* **2a. Empty root:** `ErrBadPath`.
* **2b. Root is a file:** `ErrNotDirectory`.
* **2c. Second `Init`:** success (idempotent).
* **2d. `WithNoLockFile`:** skip opening the lock.

## Postconditions

* Root exists as a directory. `Drop` deletes it.
