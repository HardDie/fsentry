package fs

import "os"

const (
	openReadFlags  = os.O_RDONLY
	openWriteFlags = os.O_WRONLY | os.O_TRUNC
)

// OpenRead opens an existing file for reading. It does not create the file.
func OpenRead(path string) (*os.File, error) {
	file, err := os.OpenFile(path, openReadFlags, 0)
	if err != nil {
		return nil, mapError(err)
	}
	return file, nil
}

// OpenWrite opens an existing file for writing and truncates it.
// It does not create a missing file (ErrNotExist).
func OpenWrite(path string) (*os.File, error) {
	file, err := os.OpenFile(path, openWriteFlags, 0)
	if err != nil {
		return nil, mapError(err)
	}
	return file, nil
}
