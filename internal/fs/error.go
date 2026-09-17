package fs

import (
	"errors"
	"os"
)

// Sentinels returned by create helpers. Higher layers must match these with
// errors.Is; they must not inspect syscall.Errno.
var (
	ErrExist        = errors.New("already exists")
	ErrNotExist     = errors.New("does not exist")
	ErrPermission   = errors.New("permission denied")
	ErrNotDirectory = errors.New("not a directory")
	ErrIsDirectory  = errors.New("is a directory")
	ErrNoSpace      = errors.New("no space left on device")
	ErrReadOnly     = errors.New("read-only file system")
	ErrInternal     = errors.New("internal error")
)

func wrap(sentinel error, _ error) error {
	return sentinel
}

// mapError translates an OS error into one of the sentinels above.
// The OS error is classified and dropped so the return stays allocation-free.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, os.ErrExist):
		return wrap(ErrExist, err)
	case errors.Is(err, os.ErrNotExist):
		return wrap(ErrNotExist, err)
	case errors.Is(err, os.ErrPermission):
		return wrap(ErrPermission, err)
	}
	if mapped := mapOS(err); mapped != nil {
		return mapped
	}
	return wrap(ErrInternal, err)
}
