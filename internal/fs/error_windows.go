//go:build windows

package fs

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

func mapOS(err error) error {
	// Invented POSIX errnos used by package os on Windows (not Win32 codes).
	switch {
	case errors.Is(err, syscall.EISDIR):
		return wrap(ErrIsDirectory, err)
	case errors.Is(err, syscall.ENOTEMPTY):
		return wrap(ErrExist, err)
	case errors.Is(err, syscall.EBUSY):
		return wrap(ErrBusy, err)
	case errors.Is(err, syscall.EDEADLK):
		return wrap(ErrLock, err)
	}
	errno := windowsErrno(err)
	if errno == 0 {
		return nil
	}
	switch errno {
	case windows.ERROR_ALREADY_EXISTS, windows.ERROR_FILE_EXISTS:
		return wrap(ErrExist, err)
	case windows.ERROR_FILE_NOT_FOUND, windows.ERROR_PATH_NOT_FOUND:
		return wrap(ErrNotExist, err)
	case windows.ERROR_ACCESS_DENIED:
		return wrap(ErrPermission, err)
	case windows.ERROR_DIRECTORY, windows.ERROR_BAD_PATHNAME, windows.ERROR_INVALID_NAME:
		return wrap(ErrNotDirectory, err)
	case windows.ERROR_WRITE_PROTECT:
		return wrap(ErrReadOnly, err)
	case windows.ERROR_DISK_FULL, windows.ERROR_HANDLE_DISK_FULL:
		return wrap(ErrNoSpace, err)
	case windows.ERROR_DIR_NOT_EMPTY:
		return wrap(ErrExist, err)
	case windows.ERROR_LOCK_VIOLATION:
		return wrap(ErrLock, err)
	case windows.ERROR_SHARING_VIOLATION:
		return wrap(ErrBusy, err)
	default:
		return nil
	}
}

func windowsErrno(err error) windows.Errno {
	var se syscall.Errno
	if errors.As(err, &se) {
		return windows.Errno(se)
	}
	var we windows.Errno
	if errors.As(err, &we) {
		return we
	}
	return 0
}
