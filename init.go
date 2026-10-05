package fsentry

import (
	"errors"
	"path/filepath"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/lock"
)

// Init creates the store root if it is missing and opens the lock file when locking is on.
// An existing directory is reused. Init is idempotent.
func (db *DB) Init() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.root == "" {
		return ErrBadPath
	}
	info, err := fs.Stat(db.root)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := fs.CreateFolderAll(db.root); err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return ErrNotDirectory
	}
	return db.openLockLocked()
}

// Drop unlocks, closes the lock file, and removes the store root.
func (db *DB) Drop() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.root == "" {
		return ErrBadPath
	}
	if db.lk != nil {
		_ = db.lk.Unlock()
		_ = db.lk.Close()
		db.lk = nil
	}
	err := fs.RemoveFolder(db.root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (db *DB) openLockLocked() error {
	if db.noLock || db.lk != nil {
		return nil
	}
	lk, err := lock.Open(filepath.Join(db.root, lock.FileName), db.lockTimeout)
	if err != nil {
		return err
	}
	db.lk = lk
	return nil
}

func (db *DB) withLock(write bool, fn func() error) error {
	if write {
		db.mu.Lock()
		defer db.mu.Unlock()
	} else {
		db.mu.RLock()
		defer db.mu.RUnlock()
	}
	if err := db.takeLock(write); err != nil {
		return err
	}
	defer db.releaseLock(write)
	return fn()
}

// takeLock takes the lock file: exclusive for a write, shared for a read.
func (db *DB) takeLock(write bool) error {
	if db.noLock || db.lk == nil {
		return nil
	}
	if write {
		return db.lk.Lock()
	}
	return db.lk.RLock()
}

func (db *DB) releaseLock(write bool) {
	if db.noLock || db.lk == nil {
		return
	}
	var err error
	if write {
		err = db.lk.Unlock()
	} else {
		err = db.lk.RUnlock()
	}
	if err != nil {
		db.log.Error("unlock failed", "root", db.root, "write", write, "err", err)
	}
}
