package fsentry_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/HardDie/fsentry"
)

func openDB(t *testing.T) *fsentry.DB {
	t.Helper()
	db := fsentry.New(t.TempDir(), fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestInit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "store")
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("root: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
}

func TestInitEmptyRoot(t *testing.T) {
	db := fsentry.New("", fsentry.WithNoLockFile())
	if err := db.Init(); !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("got %v", err)
	}
}

func TestInitRootIsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	db := fsentry.New(path, fsentry.WithNoLockFile())
	if err := db.Init(); !errors.Is(err, fsentry.ErrNotDirectory) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateGetFolder(t *testing.T) {
	db := openDB(t)

	_, err := db.CreateFolder[any]("", nil)
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("empty name: %v", err)
	}
	_, err = db.CreateFolder[any]("bad_path", nil, "not_exist_folder")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("missing parent: %v", err)
	}

	got, err := db.CreateFolder("My Game", map[string]int{"n": 1})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "my_game" || got.Name != "My Game" {
		t.Fatalf("%+v", got)
	}
	if got.CreatedAt.IsZero() || !got.CreatedAt.Equal(got.UpdatedAt) {
		t.Fatalf("times %+v", got)
	}
	if got.Data["n"] != 1 {
		t.Fatalf("data %v", got.Data)
	}

	same, err := db.GetFolder[map[string]int]("My Game")
	if err != nil {
		t.Fatal(err)
	}
	byID, err := db.GetFolder[map[string]int]("my_game")
	if err != nil {
		t.Fatal(err)
	}
	if same.Name != "My Game" || byID.ID != "my_game" {
		t.Fatalf("get %+v %+v", same, byID)
	}
	if same.Data["n"] != 1 {
		t.Fatalf("data %v", same.Data)
	}

	_, err = db.CreateFolder[any]("some_inner_folder", nil, "My Game")
	if err != nil {
		t.Fatal(err)
	}
	inner, err := db.GetFolder[any]("some_inner_folder", "My Game")
	if err != nil {
		t.Fatal(err)
	}
	if inner.ID != "some_inner_folder" {
		t.Fatal(inner.ID)
	}

	_, err = db.CreateFolder[any]("My Game", nil)
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("dup: %v", err)
	}
}

func TestNestedPathRequiresValidFolders(t *testing.T) {
	dir := t.TempDir()
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("outer", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("inner", nil, "outer"); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(dir, "outer", ".info.json")); err != nil {
		t.Fatal(err)
	}
	_, err := db.GetFolder[any]("inner", "outer")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("missing info: %v", err)
	}
	_, err = db.CreateFolder[any]("child", nil, "outer")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("create under missing info: %v", err)
	}
	_, err = db.CreateEntry[any]("note", nil, "outer")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("entry under missing info: %v", err)
	}
	_, err = db.List("outer")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("list under missing info: %v", err)
	}

	if _, err := db.CreateFolder[any]("games", nil); err != nil {
		t.Fatal(err)
	}
	info := filepath.Join(dir, "games", ".info.json")
	raw, err := os.ReadFile(info)
	if err != nil {
		t.Fatal(err)
	}
	patched := bytes.Replace(raw, []byte(`"id":"games"`), []byte(`"id":"other"`), 1)
	if bytes.Equal(patched, raw) {
		t.Fatalf("id not found in %s", raw)
	}
	if err := os.WriteFile(info, patched, 0o666); err != nil {
		t.Fatal(err)
	}
	_, err = db.CreateFolder[any]("title", nil, "games")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("id mismatch: %v", err)
	}
	_, err = db.GetFolder[any]("games")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("get mismatch: %v", err)
	}

	list, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.CorruptedFolder) == 0 {
		t.Fatalf("root list should still report corrupted children: %+v", list)
	}
}

func TestGetFolderErrors(t *testing.T) {
	db := openDB(t)
	_, err := db.GetFolder[any]("")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.GetFolder[any]("missing")
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	_, err = db.GetFolder[any]("x", "missing_parent")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
}

func TestMoveFolder(t *testing.T) {
	db := openDB(t)
	_, err := db.MoveFolder[any]("not_exist", "new")
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	_, err = db.MoveFolder[any]("a", "b", "missing")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
	if _, err := db.CreateFolder[any]("first_folder", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("second_folder", nil); err != nil {
		t.Fatal(err)
	}
	_, err = db.MoveFolder[any]("", "x")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.MoveFolder[any]("first_folder", "")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.MoveFolder[any]("first_folder", "second_folder")
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("%v", err)
	}
	moved, err := db.MoveFolder[any]("first_folder", "new_first_folder")
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != "new_first_folder" || moved.Name != "new_first_folder" {
		t.Fatalf("%+v", moved)
	}
	if _, err := db.GetFolder[any]("first_folder"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
}

func TestUpdateFolder(t *testing.T) {
	db := openDB(t)
	_, err := db.UpdateFolder[any]("", nil)
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.UpdateFolder[any]("missing", nil)
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	created, err := db.CreateFolder("some_folder", 5)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := db.UpdateFolder("some_folder", 15)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Data != 15 {
		t.Fatalf("data %d", updated.Data)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatal("createdAt changed")
	}
}

func TestDuplicateFolder(t *testing.T) {
	db := openDB(t)
	_, err := db.DuplicateFolder[any]("missing", "copy")
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	_, err = db.DuplicateFolder[any]("a", "b", "missing")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
	src, err := db.CreateFolder("orig", map[string]int{"n": 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("child", nil, "orig"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("note", map[string]string{"t": "hi"}, "orig"); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateBinary("pic", []byte("PNG"), "orig"); err != nil {
		t.Fatal(err)
	}

	dup, err := db.DuplicateFolder[map[string]int]("orig", "copy")
	if err != nil {
		t.Fatal(err)
	}
	if dup.ID != "copy" || dup.Name != "copy" || dup.Data["n"] != 3 {
		t.Fatalf("%+v", dup)
	}

	still, err := db.GetFolder[map[string]int]("orig")
	if err != nil {
		t.Fatal(err)
	}
	if still.Data["n"] != 3 || still.ID != src.ID {
		t.Fatal("source changed")
	}
	note, err := db.GetEntry[map[string]string]("note", "copy")
	if err != nil {
		t.Fatal(err)
	}
	if note.Data["t"] != "hi" {
		t.Fatalf("%+v", note)
	}
	pic, err := db.GetBinary("pic", nil, "copy")
	if err != nil {
		t.Fatal(err)
	}
	if string(pic) != "PNG" {
		t.Fatalf("%q", pic)
	}
	child, err := db.GetFolder[any]("child", "copy")
	if err != nil {
		t.Fatal(err)
	}
	if child.ID != "child" {
		t.Fatal(child.ID)
	}

	_, err = db.DuplicateFolder[any]("orig", "copy")
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("dup exist: %v", err)
	}
	_, err = db.DuplicateFolder[any]("", "x")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.DuplicateFolder[any]("orig", "")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
}

func TestRemoveFolder(t *testing.T) {
	db := openDB(t)
	if err := db.RemoveFolder(""); !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	if err := db.RemoveFolder("missing"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	if _, err := db.CreateFolder[any]("some_folder", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("child", nil, "some_folder"); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveFolder("some_folder"); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveFolder("some_folder"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
}

func TestListFolders(t *testing.T) {
	dir := t.TempDir()
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("alpha", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("beta", nil); err != nil {
		t.Fatal(err)
	}
	list, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Folders) != 2 {
		t.Fatalf("folders %v", list.Folders)
	}
	if len(list.Entries) != 0 || len(list.Binaries) != 0 {
		t.Fatalf("%+v", list)
	}

	if err := os.Mkdir(filepath.Join(dir, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.json"), []byte(`{}`), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pic.bin"), []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	list, err = db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.CorruptedFolder) != 1 || list.CorruptedFolder[0] != "broken" {
		t.Fatalf("corrupted %v", list.CorruptedFolder)
	}
	if len(list.Entries) != 1 || list.Entries[0] != "note" {
		t.Fatalf("entries %v", list.Entries)
	}
	if len(list.Binaries) != 1 || list.Binaries[0] != "pic" {
		t.Fatalf("binaries %v", list.Binaries)
	}
}

func TestCreateFolderPretty(t *testing.T) {
	dir := t.TempDir()
	db := fsentry.New(dir, fsentry.WithNoLockFile(), fsentry.WithPretty())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder("pretty", map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "pretty", ".info.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("\t")) {
		t.Fatalf("want tab indent: %s", raw)
	}
}

func TestDrop(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "store")
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("f", nil); err != nil {
		t.Fatal(err)
	}
	if err := db.Drop(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("root still there: %v", err)
	}
}

func TestNameToID(t *testing.T) {
	if fsentry.NameToID("My Game") != "my_game" {
		t.Fatal(fsentry.NameToID("My Game"))
	}
	if fsentry.NameToID("con") != "" {
		t.Fatal("reserved")
	}
}
