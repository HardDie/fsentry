//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/HardDie/fsentry"
)

type tag struct {
	Tag string `json:"tag"`
	N   int    `json:"n"`
}

type body struct {
	Text string `json:"text"`
}

type discardLog struct{}

func (discardLog) Debug(string, ...any) {}
func (discardLog) Info(string, ...any)  {}
func (discardLog) Warn(string, ...any)  {}
func (discardLog) Error(string, ...any) {}

type apiSnap struct {
	folders  map[string]fsentry.FolderInfo[json.RawMessage]
	entries  map[string]fsentry.Entry[json.RawMessage]
	binaries map[string][]byte
}

// TestFullPublicFlow is a single external integration test that calls every
// exported store method and checks round-trips, isolated deletes, and zip identity.
func TestFullPublicFlow(t *testing.T) {
	assertPublicHelpers(t)

	root := t.TempDir()
	db := fsentry.New(
		root,
		fsentry.WithPretty(),
		fsentry.WithLogger(discardLog{}),
		fsentry.WithLockTimeout(time.Minute),
		fsentry.WithNoLockFile(),
	)
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}

	mustValidate(t, db)
	empty, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(empty.Folders)+len(empty.Entries)+len(empty.Binaries)+len(empty.CorruptedFolder) != 0 {
		t.Fatalf("empty store: %+v", empty)
	}

	keep := mustCreateFolder(t, db, "Keep Forever", tag{Tag: "keep", N: 1})
	games := mustCreateFolder(t, db, "Games", tag{Tag: "games", N: 1})
	game := mustCreateFolder(t, db, "My Game", tag{Tag: "game", N: 1}, "Games")
	scratch := mustCreateFolder(t, db, "Scratch", tag{Tag: "scratch", N: 1})

	settings := mustCreateEntry(t, db, "settings", body{Text: "root"})
	rules := mustCreateEntry(t, db, "rules", body{Text: "draw"}, "games", "my_game")
	scratchNote := mustCreateEntry(t, db, "todo", body{Text: "temp"}, "scratch")

	mustCreateBinary(t, db, "icon", []byte("ICON-v1"))
	mustCreateBinary(t, db, "cover", []byte("COVER-v1"), "games", "my_game")
	mustCreateBinary(t, db, "blob", []byte("SCRATCH-BIN"), "scratch")

	assertList(t, db, nil, []string{"games", "keep_forever", "scratch"}, []string{"settings"}, []string{"icon"})
	assertList(t, db, []string{"games"}, []string{"my_game"}, nil, nil)
	assertList(t, db, []string{"games", "my_game"}, nil, []string{"rules"}, []string{"cover"})
	mustValidate(t, db)

	_ = keep
	_ = games
	_ = game
	_ = scratch
	_ = settings
	_ = rules
	_ = scratchNote

	updatedKeep := keep
	updatedGame := mustUpdateFolder(t, db, "My Game", tag{Tag: "game", N: 2}, "Games")
	if updatedGame.CreatedAt.After(updatedGame.UpdatedAt) {
		t.Fatalf("folder times %+v", updatedGame)
	}
	gotKeep, err := db.GetFolder[tag]("Keep Forever")
	if err != nil {
		t.Fatal(err)
	}
	assertFolder(t, updatedKeep, gotKeep)

	updatedSettings := mustUpdateEntry(t, db, "settings", body{Text: "root-v2"})
	if err := db.UpdateBinary("icon", []byte("ICON-v2")); err != nil {
		t.Fatal(err)
	}
	icon, err := db.GetBinary("icon", make([]byte, 0, 32))
	if err != nil {
		t.Fatal(err)
	}
	if string(icon) != "ICON-v2" {
		t.Fatalf("icon %q", icon)
	}
	cover, err := db.GetBinary("cover", nil, "games", "my_game")
	if err != nil {
		t.Fatal(err)
	}
	if string(cover) != "COVER-v1" {
		t.Fatal("unrelated binary changed")
	}
	_ = updatedSettings

	movedRules := mustMoveEntry(t, db, "rules", "handbook", "games", "my_game")
	if _, err := db.GetEntry[body]("rules", "games", "my_game"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("old entry still there: %v", err)
	}
	if err := db.MoveBinary("cover", "art", "games", "my_game"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetBinary("cover", nil, "games", "my_game"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatal(err)
	}
	art, err := db.GetBinary("art", nil, "games", "my_game")
	if err != nil {
		t.Fatal(err)
	}
	if string(art) != "COVER-v1" {
		t.Fatalf("moved binary %q", art)
	}

	movedScratch := mustMoveFolder(t, db, "Scratch", "Junk")
	if _, err := db.GetFolder[tag]("Scratch"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatal(err)
	}
	todo, err := db.GetEntry[body]("todo", "junk")
	if err != nil {
		t.Fatal(err)
	}
	if todo.Data.Text != "temp" {
		t.Fatalf("child after move: %+v", todo)
	}
	blob, err := db.GetBinary("blob", nil, "junk")
	if err != nil {
		t.Fatal(err)
	}
	if string(blob) != "SCRATCH-BIN" {
		t.Fatal(err)
	}
	_ = movedRules
	_ = movedScratch

	dup, err := db.DuplicateEntry[body]("settings", "settings backup")
	if err != nil {
		t.Fatal(err)
	}
	if dup.ID != "settings_backup" || dup.Data.Text != "root-v2" {
		t.Fatalf("%+v", dup)
	}
	orig, err := db.GetEntry[body]("settings")
	if err != nil {
		t.Fatal(err)
	}
	if orig.Data.Text != "root-v2" {
		t.Fatal("duplicate mutated source")
	}
	gotDup, err := db.GetEntry[body]("settings backup")
	if err != nil {
		t.Fatal(err)
	}
	assertEntry(t, dup, gotDup)
	mustValidate(t, db)

	folderDup, err := db.DuplicateFolder[tag]("Keep Forever", "Keep Copy")
	if err != nil {
		t.Fatal(err)
	}
	if folderDup.ID != "keep_copy" || folderDup.Data.Tag != "keep" {
		t.Fatalf("%+v", folderDup)
	}
	if _, err := db.GetFolder[tag]("Keep Forever"); err != nil {
		t.Fatal(err)
	}
	mustValidate(t, db)

	renamedCopy, err := db.UpdateFolderNameWithoutTimestamp[tag]("Keep Copy", "Keep Snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if renamedCopy.ID != "keep_snapshot" {
		t.Fatalf("%+v", renamedCopy)
	}
	if !renamedCopy.CreatedAt.Equal(folderDup.CreatedAt) || !renamedCopy.UpdatedAt.Equal(folderDup.UpdatedAt) {
		t.Fatalf("timestamps bumped: %+v vs %+v", folderDup, renamedCopy)
	}
	if _, err := db.GetFolder[tag]("Keep Copy"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatal(err)
	}
	mustValidate(t, db)

	beforeFiles := fileTree(t, root)
	beforeAPI := collectAPI(t, db)

	var zipBuf bytes.Buffer
	if err := db.Export(&zipBuf); err != nil {
		t.Fatal(err)
	}
	dstRoot := t.TempDir()
	dst := fsentry.New(dstRoot, fsentry.WithNoLockFile())
	if err := dst.Init(); err != nil {
		t.Fatal(err)
	}
	if err := dst.Import(bytes.NewReader(zipBuf.Bytes())); err != nil {
		t.Fatal(err)
	}
	mustValidate(t, dst)
	afterFiles := fileTree(t, dstRoot)
	if !mapsEqual(beforeFiles, afterFiles) {
		t.Fatalf("import files differ\nwant keys %v\ngot keys %v", keys(beforeFiles), keys(afterFiles))
	}
	for k, want := range beforeFiles {
		if !bytes.Equal(want, afterFiles[k]) {
			t.Fatalf("file %s not identical after import", k)
		}
	}
	afterAPI := collectAPI(t, dst)
	assertAPIEqual(t, beforeAPI, afterAPI)

	var folderZip bytes.Buffer
	if err := db.Export(&folderZip, "Games"); err != nil {
		t.Fatal(err)
	}
	folderDstRoot := t.TempDir()
	folderDst := fsentry.New(folderDstRoot, fsentry.WithNoLockFile())
	if err := folderDst.Init(); err != nil {
		t.Fatal(err)
	}
	if err := folderDst.Import(bytes.NewReader(folderZip.Bytes())); err != nil {
		t.Fatal(err)
	}
	mustValidate(t, folderDst)
	wantGames := prefixTree(beforeFiles, "games/")
	gotGames := fileTree(t, folderDstRoot)
	if !mapsEqual(wantGames, gotGames) {
		t.Fatalf("folder export/import files\nwant %v\ngot %v", keys(wantGames), keys(gotGames))
	}
	for k, want := range wantGames {
		if !bytes.Equal(want, gotGames[k]) {
			t.Fatalf("folder zip file %s mismatch", k)
		}
	}

	if err := db.RemoveEntry("settings backup"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetEntry[body]("settings backup"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatal(err)
	}
	still, err := db.GetEntry[body]("settings")
	if err != nil || still.Data.Text != "root-v2" {
		t.Fatalf("settings after sibling delete: %v %+v", err, still)
	}

	if err := db.RemoveBinary("icon"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetBinary("icon", nil); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatal(err)
	}
	art2, err := db.GetBinary("art", nil, "games", "my_game")
	if err != nil || string(art2) != "COVER-v1" {
		t.Fatalf("art after icon delete: %v %q", err, art2)
	}

	if err := db.RemoveFolder("Junk"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetFolder[tag]("Junk"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatal(err)
	}
	if _, err := db.GetEntry[body]("todo", "junk"); !errors.Is(err, fsentry.ErrNotExist) && !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("todo after junk delete: %v", err)
	}
	keep2, err := db.GetFolder[tag]("Keep Forever")
	if err != nil {
		t.Fatal(err)
	}
	assertFolder(t, gotKeep, keep2)
	handbook, err := db.GetEntry[body]("handbook", "games", "my_game")
	if err != nil || handbook.Data.Text != "draw" {
		t.Fatalf("handbook after junk delete: %v %+v", err, handbook)
	}
	mustValidate(t, db)
	assertList(t, db, nil, []string{"games", "keep_forever", "keep_snapshot"}, []string{"settings"}, nil)

	if err := db.Drop(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("drop left root: %v", err)
	}

	lockRoot := t.TempDir()
	locked := fsentry.New(lockRoot, fsentry.WithLockFile(), fsentry.WithLockTimeout(time.Minute))
	if err := locked.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(lockRoot, ".fsentry.lock")); err != nil {
		t.Fatal(err)
	}
	if err := locked.Drop(); err != nil {
		t.Fatal(err)
	}
}

func assertPublicHelpers(t *testing.T) {
	t.Helper()
	if fsentry.NameToID("My Game") != "my_game" {
		t.Fatal(fsentry.NameToID("My Game"))
	}
	raw, err := json.Marshal(fsentry.QuotedString("My Notes"))
	if err != nil {
		t.Fatal(err)
	}
	var qs fsentry.QuotedString
	if err := json.Unmarshal(raw, &qs); err != nil {
		t.Fatal(err)
	}
	if qs.String() != "My Notes" {
		t.Fatal(qs.String())
	}
	p := fsentry.Problem{Path: "broken", Code: fsentry.ProblemInfo}
	if p.String() != "info: broken" {
		t.Fatal(p.String())
	}
}

func mustCreateFolder(t *testing.T, db *fsentry.DB, name string, data tag, path ...string) fsentry.FolderInfo[tag] {
	t.Helper()
	created, err := db.CreateFolder(name, data, path...)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != fsentry.NameToID(name) || created.Name != name || created.Data != data {
		t.Fatalf("create folder return %+v", created)
	}
	if created.CreatedAt.IsZero() || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("create folder times %+v", created)
	}
	got, err := db.GetFolder[tag](name, path...)
	if err != nil {
		t.Fatal(err)
	}
	assertFolder(t, created, got)
	return created
}

func mustUpdateFolder(t *testing.T, db *fsentry.DB, name string, data tag, path ...string) fsentry.FolderInfo[tag] {
	t.Helper()
	updated, err := db.UpdateFolder(name, data, path...)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Data != data {
		t.Fatalf("%+v", updated)
	}
	got, err := db.GetFolder[tag](name, path...)
	if err != nil {
		t.Fatal(err)
	}
	assertFolder(t, updated, got)
	return updated
}

func mustMoveFolder(t *testing.T, db *fsentry.DB, oldName, newName string, path ...string) fsentry.FolderInfo[tag] {
	t.Helper()
	moved, err := db.MoveFolder[tag](oldName, newName, path...)
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != fsentry.NameToID(newName) || moved.Name != newName {
		t.Fatalf("%+v", moved)
	}
	got, err := db.GetFolder[tag](newName, path...)
	if err != nil {
		t.Fatal(err)
	}
	assertFolder(t, moved, got)
	return moved
}

func mustCreateEntry(t *testing.T, db *fsentry.DB, name string, data body, path ...string) fsentry.Entry[body] {
	t.Helper()
	created, err := db.CreateEntry(name, data, path...)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != fsentry.NameToID(name) || created.Name != name || created.Data != data {
		t.Fatalf("%+v", created)
	}
	if created.CreatedAt.IsZero() || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("times %+v", created)
	}
	got, err := db.GetEntry[body](name, path...)
	if err != nil {
		t.Fatal(err)
	}
	assertEntry(t, created, got)
	return created
}

func mustUpdateEntry(t *testing.T, db *fsentry.DB, name string, data body, path ...string) fsentry.Entry[body] {
	t.Helper()
	updated, err := db.UpdateEntry(name, data, path...)
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.GetEntry[body](name, path...)
	if err != nil {
		t.Fatal(err)
	}
	assertEntry(t, updated, got)
	return updated
}

func mustMoveEntry(t *testing.T, db *fsentry.DB, oldName, newName string, path ...string) fsentry.Entry[body] {
	t.Helper()
	moved, err := db.MoveEntry[body](oldName, newName, path...)
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.GetEntry[body](newName, path...)
	if err != nil {
		t.Fatal(err)
	}
	assertEntry(t, moved, got)
	return moved
}

func mustCreateBinary(t *testing.T, db *fsentry.DB, name string, data []byte, path ...string) {
	t.Helper()
	if err := db.CreateBinary(name, data, path...); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetBinary(name, nil, path...)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("binary %s: got %q want %q", name, got, data)
	}
}

func assertFolder(t *testing.T, want, got fsentry.FolderInfo[tag]) {
	t.Helper()
	if want.ID != got.ID || want.Name != got.Name || want.Data != got.Data {
		t.Fatalf("folder want %+v got %+v", want, got)
	}
	if !want.CreatedAt.Equal(got.CreatedAt) || !want.UpdatedAt.Equal(got.UpdatedAt) {
		t.Fatalf("folder times want %+v got %+v", want, got)
	}
}

func assertEntry(t *testing.T, want, got fsentry.Entry[body]) {
	t.Helper()
	if want.ID != got.ID || want.Name != got.Name || want.Data != got.Data {
		t.Fatalf("entry want %+v got %+v", want, got)
	}
	if !want.CreatedAt.Equal(got.CreatedAt) || !want.UpdatedAt.Equal(got.UpdatedAt) {
		t.Fatalf("entry times want %+v got %+v", want, got)
	}
}

func mustValidate(t *testing.T, db *fsentry.DB) {
	t.Helper()
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 0 {
		t.Fatalf("validate %v", probs)
	}
}

func assertList(t *testing.T, db *fsentry.DB, path []string, folders, entries, binaries []string) {
	t.Helper()
	list, err := db.List(path...)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.CorruptedFolder) != 0 {
		t.Fatalf("corrupted %v", list.CorruptedFolder)
	}
	assertIDs(t, "folders", list.Folders, folders)
	assertIDs(t, "entries", list.Entries, entries)
	assertIDs(t, "binaries", list.Binaries, binaries)
}

func assertIDs(t *testing.T, kind string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v want %v", kind, got, want)
	}
	set := map[string]int{}
	for _, w := range want {
		set[w]++
	}
	for _, g := range got {
		set[g]--
	}
	for id, n := range set {
		if n != 0 {
			t.Fatalf("%s mismatch on %s: got %v want %v", kind, id, got, want)
		}
	}
}

func collectAPI(t *testing.T, db *fsentry.DB) apiSnap {
	t.Helper()
	s := apiSnap{
		folders:  map[string]fsentry.FolderInfo[json.RawMessage]{},
		entries:  map[string]fsentry.Entry[json.RawMessage]{},
		binaries: map[string][]byte{},
	}
	var walk func(segs ...string)
	walk = func(segs ...string) {
		list, err := db.List(segs...)
		if err != nil {
			t.Fatal(err)
		}
		if len(list.CorruptedFolder) != 0 {
			t.Fatalf("corrupted %v at %v", list.CorruptedFolder, segs)
		}
		for _, id := range list.Folders {
			f, err := db.GetFolder[json.RawMessage](id, segs...)
			if err != nil {
				t.Fatal(err)
			}
			s.folders[joinPath(segs, id)] = f
			next := append(append([]string{}, segs...), id)
			walk(next...)
		}
		for _, id := range list.Entries {
			e, err := db.GetEntry[json.RawMessage](id, segs...)
			if err != nil {
				t.Fatal(err)
			}
			s.entries[joinPath(segs, id)] = e
		}
		for _, id := range list.Binaries {
			b, err := db.GetBinary(id, nil, segs...)
			if err != nil {
				t.Fatal(err)
			}
			s.binaries[joinPath(segs, id)] = append([]byte{}, b...)
		}
	}
	walk()
	return s
}

func assertAPIEqual(t *testing.T, want, got apiSnap) {
	t.Helper()
	if len(want.folders) != len(got.folders) || len(want.entries) != len(got.entries) || len(want.binaries) != len(got.binaries) {
		t.Fatalf("api size folders %d/%d entries %d/%d binaries %d/%d",
			len(want.folders), len(got.folders), len(want.entries), len(got.entries), len(want.binaries), len(got.binaries))
	}
	for k, w := range want.folders {
		g, ok := got.folders[k]
		if !ok {
			t.Fatalf("missing folder %s", k)
		}
		if w.ID != g.ID || w.Name != g.Name || !bytes.Equal(w.Data, g.Data) || !w.CreatedAt.Equal(g.CreatedAt) || !w.UpdatedAt.Equal(g.UpdatedAt) {
			t.Fatalf("folder %s want %+v got %+v", k, w, g)
		}
	}
	for k, w := range want.entries {
		g, ok := got.entries[k]
		if !ok {
			t.Fatalf("missing entry %s", k)
		}
		if w.ID != g.ID || w.Name != g.Name || !bytes.Equal(w.Data, g.Data) || !w.CreatedAt.Equal(g.CreatedAt) || !w.UpdatedAt.Equal(g.UpdatedAt) {
			t.Fatalf("entry %s want %+v got %+v", k, w, g)
		}
	}
	for k, w := range want.binaries {
		if !bytes.Equal(w, got.binaries[k]) {
			t.Fatalf("binary %s mismatch", k)
		}
	}
}

func joinPath(segs []string, id string) string {
	if len(segs) == 0 {
		return id
	}
	return path.Join(append(append([]string{}, segs...), id)...)
}

func fileTree(t *testing.T, root string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == ".fsentry.lock" {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out[rel] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func prefixTree(tree map[string][]byte, prefix string) map[string][]byte {
	out := map[string][]byte{}
	for k, v := range tree {
		if strings.HasPrefix(k, prefix) {
			out[k] = v
		}
	}
	return out
}

func mapsEqual(a, b map[string][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
