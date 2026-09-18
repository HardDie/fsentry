package storage

import (
	"errors"
	"io"
	iofs "io/fs"
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/otiai10/copy"

	fsio "github.com/HardDie/fsentry/internal/io"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	CreateDirPerm   = 0755
	UpdateFileFlags = os.O_WRONLY | os.O_TRUNC
	CreateFilePerm  = 0666
)

type FS struct{}

func New() FS {
	return FS{}
}

// CreateFile allows you to create a file and fill it with some binary data.
func (r FS) CreateFile(path string, data []byte) error {
	file, err := fsio.CreateFile(path)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("CreateFile(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("CreateFile(): error close file %q: %s", path, err.Error())
		}
	}()

	n, err := file.Write(data)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	if n != len(data) {
		log.Printf("CreateFile(): the size of input and written data is different. Received: %d, written: %d", len(data), n)
	}
	return nil
}

// UpdateFile allows you to update a file.
func (r FS) UpdateFile(path string, data []byte) error {
	file, err := os.OpenFile(path, UpdateFileFlags, CreateFilePerm)
	if err != nil {
		if e := isKnownError(err); e != nil {
			switch {
			case errors.Is(e, fsentry_error.ErrorNotFile):
				return fsentry_error.Wrap(err, fsentry_error.ErrorNotExist)
			case errors.Is(e, fsentry_error.ErrorIsDirectory):
				return fsentry_error.Wrap(err, fsentry_error.ErrorNotExist)
			}
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		if err = file.Sync(); err != nil {
			log.Printf("UpdateFile(): error sync file %q: %s", path, err.Error())
		}
		if err = file.Close(); err != nil {
			log.Printf("UpdateFile(): error close file %q: %s", path, err.Error())
		}
	}()

	n, err := file.Write(data)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	if n != len(data) {
		log.Printf("UpdateFile(): the size of input and written data is different. Received: %d, written: %d", len(data), n)
	}
	return nil
}

// ReadFile attempts to open and read all binary data from the desired file.
func (r FS) ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return nil, e
		}
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer func() {
		var err error
		if err = file.Close(); err != nil {
			log.Printf("ReadFile(): error close file %q: %s", path, err.Error())
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		if e := isKnownError(err); e != nil {
			switch {
			case errors.Is(e, fsentry_error.ErrorNotFile):
				return nil, fsentry_error.Wrap(err, fsentry_error.ErrorNotExist)
			case errors.Is(e, fsentry_error.ErrorIncorrectFunction):
				return nil, fsentry_error.Wrap(err, fsentry_error.ErrorNotExist)
			}
			return nil, e
		}
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return data, nil
}

// RemoveFile allows you to delete a file or an empty folder.
// If the folder is not empty, an ErrorExist error will be returned.
func (r FS) RemoveFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// CreateFolder allows you to create a folder in the file system.
func (r FS) CreateFolder(path string) error {
	err := os.Mkdir(path, CreateDirPerm)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// CreateAllFolder allows you to create a folder in the file system,
// and if some intermediate folders in the desired path do not exist, they will also be created.
// If the specified folder already exists, there will be no error.
func (r FS) CreateAllFolder(path string) error {
	err := os.MkdirAll(path, CreateDirPerm)
	if err != nil {
		if e := isKnownError(err); e != nil {
			switch {
			case errors.Is(e, fsentry_error.ErrorNotDirectory):
				return fsentry_error.Wrap(err, fsentry_error.ErrorExist)
			case errors.Is(e, fsentry_error.ErrorNotExist): // windows do not treat files and folders in the same way
				return fsentry_error.Wrap(err, fsentry_error.ErrorExist)
			}
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// RemoveFolder will delete the desired folder even if it is not empty with all the data it contains.
func (r FS) RemoveFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		if e := isKnownError(err); e != nil {
			return e
		}
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// Rename allows you to rename a file/directory or move it to another path.
func (r FS) Rename(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// CopyFolder will recursively copy the source folder to the desired destination path.
func (r FS) CopyFolder(srcPath, dstPath string) error {
	err := copy.Copy(srcPath, dstPath)
	if err != nil {
		// TODO: process different types of errors
		return fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return nil
}

// List will read the complete list of objects on the specified path and return them.
func (r FS) List(path string) ([]os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		// TODO: process different types of errors
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		// TODO: process different types of errors
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return files, nil
}

// IsFileExist checks if an object that is a file, not a folder, exists at the specified path.
func (r FS) IsFileExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// entry not exist
			return false, nil
		}
		// other error
		return false, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}

	// check if it is not a folder
	if stat.IsDir() {
		return false, fsentry_error.ErrorBadPath
	}

	// entry exists
	return true, nil
}

// IsFolderExist checks if an object that is a folder, exists at the specified path.
func (r FS) IsFolderExist(path string) (isExist bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// folder not exist
			return false, nil
		}
		// other error
		return false, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}

	// check if it is a folder
	if !stat.IsDir() {
		return false, fsentry_error.ErrorBadPath
	}

	// folder exists
	return true, nil
}

func isKnownError(err error) error {
	var pathErr *iofs.PathError
	if errors.As(err, &pathErr) {
		switch {
		case os.IsExist(pathErr):
			return fsentry_error.Wrap(err, fsentry_error.ErrorExist)
		case os.IsNotExist(pathErr):
			return fsentry_error.Wrap(err, fsentry_error.ErrorNotExist)
		case os.IsPermission(pathErr):
			return fsentry_error.Wrap(err, fsentry_error.ErrorPermissions)
		}
	}
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		switch uintptr(syscallErr) {
		case 20:
			return fsentry_error.Wrap(err, fsentry_error.ErrorNotDirectory)
		case 21:
			return fsentry_error.Wrap(err, fsentry_error.ErrorNotFile)
		}
		if runtime.GOOS == "windows" {
			switch uintptr(syscallErr) {
			case 1:
				return fsentry_error.Wrap(err, fsentry_error.ErrorIncorrectFunction)
			case 536870954:
				return fsentry_error.Wrap(err, fsentry_error.ErrorIsDirectory)
			}
		}
	}
	return nil
}
