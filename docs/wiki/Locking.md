# Locking

Two layers serialize access to a store. Tests usually turn the second layer off — see [Testing](Testing).

## 1. In-process mutex

Each `*DB` has a `sync.RWMutex`. Writes take `Lock`; `Get*` and `List` take `RLock`. A shared lock-file descriptor does not serialize goroutines, so this layer is always on.

## 2. Inter-process lock file

Production default: exclusive advisory lock on `<root>/.fsentry.lock` for **every** public method (reads included). Unix uses `flock`; Windows uses `LockFileEx`. `Init` creates the file. `List` ignores it. `Drop` unlocks, closes, then removes the root.

On acquire, the locker writes an 8-byte unix-nano stamp. Waiters poll `TryLock` (about 100ms). If the stamp (or mtime if the stamp is missing) is older than the timeout, the waiter **steals**: unlinks the file, opens a new inode, locks that. Default timeout is **10 minutes**. A crash usually drops the OS lock immediately; steal covers a hung process or a filesystem that keeps the lock.

A live holder past the timeout is treated the same (possible split brain). Keep the timeout long enough for real work, or pass `WithLockTimeout`.

```go
db := fsentry.New("data", fsentry.WithLockTimeout(15*time.Minute))
if err := db.Init(); err != nil {
	return err
}
```

`WithNoLockFile()` skips layer 2 only. Production callers omit it.

## Order

Mutex first, then flock; reverse on the way out.

Two `*DB` handles on the same root in one process still serialize through the lock file if locking is on. Two goroutines sharing **one** `*DB` serialize through the mutex.

## Failures

Cannot open or lock: [`ErrLock`](Errors). Unexpected unlock/sync problems may go to `WithLogger` (no payload bytes).
