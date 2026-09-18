//go:build windows

package fs

import (
	"errors"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	// lockRegionOff is the LockFileEx offset so bytes 0–7 stay readable
	// (unix-nano stamp) while the region at offset 8 is locked. Unix flock
	// is whole-file (see lock_unix.go).
	lockRegionOff = 8
	lockBytes     = 1
)

func lockOverlapped() windows.Overlapped {
	return windows.Overlapped{Offset: lockRegionOff}
}

// Lock takes an exclusive advisory lock on file (blocks until available).
func Lock(file *os.File) error {
	ol := lockOverlapped()
	err := windows.LockFileEx(
		windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		lockBytes,
		0,
		&ol,
	)
	if err != nil {
		return mapError(err)
	}
	return nil
}

// TryLock takes an exclusive advisory lock without blocking.
// ErrBusy means another handle holds the lock.
func TryLock(file *os.File) error {
	ol := lockOverlapped()
	err := windows.LockFileEx(
		windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		lockBytes,
		0,
		&ol,
	)
	if err == nil {
		return nil
	}
	if errors.Is(err, syscall.Errno(windows.ERROR_LOCK_VIOLATION)) {
		return ErrBusy
	}
	return mapError(err)
}

// Unlock releases the advisory lock on file.
func Unlock(file *os.File) error {
	ol := lockOverlapped()
	err := windows.UnlockFileEx(windows.Handle(file.Fd()), 0, lockBytes, 0, &ol)
	if err != nil {
		return mapError(err)
	}
	return nil
}
