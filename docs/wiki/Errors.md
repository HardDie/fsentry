# Errors

Public methods return package sentinels. Use `errors.Is`. OS errors are classified and dropped; do not match `os.ErrNotExist` as the only check. Related: [getting started](Getting-Started), [paths and IDs](Paths-and-IDs).

```go
import (
	"errors"

	"github.com/HardDie/fsentry"
)

_, err := db.CreateFolder("My Notes", struct{}{})
if errors.Is(err, fsentry.ErrExist) {
	// already there
}
```

| Sentinel | Typical cause |
|---|---|
| `ErrBadName` | Name sanitizes to empty or a Windows reserved device (`con`, `prn`, …) |
| `ErrBadPath` | Empty root on `Init`, missing parent, `.` / `..`, path would leave the root |
| `ErrExist` | Create or move target already exists |
| `ErrNotExist` | Get / update / remove / move source missing |
| `ErrNotFile` | Expected a file, found something else |
| `ErrNotDirectory` | Expected a directory (including `Init` when root is a file) |
| `ErrIsDirectory` | Expected a file, found a directory |
| `ErrFolderCorrupted` | Folder directory exists but `.info.json` is missing or unreadable |
| `ErrBadArchive` | Invalid zip, or an entry would extract outside the destination — see [Export and import](Export-and-Import) |
| `ErrPermission` | OS permission denied |
| `ErrNoSpace` | Disk full |
| `ErrReadOnly` | Read-only filesystem |
| `ErrLock` | Lock file could not be opened or locked — see [Locking](Locking) |
| `ErrBusy` | Internal: `TryLock` while another holder has the lock (steal path) |
| `ErrInternal` | Unexpected OS or JSON failure |

Methods do not panic on missing files. Zero `Entry[T]` / `FolderInfo[T]` plus a non-nil error means the operation failed.

## Example: get or create

```go
note, err := db.GetFolder[Notebook]("My Notes")
if errors.Is(err, fsentry.ErrNotExist) {
	note, err = db.CreateFolder("My Notes", Notebook{Color: "blue"})
}
if err != nil {
	return err
}
```
