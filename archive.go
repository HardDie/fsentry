package fsentry

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/jsonutil"
	"github.com/HardDie/fsentry/internal/lock"
)

const zipCopyChunk = 32 * 1024

// Export writes a zip archive of the directory at path (empty path is the store
// root). Zip names use `/`. The lock file is omitted. A non-root path is stored
// under a top-level folder named with that folder's ID (DeckBuilder-style).
// To export one folder by name, use ExportFolder.
func (db *DB) Export(w io.Writer, pathSegs ...string) error {
	if w == nil {
		return ErrInternal
	}
	return db.withLock(false, func() error {
		dir, err := db.ensurePath(pathSegs...)
		if err != nil {
			return err
		}
		prefix := ""
		if len(pathSegs) > 0 {
			prefix, err = db.objectID(pathSegs[len(pathSegs)-1])
			if err != nil {
				return err
			}
		}
		return db.writeZip(w, dir, prefix)
	})
}

// ExportFolder writes a zip of one folder and everything under it.
// name is that folder (display name or ID). path is its parent chain; empty
// path means a folder at the store root (a DeckBuilder game under "games"
// is ExportFolder(w, gameID, "games")). Zip names use `/` and are prefixed
// with the folder's ID. The lock file is omitted.
func (db *DB) ExportFolder(w io.Writer, name string, path ...string) error {
	if w == nil {
		return ErrInternal
	}
	return db.withLock(false, func() error {
		dir, id, err := db.folderPath(name, path...)
		if err != nil {
			return err
		}
		if err := db.statDir(dir, false); err != nil {
			return err
		}
		if err := db.checkFolderValid(dir, id); err != nil {
			return err
		}
		return db.writeZip(w, dir, id)
	})
}

// Import extracts a zip archive into the directory at path (empty path is the
// store root). Existing files are not overwritten (ErrExist). Paths that would
// leave the destination are ErrBadArchive. `.fsentry.lock` entries are skipped.
// A failed import removes files and directories this call created.
// To import one folder archive, use ImportFolder.
func (db *DB) Import(r io.Reader, pathSegs ...string) error {
	return db.importZip(r, pathSegs...)
}

// ImportFolder extracts one folder archive into the directory at path, which
// may already hold other folders (a DeckBuilder game into "games"). name is
// the display name to give that folder; empty keeps the archive's root id and
// its .info.json. The returned id is the folder written under path.
//
// The archive must contain exactly one top-level folder, and that name must
// already be a valid id. A whole-store zip, several root folders, a file at
// the zip root, or a root that is not an id is ErrBadArchive.
//
// When name is set, the zip root is rewritten to NameToID(name) before any
// file is written, and the folder's .info.json id and name are updated with
// timestamps left unchanged. An existing folder with the source id is left in
// place, so importing "Foo" as "Bar" while "Foo" already exists succeeds.
// An existing folder with the destination id is replaced. Other children of
// path are not touched. A failed import restores that previous folder and
// removes files this call created.
func (db *DB) ImportFolder(r io.Reader, name string, path ...string) (string, error) {
	if r == nil {
		return "", ErrBadArchive
	}
	var id string
	err := db.withLock(true, func() error {
		parent, err := db.ensurePath(path...)
		if err != nil {
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
		root, err := requireSingleNode(entries)
		if err != nil {
			return err
		}
		if root == "" || NameToID(root) != root {
			return ErrBadArchive
		}
		destID := root
		if name != "" {
			destID, err = db.objectID(name)
			if err != nil {
				return err
			}
			if err := retargetNode(entries, root, destID); err != nil {
				return err
			}
			if err := db.rewriteImportedInfo(entries, destID, name); err != nil {
				return err
			}
		}
		for _, e := range entries {
			dest := filepath.Join(parent, filepath.FromSlash(e.rel))
			if !underRoot(parent, dest) || !underRoot(db.root, dest) {
				return ErrBadArchive
			}
		}
		destDir := filepath.Join(parent, destID)
		if !underRoot(parent, destDir) || !underRoot(db.root, destDir) {
			return ErrBadArchive
		}
		aside := destDir + importReplaceSuffix
		if _, err := fs.Stat(aside); err == nil {
			return ErrExist
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		moved := false
		info, err := fs.Stat(destDir)
		switch {
		case err == nil && info.IsDir():
			if err := fs.RenameFolder(destDir, aside); err != nil {
				return err
			}
			moved = true
		case err == nil:
			return ErrExist
		case errors.Is(err, fs.ErrNotExist):
		default:
			return err
		}
		var created []string
		err = extractZip(parent, entries, &created)
		if err != nil {
			rollbackCreated(created)
			if moved {
				_ = fs.RenameFolder(aside, destDir)
			}
			return err
		}
		if moved {
			if err := fs.RemoveFolder(aside); err != nil {
				return err
			}
		}
		id = destID
		return nil
	})
	return id, err
}

func (db *DB) importZip(r io.Reader, pathSegs ...string) error {
	if r == nil {
		return ErrBadArchive
	}
	return db.withLock(true, func() error {
		dir, err := db.ensurePath(pathSegs...)
		if err != nil {
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

const importReplaceSuffix = ".replacing"

type zipEntry struct {
	rel  string
	dir  bool
	file *zip.File
	body []byte // set when ImportFolder rewrites .info.json before extract
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

func (db *DB) writeZip(w io.Writer, dir, prefix string) error {
	zw := zip.NewWriter(w)
	err := db.addZipTree(zw, dir, prefix)
	closeErr := zw.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return ErrInternal
	}
	return nil
}

// requireSingleNode returns the single top-level folder name. Entries must
// all live under that folder and include at least one file there.
func requireSingleNode(entries []zipEntry) (string, error) {
	if len(entries) == 0 {
		return "", ErrBadArchive
	}
	var root string
	var fileUnder bool
	for _, e := range entries {
		top, rest, nested := strings.Cut(e.rel, "/")
		if root == "" {
			root = top
		} else if top != root {
			return "", ErrBadArchive
		}
		if !nested {
			if !e.dir {
				return "", ErrBadArchive
			}
			continue
		}
		if rest == "" || rest == "." {
			return "", ErrBadArchive
		}
		if !e.dir {
			fileUnder = true
		}
	}
	if root == "" || !fileUnder {
		return "", ErrBadArchive
	}
	return root, nil
}

func retargetNode(entries []zipEntry, from, to string) error {
	if from == to {
		return nil
	}
	prefix := from + "/"
	for i := range entries {
		rel := entries[i].rel
		if rel == from {
			entries[i].rel = to
			continue
		}
		if !strings.HasPrefix(rel, prefix) {
			return ErrBadArchive
		}
		entries[i].rel = to + "/" + strings.TrimPrefix(rel, prefix)
	}
	return nil
}

// rewriteImportedInfo sets the folder envelope id and name. Timestamps and
// data stay as they were in the archive. A missing or unreadable .info.json
// is left unchanged so the caller can reject the folder after import.
func (db *DB) rewriteImportedInfo(entries []zipEntry, dirID, displayName string) error {
	infoRel := dirID + "/" + infoFile
	for i := range entries {
		if entries[i].dir || entries[i].rel != infoRel {
			continue
		}
		body, err := readZipBytes(entries[i])
		if err != nil {
			return err
		}
		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return nil
		}
		env.ID = dirID
		env.Name = QuotedString(displayName)
		raw, err := jsonutil.Encode(env, db.pretty)
		if err != nil {
			return ErrInternal
		}
		entries[i].body = raw
		return nil
	}
	return nil
}

func readZipBytes(e zipEntry) ([]byte, error) {
	if e.body != nil {
		return e.body, nil
	}
	if e.file == nil {
		return nil, ErrBadArchive
	}
	src, err := e.file.Open()
	if err != nil {
		return nil, ErrBadArchive
	}
	defer func() { _ = src.Close() }()
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, ErrBadArchive
	}
	return data, nil
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
	defer func() { _ = fs.Close(f) }()
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
		if err := extractZipFile(e.file, e.body, dest); err != nil {
			return err
		}
		*created = append(*created, dest)
	}
	return nil
}

func extractZipFile(zf *zip.File, body []byte, dest string) error {
	var src io.ReadCloser
	if body != nil {
		src = io.NopCloser(bytes.NewReader(body))
	} else {
		var err error
		src, err = zf.Open()
		if err != nil {
			return ErrBadArchive
		}
	}
	defer func() { _ = src.Close() }()
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
