package fsentry

import (
	"errors"
	"path/filepath"

	"github.com/HardDie/fsentry/internal/fs"
)

const entrySuffix = ".json"

// CreateEntry creates `<id>.json` under path (empty path is the store root).
func (db *DB) CreateEntry[T any](name string, data T, path ...string) (Entry[T], error) {
	var out Entry[T]
	err := db.withLock(true, func() error {
		file, id, err := db.entryPath(name, path...)
		if err != nil {
			return err
		}
		payload, err := marshalPayload(data)
		if err != nil {
			return err
		}
		now := db.clock()
		env := envelope{
			ID:        id,
			Name:      QuotedString(name),
			CreatedAt: now,
			UpdatedAt: now,
			Data:      payload,
		}
		if err := db.writeJSONCreate(file, env); err != nil {
			return err
		}
		out = Entry[T]{
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

// GetEntry reads `<id>.json` and unmarshals data as T.
func (db *DB) GetEntry[T any](name string, path ...string) (Entry[T], error) {
	var out Entry[T]
	err := db.withLock(false, func() error {
		file, _, err := db.entryPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(file); err != nil {
			return err
		}
		env, err := db.readEnvelope(file)
		if err != nil {
			return mapEnvelopeRead(err)
		}
		out, err = decodeEntry[T](env)
		return err
	})
	return out, err
}

// MoveEntry renames the file and updates id/name in the envelope. updatedAt is now.
func (db *DB) MoveEntry[T any](oldName, newName string, path ...string) (Entry[T], error) {
	var out Entry[T]
	err := db.withLock(true, func() error {
		oldFile, _, err := db.entryPath(oldName, path...)
		if err != nil {
			return err
		}
		newFile, newID, err := db.entryPath(newName, path...)
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
		env, err := db.readEnvelope(oldFile)
		if err != nil {
			return mapEnvelopeRead(err)
		}
		if err := fs.RenameFile(oldFile, newFile); err != nil {
			return err
		}
		now := db.clock()
		env.ID = newID
		env.Name = QuotedString(newName)
		env.UpdatedAt = now
		if err := db.writeJSONReplace(newFile, env); err != nil {
			_ = fs.RenameFile(newFile, oldFile)
			return err
		}
		out, err = decodeEntry[T](env)
		return err
	})
	return out, err
}

// UpdateEntry replaces the payload and bumps updatedAt.
func (db *DB) UpdateEntry[T any](name string, data T, path ...string) (Entry[T], error) {
	var out Entry[T]
	err := db.withLock(true, func() error {
		file, _, err := db.entryPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(file); err != nil {
			return err
		}
		env, err := db.readEnvelope(file)
		if err != nil {
			return mapEnvelopeRead(err)
		}
		payload, err := marshalPayload(data)
		if err != nil {
			return err
		}
		now := db.clock()
		env.Data = payload
		env.UpdatedAt = now
		if err := db.writeJSONReplace(file, env); err != nil {
			return err
		}
		out = Entry[T]{
			ID:        env.ID,
			Name:      env.Name.String(),
			CreatedAt: env.CreatedAt.UTC(),
			UpdatedAt: now,
			Data:      data,
		}
		return nil
	})
	return out, err
}

// DuplicateEntry copies an entry to a new name with fresh timestamps.
func (db *DB) DuplicateEntry[T any](srcName, dstName string, path ...string) (Entry[T], error) {
	var out Entry[T]
	err := db.withLock(true, func() error {
		srcFile, _, err := db.entryPath(srcName, path...)
		if err != nil {
			return err
		}
		dstFile, dstID, err := db.entryPath(dstName, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(srcFile); err != nil {
			return err
		}
		env, err := db.readEnvelope(srcFile)
		if err != nil {
			return mapEnvelopeRead(err)
		}
		now := db.clock()
		env.ID = dstID
		env.Name = QuotedString(dstName)
		env.CreatedAt = now
		env.UpdatedAt = now
		if err := db.writeJSONCreate(dstFile, env); err != nil {
			return err
		}
		out, err = decodeEntry[T](env)
		return err
	})
	return out, err
}

// RemoveEntry deletes `<id>.json`.
func (db *DB) RemoveEntry(name string, path ...string) error {
	return db.withLock(true, func() error {
		file, _, err := db.entryPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statFile(file); err != nil {
			return err
		}
		return fs.RemoveFile(file)
	})
}

func (db *DB) entryPath(name string, path ...string) (file, id string, err error) {
	parent, err := db.ensurePath(path...)
	if err != nil {
		return "", "", err
	}
	id, err = db.objectID(name)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(parent, id+entrySuffix), id, nil
}
