# UC-03: Map an OS error to a sentinel

**Module:** `internal/fs`  
**Status:** Done  
**Actors:** `CreateFile` and `CreateFolder` (later: read/update/remove wrappers)  
**Goal:** Every OS failure becomes exactly one known sentinel so the rest of the library never inspects `syscall.Errno` or guesses Windows vs Unix codes  
**Preconditions:** A syscall returned a non-nil error (or the helper received `nil`)

## Main scenario (happy path)

1. A create helper gets a non-nil error from `os.OpenFile` or `os.Mkdir`.
2. `mapError` matches portable `os.ErrExist` / `os.ErrNotExist` / `os.ErrPermission`, then OS-specific errno (`error_unix.go` or `error_windows.go`).
3. It returns the package sentinel only (`wrap` does not `Join` or format the OS error). That keeps **zero allocations**. Callers `errors.Is` the sentinel; the original errno is not retained.
4. If `err` is `nil`, `mapError` returns `nil` with **zero allocations**.

Rename of a non-empty destination directory maps `ENOTEMPTY` / `ERROR_DIR_NOT_EMPTY` to `ErrExist`.

## Alternative scenarios and errors

Sentinels (complete list for this package):

| Sentinel | Typical cause |
|---|---|
| `ErrExist` | already exists |
| `ErrNotExist` | missing path component |
| `ErrPermission` | EACCES, EPERM, ACCESS_DENIED |
| `ErrNotDirectory` | ENOTDIR, ERROR_DIRECTORY |
| `ErrIsDirectory` | EISDIR |
| `ErrNoSpace` | ENOSPC, EDQUOT, ERROR_DISK_FULL |
| `ErrReadOnly` | EROFS, ERROR_WRITE_PROTECT |
| `ErrInternal` | anything not in this table |

* **2a. Unknown errno:** `ErrInternal`. The original error is not attached.
* **2b. Windows PATH_NOT_FOUND when a parent is a file:** maps to `ErrNotExist` (what the OS reported), not a guessed `ErrNotDirectory`.

## Postconditions

* Callers above `internal/fs` branch only with `errors.Is` on the table above. They do not import `syscall` for create failures. `wrap` and `mapError` allocate nothing.
