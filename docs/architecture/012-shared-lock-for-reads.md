# 12. Shared lock file for reads

* **Status:** Accepted
* **Date:** 2026-10-04
* **Authors:** @oleg

---

## Context

1. ADR 007 took the exclusive lock for every public method, reads included.
2. Every acquire wrote the stamp, then truncated and fsynced the file.
3. On macOS fsync is a full flush: about 3 ms per call.
4. A read of a 340 KB binary costs about 0.1 ms; the lock was 30× that.
5. DeckBuilder makes thousands of reads per render (about 5,000 for 1,000 cards).
6. Reads from two processes also waited for each other.
7. Goroutines on one `*DB` share one lock FD under `RLock`.
   1. Each called the exclusive `Lock` / `Unlock` on that FD.
   2. The first reader to finish released the OS lock for the others.
   3. A steal could swap the FD under another reader.

## Considered options

1. Keep the exclusive lock and drop only the fsync.
2. Shared OS lock for reads, exclusive for writes.
3. No lock file for reads.

## Decision

Use option 2.

1. Reads take `LOCK_SH` (Windows: `LockFileEx` without the exclusive flag).
2. Writes keep `LOCK_EX`, the stamp, truncate, and fsync.
3. `internal/fs` adds `TryLockShared`.
4. `internal/lock.File` adds `RLock` / `RUnlock`.
   1. A mutex guards the FD and a reader count.
   2. The first reader takes the OS lock; the last one releases it.
   3. Goroutines on one `File` count as one holder.
5. A reader writes the stamp too, without truncate or fsync.
   1. Waiters read it through the page cache.
   2. After a crash the OS lock is gone, so the stamp need not reach the disk.
   3. Without it, a writer would see the last writer's old stamp and steal.
6. Steal rules are the same for readers and writers.
7. `withLock(false)` takes `RLock`; `withLock(true)` takes `Lock`.

## Consequences

### Positive

1. A read lock costs about 4 µs instead of about 3 ms.
2. Reads in different processes run at the same time.
3. Reader goroutines on one `*DB` no longer release each other's lock.

### Negative and risks

1. Readers that keep overlapping can starve a writer in another process.
2. Only the first reader of a group stamps.
   1. A group that overlaps longer than the timeout can be stolen from.
3. Two 8-byte stamp writes from two readers can race.
   1. Both values are fresh, so the steal check stays correct.

### Neutral

1. Writes cost the same as before.
2. The on-disk format does not change.
