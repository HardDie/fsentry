package fs

import "os"

// ReadDir lists the names in a directory. It does not recurse.
func ReadDir(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, mapError(err)
	}
	return entries, nil
}
