# UC-49: Readers share the lock file

**Module:** `internal/lock`  
**Status:** Done  
**Actors:** `*DB` read methods (`Get*`, `List`, `Export*`, `Validate`)  
**Goal:** Reads do not wait for each other; a write still has the store alone  
**Preconditions:** Lock file on. One or more `Open` handles on `<root>/.fsentry.lock`.

## Main scenario (happy path)

1. Process A `RLock`s: OS shared lock, then an 8-byte stamp (no fsync).
2. Process B `RLock`s while A holds it: it gets the shared lock at once.
3. Goroutines on A's handle `RLock` too: they join A's hold without an OS call.
4. Each `RUnlock` drops one reader. The last one on a handle releases the OS lock.

## Alternative scenarios and errors

* **2a. A writer holds the lock:** `RLock` polls `TryLockShared` every 100ms until it `Unlock`s.
* **2b. Readers hold the lock, a writer arrives:** `Lock` polls until every reader has released.
* **2c. The holder's stamp is older than the timeout:** the waiter steals, as in [UC-19](uc-19-steal-stale-lock.md).
* **2d. The last writer's stamp is old, but a reader holds the lock now:** the reader wrote a fresh stamp, so the writer waits instead of stealing.
* **2e. `RUnlock` with no reader:** no-op.

## Postconditions

* Any number of readers may hold the lock together, in any number of processes.
* A writer never overlaps a reader.
