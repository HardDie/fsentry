package fs

import (
	"io/fs"
	"os"
)

const (
	createFileFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	createFilePerm  = fs.FileMode(0666)
	createDirPerm   = fs.FileMode(0755)
)

// CreateFile creates a new file with O_EXCL. The caller must close the handle.
// It does not write bytes or create missing parents.
func CreateFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, createFileFlags, createFilePerm)
	if err != nil {
		return nil, mapError(err)
	}
	return file, nil
}

// CreateFolder creates a single directory. It does not create missing parents
// (unlike os.MkdirAll).
func CreateFolder(path string) error {
	err := os.Mkdir(path, createDirPerm)
	if err != nil {
		return mapError(err)
	}
	return nil
}
