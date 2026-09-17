package fs

import (
	"io"
	"os"
	"syscall"
)

const (
	createFileFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	createFilePerm  = os.FileMode(0666)
	openReadFlags   = os.O_RDONLY
	openWriteFlags  = os.O_WRONLY | os.O_TRUNC
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

// Write writes all of data to file. It does not create, seek, sync, or close.
func Write(file *os.File, data []byte) error {
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			return mapError(err)
		}
		data = data[n:]
	}
	return nil
}

// WriteAt writes all of data at off. It does not seek, sync, or close.
func WriteAt(file *os.File, data []byte, off int64) error {
	for len(data) > 0 {
		n, err := file.WriteAt(data, off)
		if err != nil {
			return mapError(err)
		}
		data = data[n:]
		off += int64(n)
	}
	return nil
}

// Read reads from the current offset until EOF into buf.
// If buf is too small, it is grown with append. A sized buf with cap >= size
// is the zero-extra-alloc path. io.EOF is success, not an error.
func Read(file *os.File, buf []byte) ([]byte, error) {
	out := buf[:0]
	for {
		if cap(out) == len(out) {
			n := 64
			if c := cap(out); c > 0 {
				n = c * 2
			}
			grown := make([]byte, len(out), n)
			copy(grown, out)
			out = grown
		}
		n, err := file.Read(out[len(out):cap(out)])
		out = out[:len(out)+n]
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return out, mapError(err)
		}
	}
}

// ReadAt reads from off into buf. A short read (including EOF) is success
// with n < len(buf).
func ReadAt(file *os.File, buf []byte, off int64) (int, error) {
	n, err := file.ReadAt(buf, off)
	if err == io.EOF {
		return n, nil
	}
	if err != nil {
		return n, mapError(err)
	}
	return n, nil
}

// Truncate changes file's size. It does not seek, sync, or close.
func Truncate(file *os.File, size int64) error {
	if err := file.Truncate(size); err != nil {
		return mapError(err)
	}
	return nil
}

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

// RemoveFile unlinks a file. It does not remove directories (ErrIsDirectory).
func RemoveFile(path string) error {
	err := syscall.Unlink(path)
	if err != nil {
		return mapError(err)
	}
	return nil
}
