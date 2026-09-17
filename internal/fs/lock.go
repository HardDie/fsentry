package fs

import "os"

const (
	openLockFlags = os.O_RDWR | os.O_CREATE
	// lockRegionOff is the Windows LockFileEx offset. Unix flock is whole-file.
	// Bytes 0–7 stay readable by waiters while the region at offset 8 is locked.
	lockRegionOff = 8
	lockBytes     = 1
)

// OpenLock opens path for advisory locking, creating it if needed.
// It does not truncate. The caller must Close the handle (after Unlock).
func OpenLock(path string) (*os.File, error) {
	file, err := os.OpenFile(path, openLockFlags, createFilePerm)
	if err != nil {
		return nil, mapError(err)
	}
	return file, nil
}
