package fs

import "os"

// Close closes file. The caller must not use file afterward.
func Close(file *os.File) error {
	err := file.Close()
	if err != nil {
		return mapError(err)
	}
	return nil
}

// Sync flushes file data and metadata to stable storage.
func Sync(file *os.File) error {
	err := file.Sync()
	if err != nil {
		return mapError(err)
	}
	return nil
}
