package fsentry

import (
	"path/filepath"
	"strings"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/lock"
)

// List returns IDs of folders, entries, binaries, and corrupted folders at path.
// Empty path is the store root. `.info.json` and `.fsentry.lock` are skipped.
func (db *DB) List(path ...string) (List, error) {
	var out List
	err := db.withLock(false, func() error {
		dir, err := db.resolve(path...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, len(path) > 0); err != nil {
			return err
		}
		entries, err := fs.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			n := e.Name()
			if n == lock.FileName || n == infoFile {
				continue
			}
			full := filepath.Join(dir, n)
			if e.IsDir() {
				_, err := db.readInfo(full)
				if err != nil {
					out.CorruptedFolder = append(out.CorruptedFolder, n)
					continue
				}
				out.Folders = append(out.Folders, n)
				continue
			}
			switch {
			case strings.HasSuffix(n, ".json"):
				out.Entries = append(out.Entries, strings.TrimSuffix(n, ".json"))
			case strings.HasSuffix(n, ".bin"):
				out.Binaries = append(out.Binaries, strings.TrimSuffix(n, ".bin"))
			}
		}
		return nil
	})
	return out, err
}
