package fsentry

import (
	"errors"
	"strings"

	"github.com/HardDie/fsentry/internal/fs"
)

// Sentinels for errors.Is. OS errors are classified in internal/fs and dropped.
var (
	ErrBadName         = errors.New("bad name")
	ErrBadPath         = errors.New("bad path")
	ErrExist           = fs.ErrExist
	ErrNotExist        = fs.ErrNotExist
	ErrNotFile         = errors.New("not a file")
	ErrNotDirectory    = fs.ErrNotDirectory
	ErrIsDirectory     = fs.ErrIsDirectory
	ErrFolderCorrupted = errors.New("folder corrupted")
	ErrBadArchive      = errors.New("bad zip archive")
	ErrPermission      = fs.ErrPermission
	ErrNoSpace         = fs.ErrNoSpace
	ErrReadOnly        = fs.ErrReadOnly
	ErrLock            = fs.ErrLock
	ErrBusy            = fs.ErrBusy
	ErrInternal        = fs.ErrInternal
)

// BadPathError is a missing parent segment from ensurePath.
// errors.Is(err, ErrBadPath) is true.
// Path is the prefix ensurePath was checking, from the first segment through
// the missing one, using the caller's path strings (not on-disk IDs).
// Other bad paths (empty root, ".", "..", a separator in a segment) stay the
// bare ErrBadPath sentinel.
type BadPathError struct {
	Path []string
}

// Error returns "bad path" and the failed parent, joined with "/".
func (e *BadPathError) Error() string {
	if e == nil || len(e.Path) == 0 {
		return ErrBadPath.Error()
	}
	return ErrBadPath.Error() + ": " + strings.Join(e.Path, "/")
}

// Unwrap returns ErrBadPath.
func (e *BadPathError) Unwrap() error {
	return ErrBadPath
}

func missingParent(path []string) error {
	cp := make([]string, len(path))
	copy(cp, path)
	return &BadPathError{Path: cp}
}
