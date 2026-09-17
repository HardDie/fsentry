// Package lock is the inter-process lock file used by *DB.
//
// Open the path, then Lock before a store operation and Unlock after.
// Lock writes a unix-nano stamp. If another process still holds the OS lock
// but the stamp (or file mtime if there is no stamp) is older than Timeout,
// Lock unlinks the file and takes a new inode so a crashed holder cannot
// block the store forever. A live holder past Timeout is treated the same
// (split brain). Default Timeout is 10 minutes; pass a positive duration to
// Open, or 0 to use the default. *DB Init will expose this as WithLockTimeout.
package lock

import (
	"errors"
	"os"
	"time"

	"github.com/HardDie/fsentry/internal/fs"
)

const (
	// DefaultTimeout is how long a holder may keep the lock before a waiter
	// steals it. *DB Init may pass a different duration into Open.
	DefaultTimeout = 10 * time.Minute

	// FileName is the lock file under the store root.
	FileName = ".fsentry.lock"

	pollInterval = 100 * time.Millisecond
)

// File is an open lock file handle plus steal timeout.
type File struct {
	path    string
	file    *os.File
	timeout time.Duration
	now     func() time.Time
}

// Open opens (or creates) the lock file at path.
// timeout <= 0 means DefaultTimeout.
func Open(path string, timeout time.Duration) (*File, error) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	file, err := fs.OpenLock(path)
	if err != nil {
		return nil, err
	}
	return &File{
		path:    path,
		file:    file,
		timeout: timeout,
		now:     time.Now,
	}, nil
}

// Lock takes the exclusive lock, stealing if the stamp is older than timeout.
func (l *File) Lock() error {
	for {
		err := fs.TryLock(l.file)
		switch {
		case err == nil:
			return l.writeStamp()
		case !errors.Is(err, fs.ErrBusy):
			return err
		}
		stale, err := l.stale()
		if err != nil {
			return err
		}
		if stale {
			if err := l.steal(); err != nil {
				return err
			}
			continue
		}
		time.Sleep(pollInterval)
	}
}

// Unlock releases the OS lock. The stamp is left for the next holder to overwrite.
func (l *File) Unlock() error {
	if l.file == nil {
		return nil
	}
	return fs.Unlock(l.file)
}

// Close closes the handle. The OS drops the lock if Unlock was skipped.
func (l *File) Close() error {
	if l.file == nil {
		return nil
	}
	err := fs.Close(l.file)
	l.file = nil
	return err
}

func (l *File) writeStamp() error {
	return fs.WriteLockStamp(l.file, l.now().UnixNano())
}

func (l *File) stale() (bool, error) {
	ns, ok, err := fs.ReadLockStamp(l.file)
	if err != nil {
		return false, err
	}
	var held time.Time
	if ok {
		held = time.Unix(0, ns)
	} else {
		info, err := fs.Stat(l.path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return true, nil
			}
			return false, err
		}
		held = info.ModTime()
	}
	return !held.After(l.now().Add(-l.timeout)), nil
}

func (l *File) steal() error {
	if l.file != nil {
		_ = fs.Close(l.file)
		l.file = nil
	}
	err := fs.RemoveFile(l.path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		file, openErr := fs.OpenLock(l.path)
		if openErr != nil {
			return openErr
		}
		l.file = file
		return err
	}
	file, err := fs.OpenLock(l.path)
	if err != nil {
		return err
	}
	l.file = file
	return nil
}
