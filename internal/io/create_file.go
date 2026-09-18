package io

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
)

// Custom Errors
var (
	ErrorExist        = errors.New("file or directory already exists")
	ErrorNotExist     = errors.New("path does not exist or a component is missing")
	ErrorPermissions  = errors.New("permission denied")
	ErrorNotDirectory = errors.New("a component of the path is a file, not a directory")
	ErrorIsDirectory  = errors.New("the path specified is a directory")
	ErrorNoSpace      = errors.New("no space left on device")
	ErrorReadOnlyFS   = errors.New("file system is read-only")
	ErrorInternal     = errors.New("an unknown internal error occurred")
)

// Constants
const (
	CreateFileFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	CreateFilePerm  = fs.FileMode(0666)
)

// This variable holds the function we want to mock. By default, it's the real os.OpenFile.
var osOpenFile = os.OpenFile

// CreateFile now calls the osOpenFile variable, making it testable.
func CreateFile(path string) (*os.File, error) {
	file, err := osOpenFile(path, CreateFileFlags, CreateFilePerm) // Use the variable here
	if err != nil {
		return nil, mapSystemError(err)
	}
	return file, nil
}

// mapSystemError remains the same as in the previous answer.
func mapSystemError(err error) error {
	switch {
	case errors.Is(err, os.ErrExist):
		return ErrorExist
	case errors.Is(err, os.ErrNotExist):
		return ErrorNotExist
	case errors.Is(err, os.ErrPermission):
		return ErrorPermissions
	case errors.Is(err, syscall.ENOTDIR):
		return ErrorNotDirectory
	case errors.Is(err, syscall.EISDIR):
		return ErrorIsDirectory
	case errors.Is(err, syscall.ENOSPC):
		return ErrorNoSpace
	case errors.Is(err, syscall.EROFS):
		return ErrorReadOnlyFS
	default:
		return ErrorInternal
	}
}
