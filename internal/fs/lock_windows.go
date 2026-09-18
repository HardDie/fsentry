//go:build windows

package fs

import (
	"os"
	"runtime"

	"golang.org/x/sys/windows"
)

const (
	// lockRegionOff is the LockFileEx offset so bytes 0–7 stay readable
	// (unix-nano stamp) while the region at offset 8 is locked. Unix flock
	// is whole-file (see lock_unix.go).
	lockRegionOff = 8
	lockBytes     = 1
)

// OpenLock opens path for advisory locking, creating it if needed.
// FILE_SHARE_DELETE lets a waiter unlink the file to steal a stale lock
// (Unix unlink-while-open). The caller must Close the handle (after Unlock).
func OpenLock(path string) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, mapError(err)
	}
	h, err := windows.CreateFile(
		p,
		windows.GENERIC_READ|windows.GENERIC_WRITE|windows.DELETE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, mapError(err)
	}
	return os.NewFile(uintptr(h), path), nil
}

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
	runtime.KeepAlive(file)
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
	runtime.KeepAlive(file)
	if err == nil {
		return nil
	}
	switch windowsErrno(err) {
	case windows.ERROR_LOCK_VIOLATION, windows.ERROR_IO_PENDING, windows.ERROR_SHARING_VIOLATION:
		return ErrBusy
	}
	return mapError(err)
}

// Unlock releases the advisory lock on file.
func Unlock(file *os.File) error {
	ol := lockOverlapped()
	err := windows.UnlockFileEx(windows.Handle(file.Fd()), 0, lockBytes, 0, &ol)
	runtime.KeepAlive(file)
	if err == nil || windowsErrno(err) == windows.ERROR_NOT_LOCKED {
		return nil
	}
	return mapError(err)
}
