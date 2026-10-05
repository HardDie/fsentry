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
	"time"

	"github.com/HardDie/fsentry/internal/fs"
	"github.com/HardDie/fsentry/internal/jsonutil"
	"github.com/HardDie/fsentry/internal/lock"
)

const zipCopyChunk = 32 * 1024

// Export writes a zip archive of the directory at path (empty path is the store
// root). Zip names use `/`. The lock file and write temps (`*.tmp`) are omitted.
// A non-root path is stored
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
// with the folder's ID. The lock file and write temps (`*.tmp`) are omitted.
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
// leave the destination are ErrBadArchive. `.fsentry.lock` and write temps
// (`*.tmp`) are skipped.
// A failed import removes files and directories this call created.
// To import one folder archive, use ImportFolder.
// To follow a long import, use ImportWithProgress.
func (db *DB) Import(r io.Reader, pathSegs ...string) error {
	return db.importZip(r, nil, pathSegs...)
}

// ImportProgress is one reading of a running import.
// Files counts extracted files; directories are not counted.
// Bytes counts their uncompressed size, as the zip headers declare it.
type ImportProgress struct {
	Files      int
	FilesTotal int
	Bytes      int64
	BytesTotal int64
}

// ImportWithProgress is Import that reports how far extraction got.
// progress is called once before the first file (Files is 0), then after
// every file written. A nil progress is the same as Import.
// progress runs while the store holds its write lock: it must not call the
// same *DB, and it should return quickly.
func (db *DB) ImportWithProgress(r io.Reader, progress func(ImportProgress), pathSegs ...string) error {
	return db.importZip(r, progress, pathSegs...)
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
// An existing folder with the destination id is ErrExist and is not modified.
// Other children of path are not touched. A failed import removes files and
// directories this call created.
// To follow a long import, use ImportFolderWithProgress.
func (db *DB) ImportFolder(r io.Reader, name string, path ...string) (string, error) {
	return db.ImportFolderWithProgress(r, name, nil, path...)
}

// ImportFolderWithProgress is ImportFolder that reports how far extraction
// got. progress is called as in ImportWithProgress; it is not called when the
// archive is refused before any file is written. A nil progress is the same
// as ImportFolder.
func (db *DB) ImportFolderWithProgress(r io.Reader, name string, progress func(ImportProgress), path ...string) (string, error) {
	l := db.startImport("import folder", "path", strings.Join(path, "/"), "name", name)
	if r == nil {
		l.end(ErrBadArchive)
		return "", ErrBadArchive
	}
	var id string
	err := db.withLock(true, func() error {
		l.locked()
		parent, err := db.ensurePath(path...)
		if err != nil {
			return err
		}
		entries, err := l.readZip(r)
		if err != nil {
			return err
		}
		l.step = "find root"
		root, err := requireSingleNode(entries, l)
		if err != nil {
			return err
		}
		if root == "" || NameToID(root) != root {
			l.log.Debug("import folder: root is not an id", "root", root)
			return ErrBadArchive
		}
		destID := root
		if name != "" {
			l.step = "rename"
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
		l.at("root found", "root", root, "dest", destID)
		l.step = "check destination"
		destDir := filepath.Join(parent, destID)
		if !underRoot(parent, destDir) || !underRoot(db.root, destDir) {
			return ErrBadArchive
		}
		_, err = fs.Stat(destDir)
		switch {
		case err == nil:
			l.log.Debug("import folder: destination exists", "dir", destDir)
			return ErrExist
		case errors.Is(err, fs.ErrNotExist):
		default:
			return err
		}
		if err := l.checkTargets(parent, db.root, entries); err != nil {
			return err
		}
		if err := l.extract(parent, entries, progress); err != nil {
			return err
		}
		id = destID
		return nil
	})
	l.end(err, "id", id)
	return id, err
}

func (db *DB) importZip(r io.Reader, progress func(ImportProgress), pathSegs ...string) error {
	l := db.startImport("import", "path", strings.Join(pathSegs, "/"))
	if r == nil {
		l.end(ErrBadArchive)
		return ErrBadArchive
	}
	err := db.withLock(true, func() error {
		l.locked()
		dir, err := db.ensurePath(pathSegs...)
		if err != nil {
			return err
		}
		entries, err := l.readZip(r)
		if err != nil {
			return err
		}
		if err := l.checkTargets(dir, db.root, entries); err != nil {
			return err
		}
		return l.extract(dir, entries, progress)
	})
	l.end(err)
	return err
}

// importLog writes the Debug lines of one import, so a slow or failing import
// can be followed step by step. step is where the import is; a failure names it.
type importLog struct {
	log   Logger
	op    string
	start time.Time
	step  string
}

func (db *DB) startImport(op string, args ...any) *importLog {
	l := &importLog{log: db.log, op: op, start: time.Now()}
	l.at("started", args...)
	return l
}

func (l *importLog) at(step string, args ...any) {
	l.step = step
	l.log.Debug(l.op+": "+step, args...)
}

// locked reports how long the import waited for the store lock.
func (l *importLog) locked() {
	l.at("locked", "wait_ms", time.Since(l.start).Milliseconds())
}

func (l *importLog) end(err error, args ...any) {
	ms := time.Since(l.start).Milliseconds()
	if err != nil {
		l.log.Debug(l.op+": failed", "step", l.step, "err", err, "duration_ms", ms)
		return
	}
	l.log.Debug(l.op+": finished", append(args, "duration_ms", ms)...)
}

func (l *importLog) readZip(r io.Reader) ([]zipEntry, error) {
	l.step = "open zip"
	start := time.Now()
	zr, size, err := openZipReader(r, l)
	if err != nil {
		return nil, err
	}
	entries, err := prepareZipEntries(zr, l)
	if err != nil {
		return nil, err
	}
	var files, dirs int
	var bytes int64
	for _, e := range entries {
		if e.dir {
			dirs++
			continue
		}
		files++
		bytes += e.size()
	}
	l.at("zip read", "zip_bytes", size, "entries", len(zr.File), "skipped", len(zr.File)-len(entries),
		"files", files, "dirs", dirs, "bytes", bytes, "duration_ms", time.Since(start).Milliseconds())
	return entries, nil
}

// checkTargets refuses an entry that would leave dir or the store, or land on
// an existing file or directory. Nothing is written before it passes.
func (l *importLog) checkTargets(dir, storeRoot string, entries []zipEntry) error {
	l.step = "check targets"
	for _, e := range entries {
		dest := filepath.Join(dir, filepath.FromSlash(e.rel))
		if !underRoot(dir, dest) || !underRoot(storeRoot, dest) {
			l.log.Debug(l.op+": entry leaves destination", "entry", e.rel)
			return ErrBadArchive
		}
		if e.dir {
			continue
		}
		info, err := fs.Stat(dest)
		switch {
		case err == nil:
			l.log.Debug(l.op+": target exists", "path", dest, "dir", info.IsDir())
			if info.IsDir() {
				return ErrIsDirectory
			}
			return ErrExist
		case errors.Is(err, fs.ErrNotExist):
		default:
			return err
		}
	}
	l.at("targets free", "dir", dir)
	return nil
}

// extract writes the entries; on failure it removes what this call created.
func (l *importLog) extract(dir string, entries []zipEntry, progress func(ImportProgress)) error {
	l.step = "extract"
	var created []string
	if err := extractZip(dir, entries, &created, progress, l); err != nil {
		rollbackCreated(created, l)
		return err
	}
	return nil
}

type zipEntry struct {
	rel  string
	dir  bool
	file *zip.File
	body []byte // set when ImportFolder rewrites .info.json before extract
}

// openZipReader opens the archive and returns its size. The zip error behind
// ErrBadArchive goes to the log.
func openZipReader(r io.Reader, l *importLog) (*zip.Reader, int64, error) {
	type sizeReaderAt interface {
		io.ReaderAt
		Size() int64
	}
	if ras, ok := r.(sizeReaderAt); ok {
		zr, err := zip.NewReader(ras, ras.Size())
		if err != nil {
			l.log.Debug(l.op+": not a zip", "err", err)
			return nil, 0, ErrBadArchive
		}
		return zr, ras.Size(), nil
	}
	data, err := io.ReadAll(r)
	if err != nil {
		l.log.Debug(l.op+": read archive", "err", err)
		return nil, 0, ErrInternal
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		l.log.Debug(l.op+": not a zip", "err", err)
		return nil, 0, ErrBadArchive
	}
	return zr, int64(len(data)), nil
}

func prepareZipEntries(zr *zip.Reader, l *importLog) ([]zipEntry, error) {
	var out []zipEntry
	for _, f := range zr.File {
		rel, dir, skip, err := zipRel(f.Name)
		if err != nil {
			l.log.Debug(l.op+": bad entry name", "entry", f.Name)
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
	if base == lock.FileName || (!dir && isTempFile(base)) {
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
func requireSingleNode(entries []zipEntry, l *importLog) (string, error) {
	refuse := func(reason, entry string) (string, error) {
		l.log.Debug(l.op+": not a single folder", "reason", reason, "entry", entry)
		return "", ErrBadArchive
	}
	if len(entries) == 0 {
		return refuse("empty archive", "")
	}
	var root string
	var fileUnder bool
	for _, e := range entries {
		top, rest, nested := strings.Cut(e.rel, "/")
		if root == "" {
			root = top
		} else if top != root {
			return refuse("second root", e.rel)
		}
		if !nested {
			if !e.dir {
				return refuse("file at zip root", e.rel)
			}
			continue
		}
		if rest == "" || rest == "." {
			return refuse("empty name", e.rel)
		}
		if !e.dir {
			fileUnder = true
		}
	}
	if root == "" || !fileUnder {
		return refuse("no file under root", root)
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
		if name == lock.FileName || (!e.IsDir() && isTempFile(name)) {
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

func extractZip(dir string, entries []zipEntry, created *[]string, progress func(ImportProgress), l *importLog) error {
	var p ImportProgress
	if progress != nil {
		for _, e := range entries {
			if !e.dir {
				p.FilesTotal++
				p.BytesTotal += e.size()
			}
		}
		progress(p)
	}
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
		start := time.Now()
		if err := extractZipFile(e.file, e.body, dest); err != nil {
			l.log.Debug(l.op+": write file", "entry", e.rel, "err", err)
			return err
		}
		*created = append(*created, dest)
		l.log.Debug(l.op+": file written", "entry", e.rel, "bytes", e.size(),
			"duration_ms", time.Since(start).Milliseconds())
		if progress != nil {
			p.Files++
			p.Bytes += e.size()
			progress(p)
		}
	}
	return nil
}

// size is the uncompressed size the entry writes.
func (e zipEntry) size() int64 {
	if e.body != nil {
		return int64(len(e.body))
	}
	if e.file == nil {
		return 0
	}
	return int64(e.file.UncompressedSize64) //nolint:gosec // zip sizes fit in int64
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

func rollbackCreated(created []string, l *importLog) {
	removed := 0
	for i := len(created) - 1; i >= 0; i-- {
		p := created[i]
		info, err := fs.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			err = fs.RemoveFolder(p)
		} else {
			err = fs.RemoveFile(p)
		}
		if err != nil {
			l.log.Error(l.op+": rollback remove", "path", p, "err", err)
			continue
		}
		removed++
	}
	l.log.Warn(l.op+": rolled back", "created", len(created), "removed", removed)
}
