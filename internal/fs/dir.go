package fs

import (
	"errors"
	"os"
	"path/filepath"
)

const createDirPerm = os.FileMode(0755)

// CreateFolder creates a single directory. It does not create missing parents
// (unlike os.MkdirAll).
func CreateFolder(path string) error {
	err := os.Mkdir(path, createDirPerm)
	if err != nil {
		return mapError(err)
	}
	return nil
}

// CreateFolderAll creates path and any missing parents (os.MkdirAll).
func CreateFolderAll(path string) error {
	err := os.MkdirAll(path, createDirPerm)
	if err != nil {
		return mapError(err)
	}
	return nil
}

// RenameFile renames or moves a file. It does not copy. Destination parents
// must already exist.
func RenameFile(oldpath, newpath string) error {
	return rename(oldpath, newpath)
}

// RenameFolder renames or moves a directory. It does not copy. Destination
// parents must already exist.
func RenameFolder(oldpath, newpath string) error {
	return rename(oldpath, newpath)
}

func rename(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err == nil {
		return nil
	}
	parent := filepath.Dir(newpath)
	if parent != newpath {
		info, st := os.Stat(parent)
		if st == nil && !info.IsDir() {
			return ErrNotDirectory
		}
	}
	mapped := mapError(err)
	info, st := os.Stat(newpath)
	if st == nil && info.IsDir() && !errors.Is(mapped, ErrNotExist) {
		return ErrExist
	}
	return mapped
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

// ReadDir lists the names in a directory. It does not recurse.
func ReadDir(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err == nil {
		return entries, nil
	}
	info, st := os.Stat(path)
	if st == nil && !info.IsDir() {
		return nil, ErrNotDirectory
	}
	return nil, mapError(err)
}

// Stat returns file info for path (follows the last symlink).
func Stat(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, mapError(err)
	}
	return info, nil
}
