package fsentry

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/HardDie/fsentry/internal/fs"
)

const binarySuffix = ".bin"

// CreateBinary creates `<id>.bin` with the given bytes (no metadata file).
func (db *DB) CreateBinary(name string, data []byte, path ...string) error {
	return db.withLock(true, func() error {
		file, _, err := db.binaryPath(name, path...)
		if err != nil {
			return err
		}
		return db.writeBytesCreate(file, data)
	})
}

// GetBinary reads `<id>.bin` into buf. A sized buf with cap >= file size is the
// zero-extra-alloc path. buf == nil is allowed and allocates.
func (db *DB) GetBinary(name string, buf []byte, path ...string) ([]byte, error) {
	var out []byte
	err := db.withLock(false, func() error {
		file, _, err := db.binaryPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(file); err != nil {
			return err
		}
		f, err := fs.OpenRead(file)
		if err != nil {
			return err
		}
		got, err := fs.Read(f, buf)
		closeErr := fs.Close(f)
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		out = got
		return nil
	})
	return out, err
}

// MoveBinary renames `<id>.bin`.
func (db *DB) MoveBinary(oldName, newName string, path ...string) error {
	return db.withLock(true, func() error {
		oldFile, _, err := db.binaryPath(oldName, path...)
		if err != nil {
			return err
		}
		newFile, _, err := db.binaryPath(newName, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(oldFile); err != nil {
			return err
		}
		switch err := db.statFile(newFile); {
		case err == nil:
			return ErrExist
		case errors.Is(err, ErrNotExist):
		default:
			return err
		}
		return fs.RenameFile(oldFile, newFile)
	})
}

// UpdateBinary truncates and overwrites an existing `<id>.bin`.
func (db *DB) UpdateBinary(name string, data []byte, path ...string) error {
	return db.withLock(true, func() error {
		file, _, err := db.binaryPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(file); err != nil {
			return err
		}
		return db.writeBytesReplace(file, data)
	})
}

// RemoveBinary unlinks `<id>.bin`.
func (db *DB) RemoveBinary(name string, path ...string) error {
	return db.withLock(true, func() error {
		file, _, err := db.binaryPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(file); err != nil {
			return err
		}
		return fs.RemoveFile(file)
	})
}

func (db *DB) binaryPath(name string, path ...string) (file, id string, err error) {
	parent, err := db.resolve(path...)
	if err != nil {
		return "", "", err
	}
	if err := db.statDir(parent, true); err != nil {
		return "", "", err
	}
	id, err = db.objectID(name)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(parent, id+binarySuffix), id, nil
}

func (db *DB) writeBytesCreate(path string, data []byte) error {
	file, err := fs.CreateFile(path)
	if err != nil {
		return err
	}
	return db.writeBytesClose(file, data)
}

func (db *DB) writeBytesReplace(path string, data []byte) error {
	file, err := fs.OpenWrite(path)
	if err != nil {
		return err
	}
	return db.writeBytesClose(file, data)
}

func (db *DB) writeBytesClose(file *os.File, data []byte) error {
	if err := fs.Write(file, data); err != nil {
		_ = fs.Close(file)
		return err
	}
	if err := fs.Sync(file); err != nil {
		_ = fs.Close(file)
		if db.log != nil {
			db.log.Error("sync failed")
		}
		return err
	}
	return fs.Close(file)
}
