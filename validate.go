package fsentry

import (
	"path/filepath"
	"strings"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/lock"
)

// ProblemCode identifies a defect found by Validate.
type ProblemCode string

const (
	// ProblemInfo: folder `.info.json` is missing or not a valid envelope.
	ProblemInfo ProblemCode = "info"
	// ProblemJSON: an entry `*.json` is not a valid envelope.
	ProblemJSON ProblemCode = "json"
	// ProblemID: envelope `id` does not match the directory or file ID on disk.
	ProblemID ProblemCode = "id"
	// ProblemName: envelope `name` does not sanitize to the on-disk ID.
	ProblemName ProblemCode = "name"
	// ProblemBadName: the filesystem name is not a valid ID (NameToID).
	ProblemBadName ProblemCode = "bad_name"
	// ProblemTemp: leftover write temp (`*.tmp`).
	ProblemTemp ProblemCode = "temp"
	// ProblemUnexpected: not a folder, entry, binary, `.info.json`, or lock file.
	ProblemUnexpected ProblemCode = "unexpected"
)

// Problem is one on-disk defect. Path is relative to the validated directory
// and uses `/` (files keep their suffix: `welcome.json`, `cover.bin`).
type Problem struct {
	Path string
	Code ProblemCode
}

func (p Problem) String() string {
	if p.Path == "" {
		return string(p.Code)
	}
	return string(p.Code) + ": " + p.Path
}

// Validate walks the directory at path (empty path is the store root) and
// reports on-disk defects. It does not repair. A nil slice means the tree
// matches the store contract. Walk failures (permission, missing dest) are
// returned as err and problems may be incomplete.
func (db *DB) Validate(path ...string) ([]Problem, error) {
	var out []Problem
	err := db.withLock(false, func() error {
		dir, err := db.resolve(path...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, len(path) > 0); err != nil {
			return err
		}
		if len(path) > 0 {
			id, err := db.objectID(path[len(path)-1])
			if err != nil {
				return err
			}
			db.checkFolder(dir, id, &out)
			return db.walkValidate(dir, id, &out)
		}
		return db.walkValidate(dir, "", &out)
	})
	return out, err
}

func (db *DB) checkFolder(dir, rel string, out *[]Problem) {
	base := filepath.Base(dir)
	if !validDiskID(base) {
		*out = append(*out, Problem{Path: rel, Code: ProblemBadName})
	}
	env, err := db.readInfo(dir)
	if err != nil {
		*out = append(*out, Problem{Path: rel, Code: ProblemInfo})
		return
	}
	if env.ID != base {
		*out = append(*out, Problem{Path: rel, Code: ProblemID})
	}
	if NameToID(env.Name.String()) != base {
		*out = append(*out, Problem{Path: rel, Code: ProblemName})
	}
}

func (db *DB) walkValidate(dir, prefix string, out *[]Problem) error {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == lock.FileName {
			continue
		}
		abs := filepath.Clean(filepath.Join(dir, name))
		if !underRoot(db.root, abs) {
			return ErrBadPath
		}
		rel := name
		if prefix != "" {
			rel = prefix + "/" + name
		}
		info, err := fs.Stat(abs)
		if err != nil {
			return err
		}
		if info.IsDir() {
			db.checkFolder(abs, rel, out)
			if err := db.walkValidate(abs, rel, out); err != nil {
				return err
			}
			continue
		}
		if name == infoFile {
			if prefix == "" {
				*out = append(*out, Problem{Path: rel, Code: ProblemUnexpected})
			}
			continue
		}
		if strings.HasSuffix(name, ".tmp") {
			*out = append(*out, Problem{Path: rel, Code: ProblemTemp})
			continue
		}
		switch {
		case strings.HasSuffix(name, ".json"):
			db.checkEntry(abs, rel, strings.TrimSuffix(name, ".json"), out)
		case strings.HasSuffix(name, ".bin"):
			id := strings.TrimSuffix(name, ".bin")
			if !validDiskID(id) {
				*out = append(*out, Problem{Path: rel, Code: ProblemBadName})
			}
		default:
			*out = append(*out, Problem{Path: rel, Code: ProblemUnexpected})
		}
	}
	return nil
}

func (db *DB) checkEntry(file, rel, id string, out *[]Problem) {
	if !validDiskID(id) {
		*out = append(*out, Problem{Path: rel, Code: ProblemBadName})
	}
	env, err := db.readEnvelope(file)
	if err != nil {
		*out = append(*out, Problem{Path: rel, Code: ProblemJSON})
		return
	}
	if env.ID != id {
		*out = append(*out, Problem{Path: rel, Code: ProblemID})
	}
	if NameToID(env.Name.String()) != id {
		*out = append(*out, Problem{Path: rel, Code: ProblemName})
	}
}

func validDiskID(id string) bool {
	return id != "" && NameToID(id) == id
}
