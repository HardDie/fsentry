package fsentry

import (
	"errors"
	"path/filepath"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/lock"
)

// CreateFolder creates a folder named name under path (empty path is the store root).
func (db *DB) CreateFolder[T any](name string, data T, path ...string) (FolderInfo[T], error) {
	var out FolderInfo[T]
	err := db.withLock(true, func() error {
		parent, err := db.ensurePath(path...)
		if err != nil {
			return err
		}
		id, err := db.objectID(name)
		if err != nil {
			return err
		}
		dir := filepath.Join(parent, id)
		payload, err := marshalPayload(data)
		if err != nil {
			return err
		}
		now := db.clock()
		disk := envelope{
			ID:        id,
			Name:      QuotedString(name),
			CreatedAt: now,
			UpdatedAt: now,
			Data:      payload,
		}
		if err := fs.CreateFolder(dir); err != nil {
			return err
		}
		if err := db.writeInfoCreate(dir, disk); err != nil {
			if rmErr := fs.RemoveFolder(dir); rmErr != nil && db.log != nil {
				db.log.Error("remove folder after info create failed")
			}
			return err
		}
		out = FolderInfo[T]{
			ID:        id,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
			Data:      data,
		}
		return nil
	})
	return out, err
}

// GetFolder reads `.info.json` for name under path and unmarshals data as T.
func (db *DB) GetFolder[T any](name string, path ...string) (FolderInfo[T], error) {
	var out FolderInfo[T]
	err := db.withLock(false, func() error {
		dir, id, err := db.folderPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, false); err != nil {
			if errors.Is(err, ErrNotExist) {
				return ErrNotExist
			}
			if errors.Is(err, ErrNotDirectory) {
				return ErrNotDirectory
			}
			return err
		}
		disk, err := db.readValidInfo(dir, id)
		if err != nil {
			return err
		}
		out, err = decodeFolder[T](disk)
		return err
	})
	return out, err
}

// MoveFolder renames a folder on disk and updates id/name in `.info.json`.
// updatedAt is set to now. T is the payload type in `.info.json`.
func (db *DB) MoveFolder[T any](oldName, newName string, path ...string) (FolderInfo[T], error) {
	var out FolderInfo[T]
	err := db.withLock(true, func() error {
		parent, err := db.ensurePath(path...)
		if err != nil {
			return err
		}
		oldID, err := db.objectID(oldName)
		if err != nil {
			return err
		}
		newID, err := db.objectID(newName)
		if err != nil {
			return err
		}
		oldDir := filepath.Join(parent, oldID)
		newDir := filepath.Join(parent, newID)
		if err := db.statDir(oldDir, false); err != nil {
			return err
		}
		switch err := db.statDir(newDir, false); {
		case err == nil:
			return ErrExist
		case errors.Is(err, ErrNotExist):
		default:
			return err
		}
		disk, err := db.readValidInfo(oldDir, oldID)
		if err != nil {
			return err
		}
		if err := fs.RenameFolder(oldDir, newDir); err != nil {
			return err
		}
		now := db.clock()
		disk.ID = newID
		disk.Name = QuotedString(newName)
		disk.UpdatedAt = now
		if err := db.writeInfoReplace(newDir, disk); err != nil {
			_ = fs.RenameFolder(newDir, oldDir)
			return err
		}
		out, err = decodeFolder[T](disk)
		return err
	})
	return out, err
}

// UpdateFolder replaces the folder payload and bumps updatedAt.
func (db *DB) UpdateFolder[T any](name string, data T, path ...string) (FolderInfo[T], error) {
	var out FolderInfo[T]
	err := db.withLock(true, func() error {
		dir, id, err := db.folderPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, false); err != nil {
			return err
		}
		disk, err := db.readValidInfo(dir, id)
		if err != nil {
			return err
		}
		payload, err := marshalPayload(data)
		if err != nil {
			return err
		}
		disk.Data = payload
		now := db.clock()
		disk.UpdatedAt = now
		if err := db.writeInfoReplace(dir, disk); err != nil {
			return err
		}
		out = FolderInfo[T]{
			ID:        disk.ID,
			Name:      disk.Name.String(),
			CreatedAt: disk.CreatedAt.UTC(),
			UpdatedAt: now,
			Data:      data,
		}
		return nil
	})
	return out, err
}

// DuplicateFolder copies a folder tree to a new name. Nested objects keep their
// ids, names, and timestamps. The destination `.info.json` gets the new id/name
// and fresh createdAt/updatedAt.
func (db *DB) DuplicateFolder[T any](srcName, dstName string, path ...string) (FolderInfo[T], error) {
	var out FolderInfo[T]
	err := db.withLock(true, func() error {
		parent, err := db.ensurePath(path...)
		if err != nil {
			return err
		}
		srcID, err := db.objectID(srcName)
		if err != nil {
			return err
		}
		dstID, err := db.objectID(dstName)
		if err != nil {
			return err
		}
		srcDir := filepath.Join(parent, srcID)
		dstDir := filepath.Join(parent, dstID)
		if err := db.statDir(srcDir, false); err != nil {
			return err
		}
		switch err := db.statDir(dstDir, false); {
		case err == nil:
			return ErrExist
		case errors.Is(err, ErrNotExist):
		default:
			return err
		}
		disk, err := db.readValidInfo(srcDir, srcID)
		if err != nil {
			return err
		}
		if err := db.copyDir(srcDir, dstDir); err != nil {
			_ = fs.RemoveFolder(dstDir)
			return err
		}
		now := db.clock()
		disk.ID = dstID
		disk.Name = QuotedString(dstName)
		disk.CreatedAt = now
		disk.UpdatedAt = now
		if err := db.writeInfoReplace(dstDir, disk); err != nil {
			_ = fs.RemoveFolder(dstDir)
			return err
		}
		out, err = decodeFolder[T](disk)
		return err
	})
	return out, err
}

func (db *DB) copyDir(src, dst string) error {
	if err := fs.CreateFolder(dst); err != nil {
		return err
	}
	entries, err := fs.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == lock.FileName {
			continue
		}
		from := filepath.Join(src, name)
		to := filepath.Join(dst, name)
		info, err := fs.Stat(from)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := db.copyDir(from, to); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(from, to); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := fs.OpenRead(src)
	if err != nil {
		return err
	}
	defer func() { _ = fs.Close(in) }()
	out, err := fs.CreateFile(dst)
	if err != nil {
		return err
	}
	buf := make([]byte, zipCopyChunk)
	var off int64
	for {
		n, err := fs.ReadAt(in, buf, off)
		if err != nil {
			_ = fs.Close(out)
			return err
		}
		if n == 0 {
			break
		}
		if err := fs.Write(out, buf[:n]); err != nil {
			_ = fs.Close(out)
			return err
		}
		off += int64(n)
		if n < len(buf) {
			break
		}
	}
	if err := fs.Sync(out); err != nil {
		_ = fs.Close(out)
		return err
	}
	return fs.Close(out)
}

// RemoveFolder deletes the folder and its contents. Missing `.info.json` is ErrFolderCorrupted.
func (db *DB) RemoveFolder(name string, path ...string) error {
	return db.withLock(true, func() error {
		dir, id, err := db.folderPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, false); err != nil {
			return err
		}
		if _, err := db.readValidInfo(dir, id); err != nil {
			return err
		}
		return fs.RemoveFolder(dir)
	})
}

func (db *DB) folderPath(name string, path ...string) (dir, id string, err error) {
	parent, err := db.ensurePath(path...)
	if err != nil {
		return "", "", err
	}
	id, err = db.objectID(name)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(parent, id), id, nil
}
