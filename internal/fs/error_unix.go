//go:build unix

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
	case syscall.EEXIST:
		return wrap(ErrExist, err)
	case syscall.ENOENT:
		return wrap(ErrNotExist, err)
	case syscall.EACCES, syscall.EPERM:
		return wrap(ErrPermission, err)
	case syscall.ENOTDIR:
		return wrap(ErrNotDirectory, err)
	case syscall.EISDIR:
		return wrap(ErrIsDirectory, err)
	case syscall.ENOSPC, syscall.EDQUOT:
		return wrap(ErrNoSpace, err)
	case syscall.EROFS:
		return wrap(ErrReadOnly, err)
	case syscall.ENOTEMPTY:
		return wrap(ErrExist, err)
	case syscall.EDEADLK:
		return wrap(ErrLock, err)
	default:
		return nil
	}
}
