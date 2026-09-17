package fs

import (
	"os"
	"syscall"
)

// RemoveFile unlinks a file. It does not remove directories (ErrIsDirectory).
func RemoveFile(path string) error {
	err := syscall.Unlink(path)
	if err != nil {
		return mapError(err)
	}
	return nil
}

// RemoveFolder recursively deletes a directory and its contents.
// Missing path is ErrNotExist. A file at path is ErrNotDirectory.
func RemoveFolder(path string) error {
	info, err := Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return ErrNotDirectory
	}
	err = os.RemoveAll(path)
	if err != nil {
		return mapError(err)
	}
	return nil
}
