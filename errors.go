package fsentry

import (
	"errors"

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
	ErrPermission      = fs.ErrPermission
	ErrNoSpace         = fs.ErrNoSpace
	ErrReadOnly        = fs.ErrReadOnly
	ErrLock            = fs.ErrLock
	ErrBusy            = fs.ErrBusy
	ErrInternal        = fs.ErrInternal
)
