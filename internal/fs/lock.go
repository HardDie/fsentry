package fs

import (
	"encoding/binary"
	"io"
	"os"
)

const (
	openLockFlags = os.O_RDWR | os.O_CREATE
	lockStampSize = 8
	// lockRegionOff is the Windows LockFileEx offset. Unix flock is whole-file.
	// Stamp bytes 0–7 stay readable by waiters while the region at offset 8 is locked.
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

// WriteLockStamp writes unixNano as 8 big-endian bytes at offset 0 and syncs.
func WriteLockStamp(file *os.File, unixNano int64) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return mapError(err)
	}
	var buf [lockStampSize]byte
	binary.BigEndian.PutUint64(buf[:], uint64(unixNano))
	if err := Write(file, buf[:]); err != nil {
		return err
	}
	if err := file.Truncate(lockStampSize); err != nil {
		return mapError(err)
	}
	return Sync(file)
}

// ReadLockStamp reads the 8-byte stamp at offset 0.
// ok is false when the file is shorter than 8 bytes (no stamp yet).
func ReadLockStamp(file *os.File) (unixNano int64, ok bool, err error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return 0, false, mapError(err)
	}
	var buf [lockStampSize]byte
	n, err := file.Read(buf[:])
	if n == lockStampSize {
		return int64(binary.BigEndian.Uint64(buf[:])), true, nil
	}
	if err != nil && err != io.EOF {
		return 0, false, mapError(err)
	}
	return 0, false, nil
}
