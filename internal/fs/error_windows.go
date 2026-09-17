//go:build windows

package fs

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

func mapOS(err error) error {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return nil
	}
	switch windows.Errno(errno) {
	case windows.ERROR_ALREADY_EXISTS, windows.ERROR_FILE_EXISTS:
		return wrap(ErrExist, err)
	case windows.ERROR_FILE_NOT_FOUND, windows.ERROR_PATH_NOT_FOUND:
		return wrap(ErrNotExist, err)
	case windows.ERROR_ACCESS_DENIED:
		return wrap(ErrPermission, err)
	case windows.ERROR_DIRECTORY:
		return wrap(ErrNotDirectory, err)
	case windows.ERROR_WRITE_PROTECT:
		return wrap(ErrReadOnly, err)
	case windows.ERROR_DISK_FULL, windows.ERROR_HANDLE_DISK_FULL:
		return wrap(ErrNoSpace, err)
	case windows.ERROR_DIR_NOT_EMPTY:
		return wrap(ErrExist, err)
	case windows.ERROR_LOCK_VIOLATION:
		return wrap(ErrLock, err)
	default:
		return nil
	}
}
