//go:build unix

package fs

import (
	"os"
	"syscall"
)

const openLockFlags = os.O_RDWR | os.O_CREATE

// OpenLock opens path for advisory locking, creating it if needed.
// It does not truncate. The caller must Close the handle (after Unlock).
func OpenLock(path string) (*os.File, error) {
	file, err := os.OpenFile(path, openLockFlags, createFilePerm)
	if err != nil {
		return nil, mapError(err)
	}
	return file, nil
}

// Lock takes an exclusive advisory lock on file (blocks until available).
func Lock(file *os.File) error {
	return flock(file, syscall.LOCK_EX)
}

// TryLock takes an exclusive advisory lock without blocking.
// ErrBusy means another handle holds the lock.
func TryLock(file *os.File) error {
	return flock(file, syscall.LOCK_EX|syscall.LOCK_NB)
}

// Unlock releases the advisory lock on file.
func Unlock(file *os.File) error {
	return flock(file, syscall.LOCK_UN)
}

func flock(file *os.File, how int) error {
	fd := int(file.Fd())
	for {
		err := syscall.Flock(fd, how)
		if err == syscall.EINTR {
			continue
		}
		if err == nil {
			return nil
		}
		if how&syscall.LOCK_NB != 0 && (err == syscall.EAGAIN || err == syscall.EWOULDBLOCK) {
			return ErrBusy
		}
		return mapError(err)
	}
}
