package fs

import "os"

// Stat returns file info for path (follows the last symlink).
func Stat(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, mapError(err)
	}
	return info, nil
}
