//go:build windows

package fs

import (
	"errors"
	"syscall"
)

func mapOS(err error) error {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return nil
	}
	switch errno {
	case syscall.ERROR_ALREADY_EXISTS, syscall.ERROR_FILE_EXISTS:
		return wrap(ErrExist, err)
	case syscall.ERROR_FILE_NOT_FOUND, syscall.ERROR_PATH_NOT_FOUND:
		return wrap(ErrNotExist, err)
	case syscall.ERROR_ACCESS_DENIED:
		return wrap(ErrPermission, err)
	case syscall.ERROR_DIRECTORY:
		return wrap(ErrNotDirectory, err)
	case syscall.ERROR_WRITE_PROTECT:
		return wrap(ErrReadOnly, err)
	case syscall.ERROR_DISK_FULL, syscall.ERROR_HANDLE_DISK_FULL:
		return wrap(ErrNoSpace, err)
	case syscall.ERROR_DIR_NOT_EMPTY:
		return wrap(ErrExist, err)
	default:
		return nil
	}
}
