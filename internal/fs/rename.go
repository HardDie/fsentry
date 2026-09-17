package fs

import "os"

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
	if err != nil {
		return mapError(err)
	}
	return nil
}
