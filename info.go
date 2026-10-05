package fsentry

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/jsonutil"
)

const infoFile = ".info.json"

var errMalformedJSON = errors.New("malformed json")

type envelope struct {
	ID        string          `json:"id"`
	Name      QuotedString    `json:"name"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Data      json.RawMessage `json:"data"`
}

func decodeEnvelope[T any](e envelope) (id, name string, created, updated time.Time, data T, err error) {
	updated = e.UpdatedAt
	if updated.IsZero() {
		updated = e.CreatedAt
	}
	raw := e.Data
	if len(raw) == 0 {
		raw = json.RawMessage("null")
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return "", "", time.Time{}, time.Time{}, data, ErrInternal
	}
	return e.ID, e.Name.String(), e.CreatedAt.UTC(), updated.UTC(), data, nil
}

func decodeFolder[T any](e envelope) (FolderInfo[T], error) {
	id, name, created, updated, data, err := decodeEnvelope[T](e)
	if err != nil {
		return FolderInfo[T]{}, err
	}
	return FolderInfo[T]{ID: id, Name: name, CreatedAt: created, UpdatedAt: updated, Data: data}, nil
}

func decodeEntry[T any](e envelope) (Entry[T], error) {
	id, name, created, updated, data, err := decodeEnvelope[T](e)
	if err != nil {
		return Entry[T]{}, err
	}
	return Entry[T]{ID: id, Name: name, CreatedAt: created, UpdatedAt: updated, Data: data}, nil
}

func marshalPayload[T any](data T) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, ErrInternal
	}
	return raw, nil
}

func (db *DB) writeInfoCreate(dir string, env envelope) error {
	return db.writeJSONCreate(filepath.Join(dir, infoFile), env)
}

func (db *DB) writeInfoReplace(dir string, env envelope) error {
	return db.writeJSONReplace(filepath.Join(dir, infoFile), env)
}

func (db *DB) writeJSONCreate(path string, env envelope) error {
	raw, err := jsonutil.Encode(env, db.pretty)
	if err != nil {
		return ErrInternal
	}
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
		db.log.Error("sync failed", "path", path, "err", err)
		return err
	}
	return fs.Close(file)
}

func (db *DB) writeJSONReplace(path string, env envelope) error {
	raw, err := jsonutil.Encode(env, db.pretty)
	if err != nil {
		return ErrInternal
	}
	return db.writeFileReplace(path, raw)
}

// isTempFile reports a write temp left by a crash (`<file>.tmp`).
// IDs never hold a dot, so no object file ends in `.tmp`.
func isTempFile(name string) bool {
	return strings.HasSuffix(name, ".tmp")
}

// writeFileReplace replaces path with data atomically: it writes `<path>.tmp`
// in the same folder, fsyncs it, then renames it over path.
// A crash leaves the old file or the new one, never a partial one.
// A `.tmp` left by a crash is removed by the next replace (Validate reports it).
func (db *DB) writeFileReplace(path string, data []byte) error {
	tmp := path + ".tmp"
	_ = fs.RemoveFile(tmp)
	file, err := fs.CreateFile(tmp)
	if err != nil {
		return err
	}
	if err := fs.Write(file, data); err != nil {
		_ = fs.Close(file)
		_ = fs.RemoveFile(tmp)
		return err
	}
	if err := fs.Sync(file); err != nil {
		_ = fs.Close(file)
		_ = fs.RemoveFile(tmp)
		db.log.Error("sync failed", "path", tmp, "err", err)
		return err
	}
	if err := fs.Close(file); err != nil {
		_ = fs.RemoveFile(tmp)
		return err
	}
	if err := fs.RenameFile(tmp, path); err != nil {
		_ = fs.RemoveFile(tmp)
		return err
	}
	return nil
}

func (db *DB) readInfo(dir string) (envelope, error) {
	env, err := db.readEnvelope(filepath.Join(dir, infoFile))
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, errMalformedJSON) {
		return envelope{}, ErrFolderCorrupted
	}
	return env, err
}

func (db *DB) readValidInfo(dir, id string) (envelope, error) {
	env, err := db.readInfo(dir)
	if err != nil {
		return envelope{}, err
	}
	if env.ID != id || NameToID(env.Name.String()) != id {
		return envelope{}, ErrFolderCorrupted
	}
	return env, nil
}

func (db *DB) checkFolderValid(dir, id string) error {
	_, err := db.readValidInfo(dir, id)
	return err
}

func (db *DB) readEnvelope(path string) (envelope, error) {
	file, err := fs.OpenRead(path)
	if err != nil {
		return envelope{}, err
	}
	buf, err := fs.Read(file, nil)
	closeErr := fs.Close(file)
	if err != nil {
		return envelope{}, err
	}
	if closeErr != nil {
		return envelope{}, closeErr
	}
	var env envelope
	if err := json.Unmarshal(buf, &env); err != nil {
		return envelope{}, errMalformedJSON
	}
	return env, nil
}

func mapEnvelopeRead(err error) error {
	if errors.Is(err, errMalformedJSON) {
		return ErrInternal
	}
	return err
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

func (db *DB) statFile(path string) error {
	info, err := fs.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNotExist
		}
		return err
	}
	if info.IsDir() {
		return ErrNotFile
	}
	return nil
}
