package fsentry

import (
	"path/filepath"
	"strings"
	"time"

	nameto "github.com/HardDie/fsentry/internal/name"
)

func (db *DB) resolve(path ...string) (string, error) {
	if db.root == "" {
		return "", ErrBadPath
	}
	out := filepath.Clean(db.root)
	for _, seg := range path {
		if err := checkPathSegment(seg); err != nil {
			return "", err
		}
		id := nameto.Append(nil, seg)
		if len(id) == 0 {
			return "", ErrBadName
		}
		out = filepath.Join(out, string(id))
	}
	out = filepath.Clean(out)
	if !underRoot(db.root, out) {
		return "", ErrBadPath
	}
	return out, nil
}

// ensurePath resolves path from the store root. Empty path is the root (not a
// folder). Each path segment must be a directory with a valid `.info.json`
// (id and name match the disk ID). A missing segment is ErrBadPath.
func (db *DB) ensurePath(path ...string) (string, error) {
	if db.root == "" {
		return "", ErrBadPath
	}
	out := filepath.Clean(db.root)
	if err := db.statDir(out, false); err != nil {
		return "", err
	}
	for i, seg := range path {
		next, err := db.resolve(path[:i+1]...)
		if err != nil {
			return "", err
		}
		if err := db.statDir(next, true); err != nil {
			return "", err
		}
		id, err := db.objectID(seg)
		if err != nil {
			return "", err
		}
		if err := db.checkFolderValid(next, id); err != nil {
			return "", err
		}
		out = next
	}
	return out, nil
}

func checkPathSegment(seg string) error {
	if seg == "" || seg == "." || seg == ".." {
		return ErrBadPath
	}
	if strings.ContainsRune(seg, '/') || strings.ContainsRune(seg, '\\') {
		return ErrBadPath
	}
	if filepath.Separator != '/' && strings.ContainsRune(seg, filepath.Separator) {
		return ErrBadPath
	}
	return nil
}

func underRoot(root, p string) bool {
	root = filepath.Clean(root)
	p = filepath.Clean(p)
	if p == root {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(p, root+sep)
}

func (db *DB) objectID(name string) (string, error) {
	id := nameto.Append(nil, name)
	if len(id) == 0 {
		return "", ErrBadName
	}
	return string(id), nil
}

func (db *DB) clock() time.Time {
	if db.now != nil {
		return db.now().UTC()
	}
	return time.Now().UTC()
}
