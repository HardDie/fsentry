package fsentry

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/jsonutil"
)

const infoFile = ".info.json"

type folderDisk struct {
	ID        string          `json:"id"`
	Name      QuotedString    `json:"name"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Data      json.RawMessage `json:"data"`
}

func decodeFolder[T any](d folderDisk) (FolderInfo[T], error) {
	updated := d.UpdatedAt
	if updated.IsZero() {
		updated = d.CreatedAt
	}
	var data T
	raw := d.Data
	if len(raw) == 0 {
		raw = json.RawMessage("null")
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return FolderInfo[T]{}, ErrInternal
	}
	return FolderInfo[T]{
		ID:        d.ID,
		Name:      d.Name.String(),
		CreatedAt: d.CreatedAt.UTC(),
		UpdatedAt: updated.UTC(),
		Data:      data,
	}, nil
}

func marshalPayload[T any](data T) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, ErrInternal
	}
	return raw, nil
}

func (db *DB) writeInfoCreate(dir string, disk folderDisk) error {
	raw, err := jsonutil.Encode(disk, db.pretty)
	if err != nil {
		return ErrInternal
	}
	path := filepath.Join(dir, infoFile)
	file, err := fs.CreateFile(path)
	if err != nil {
		return err
	}
	if err := fs.Write(file, raw); err != nil {
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

func (db *DB) writeInfoReplace(dir string, disk folderDisk) error {
	raw, err := jsonutil.Encode(disk, db.pretty)
	if err != nil {
		return ErrInternal
	}
	tmp := filepath.Join(dir, infoFile+".tmp")
	final := filepath.Join(dir, infoFile)
	_ = fs.RemoveFile(tmp)
	file, err := fs.CreateFile(tmp)
	if err != nil {
		return err
	}
	if err := fs.Write(file, raw); err != nil {
		_ = fs.Close(file)
		_ = fs.RemoveFile(tmp)
		return err
	}
	if err := fs.Sync(file); err != nil {
		_ = fs.Close(file)
		_ = fs.RemoveFile(tmp)
		if db.log != nil {
			db.log.Error("sync failed")
		}
		return err
	}
	if err := fs.Close(file); err != nil {
		_ = fs.RemoveFile(tmp)
		return err
	}
	if err := fs.RenameFile(tmp, final); err != nil {
		_ = fs.RemoveFile(tmp)
		return err
	}
	return nil
}

func (db *DB) readInfo(dir string) (folderDisk, error) {
	path := filepath.Join(dir, infoFile)
	file, err := fs.OpenRead(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return folderDisk{}, ErrFolderCorrupted
		}
		return folderDisk{}, err
	}
	buf, err := fs.Read(file, nil)
	closeErr := fs.Close(file)
	if err != nil {
		return folderDisk{}, err
	}
	if closeErr != nil {
		return folderDisk{}, closeErr
	}
	var disk folderDisk
	if err := json.Unmarshal(buf, &disk); err != nil {
		return folderDisk{}, ErrFolderCorrupted
	}
	return disk, nil
}

func (db *DB) statDir(path string, missingBadPath bool) error {
	info, err := fs.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if missingBadPath {
				return ErrBadPath
			}
			return ErrNotExist
		}
		return err
	}
	if !info.IsDir() {
		return ErrNotDirectory
	}
	return nil
}
