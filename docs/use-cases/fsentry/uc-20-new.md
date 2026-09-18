# UC-20: Construct a store handle

**Module:** `fsentry`  
**Status:** Done  
**Actors:** application or test  
**Goal:** Obtain a `*DB` for a root path without touching the disk  
**Preconditions:** none

## Main scenario (happy path)

1. The caller invokes `New(root, opts…)`.
2. `root` is stored (`filepath.Clean` unless it is empty). Options apply in order (`WithPretty`, `WithLogger`, `WithNoLockFile` / `WithLockFile`, `WithLockTimeout`).
3. The caller receives a non-nil `*DB`. The directory is not created. The lock file is not opened.

## Alternative scenarios and errors

* **2a. Empty `root`:** stored as empty; later `Init` fails with `ErrBadPath` (not implemented in this scenario).
* **2b. Nil option in the slice:** skipped.
* **2c. `WithLogger(nil)`:** ignored; a previous logger stays.
* **2d. `WithLockTimeout(d)` with `d <= 0`:** default timeout (10 minutes at `Init`) stays.

## Postconditions

* `New` never returns an error or nil. Disk and lock are unchanged. Lock is **on** unless `WithNoLockFile` was last. Pretty JSON is off unless `WithPretty`.
