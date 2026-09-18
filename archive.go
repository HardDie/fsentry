package fsentry

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/lock"
)

const zipCopyChunk = 32 * 1024

// Export writes a zip archive of the directory at path (empty path is the store
// root). Zip names use `/`. The lock file is omitted. A non-root path is stored
// under a top-level folder named with that folder's ID (DeckBuilder-style).
func (db *DB) Export(w io.Writer, pathSegs ...string) error {
	if w == nil {
		return ErrInternal
	}
	return db.withLock(false, func() error {
		dir, err := db.resolve(pathSegs...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, len(pathSegs) > 0); err != nil {
			return err
		}
		prefix := ""
		if len(pathSegs) > 0 {
			prefix, err = db.objectID(pathSegs[len(pathSegs)-1])
			if err != nil {
				return err
			}
		}
		zw := zip.NewWriter(w)
		err = db.addZipTree(zw, dir, prefix)
		closeErr := zw.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return ErrInternal
		}
		return nil
	})
}

// Import extracts a zip archive into the directory at path (empty path is the
// store root). Existing files are not overwritten (ErrExist). Paths that would
// leave the destination are ErrBadArchive. `.fsentry.lock` entries are skipped.
// A failed import removes files and directories this call created.
func (db *DB) Import(r io.Reader, pathSegs ...string) error {
	if r == nil {
		return ErrBadArchive
	}
	return db.withLock(true, func() error {
		dir, err := db.resolve(pathSegs...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, len(pathSegs) > 0); err != nil {
			return err
		}
		zr, err := openZipReader(r)
		if err != nil {
			return err
		}
		entries, err := prepareZipEntries(zr)
		if err != nil {
			return err
		}
		for _, e := range entries {
			dest := filepath.Join(dir, filepath.FromSlash(e.rel))
			if !underRoot(dir, dest) || !underRoot(db.root, dest) {
				return ErrBadArchive
			}
			if e.dir {
				continue
			}
			info, err := fs.Stat(dest)
			switch {
			case err == nil:
				if info.IsDir() {
					return ErrIsDirectory
				}
				return ErrExist
			case errors.Is(err, fs.ErrNotExist):
			default:
				return err
			}
		}
		var created []string
		err = extractZip(dir, entries, &created)
		if err != nil {
			rollbackCreated(created)
			return err
		}
		return nil
	})
}

type zipEntry struct {
	rel  string
	dir  bool
	file *zip.File
}

func openZipReader(r io.Reader) (*zip.Reader, error) {
	type sizeReaderAt interface {
		io.ReaderAt
		Size() int64
	}
	if ras, ok := r.(sizeReaderAt); ok {
		zr, err := zip.NewReader(ras, ras.Size())
		if err != nil {
			return nil, ErrBadArchive
		}
		return zr, nil
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, ErrInternal
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, ErrBadArchive
	}
	return zr, nil
}

func prepareZipEntries(zr *zip.Reader) ([]zipEntry, error) {
	var out []zipEntry
	for _, f := range zr.File {
		rel, dir, skip, err := zipRel(f.Name)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}
		isDir := dir || f.FileInfo().IsDir()
		out = append(out, zipEntry{rel: rel, dir: isDir, file: f})
	}
	return out, nil
}

func zipRel(name string) (rel string, dir, skip bool, err error) {
	name = strings.ReplaceAll(name, "\\", "/")
	if strings.HasSuffix(name, "/") {
		dir = true
		name = strings.TrimSuffix(name, "/")
	}
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		return "", false, true, nil
	}
	clean := path.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
		return "", false, false, ErrBadArchive
	}
	base := path.Base(clean)
	if base == lock.FileName {
		return "", false, true, nil
	}
	return clean, dir, false, nil
}

func (db *DB) addZipTree(zw *zip.Writer, absDir, zipPrefix string) error {
	entries, err := fs.ReadDir(absDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == lock.FileName {
			continue
		}
		abs := filepath.Clean(filepath.Join(absDir, name))
		if !underRoot(db.root, abs) {
			return ErrBadPath
		}
		zname := name
		if zipPrefix != "" {
			zname = zipPrefix + "/" + name
		}
		info, err := fs.Stat(abs)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := db.addZipTree(zw, abs, zname); err != nil {
				return err
			}
			continue
		}
		if err := addZipFile(zw, abs, zname); err != nil {
			return err
		}
	}
	return nil
}

func addZipFile(zw *zip.Writer, abs, zipName string) error {
	f, err := fs.OpenRead(abs)
	if err != nil {
		return err
	}
	defer fs.Close(f)
	w, err := zw.Create(zipName)
	if err != nil {
		return ErrInternal
	}
	return copyFileToWriter(w, f)
}

func copyFileToWriter(w io.Writer, file *os.File) error {
	buf := make([]byte, zipCopyChunk)
	var off int64
	for {
		n, err := fs.ReadAt(file, buf, off)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		if _, werr := w.Write(buf[:n]); werr != nil {
			return ErrInternal
		}
		off += int64(n)
		if n < len(buf) {
			return nil
		}
	}
}

func extractZip(dir string, entries []zipEntry, created *[]string) error {
	for _, e := range entries {
		dest := filepath.Join(dir, filepath.FromSlash(e.rel))
		if !underRoot(dir, dest) {
			return ErrBadArchive
		}
		if e.dir {
			if err := ensureDir(dest, dir, created); err != nil {
				return err
			}
			continue
		}
		if err := ensureDir(filepath.Dir(dest), dir, created); err != nil {
			return err
		}
		if err := extractZipFile(e.file, dest); err != nil {
			return err
		}
		*created = append(*created, dest)
	}
	return nil
}

func extractZipFile(zf *zip.File, dest string) error {
	src, err := zf.Open()
	if err != nil {
		return ErrBadArchive
	}
	defer src.Close()
	out, err := fs.CreateFile(dest)
	if err != nil {
		return err
	}
	buf := make([]byte, zipCopyChunk)
	for {
		n, rerr := src.Read(buf)
		if n > 0 {
			if werr := fs.Write(out, buf[:n]); werr != nil {
				_ = fs.Close(out)
				return werr
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			_ = fs.Close(out)
			return ErrBadArchive
		}
	}
	if err := fs.Sync(out); err != nil {
		_ = fs.Close(out)
		return err
	}
	return fs.Close(out)
}

func ensureDir(p, root string, created *[]string) error {
	p = filepath.Clean(p)
	root = filepath.Clean(root)
	if p == root {
		return nil
	}
	if !underRoot(root, p) {
		return ErrBadArchive
	}
	info, err := fs.Stat(p)
	switch {
	case err == nil:
		if !info.IsDir() {
			return ErrNotDirectory
		}
		return nil
	case !errors.Is(err, fs.ErrNotExist):
		return err
	}
	if err := ensureDir(filepath.Dir(p), root, created); err != nil {
		return err
	}
	if err := fs.CreateFolder(p); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return err
	}
	*created = append(*created, p)
	return nil
}

func rollbackCreated(created []string) {
	for i := len(created) - 1; i >= 0; i-- {
		p := created[i]
		info, err := fs.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			_ = fs.RemoveFolder(p)
			continue
		}
		_ = fs.RemoveFile(p)
	}
}
