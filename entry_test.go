package fsentry_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/HardDie/fsentry"
)

func TestCreateGetEntry(t *testing.T) {
	db := openDB(t)

	_, err := db.CreateEntry[any]("", nil)
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("empty name: %v", err)
	}
	_, err = db.CreateEntry[any]("bad_path", nil, "not_exist_folder")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("missing parent: %v", err)
	}

	got, err := db.CreateEntry("My Settings", map[string]string{"theme": "dark"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "my_settings" || got.Name != "My Settings" {
		t.Fatalf("%+v", got)
	}
	if got.Data["theme"] != "dark" {
		t.Fatalf("data %v", got.Data)
	}

	same, err := db.GetEntry[map[string]string]("My Settings")
	if err != nil {
		t.Fatal(err)
	}
	byID, err := db.GetEntry[map[string]string]("my_settings")
	if err != nil {
		t.Fatal(err)
	}
	if same.Data["theme"] != "dark" || byID.ID != "my_settings" {
		t.Fatalf("%+v %+v", same, byID)
	}

	if _, err := db.CreateFolder[any]("games", nil); err != nil {
		t.Fatal(err)
	}
	nested, err := db.CreateEntry("inner", 1, "games")
	if err != nil {
		t.Fatal(err)
	}
	if nested.ID != "inner" {
		t.Fatal(nested.ID)
	}

	_, err = db.CreateEntry[any]("My Settings", nil)
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("dup: %v", err)
	}

	list, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Entries) != 1 || list.Entries[0] != "my_settings" {
		t.Fatalf("list %+v", list)
	}
}

func TestGetEntryErrors(t *testing.T) {
	db := openDB(t)
	_, err := db.GetEntry[any]("")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.GetEntry[any]("missing")
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	_, err = db.GetEntry[any]("x", "missing_parent")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
}

func TestMoveEntry(t *testing.T) {
	db := openDB(t)
	_, err := db.MoveEntry[any]("not_exist", "new")
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	_, err = db.MoveEntry[any]("a", "b", "missing")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
	if _, err := db.CreateEntry[any]("first_entry", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry[any]("second_entry", nil); err != nil {
		t.Fatal(err)
	}
	_, err = db.MoveEntry[any]("", "x")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.MoveEntry[any]("first_entry", "")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.MoveEntry[any]("first_entry", "second_entry")
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("%v", err)
	}
	moved, err := db.MoveEntry[any]("first_entry", "new_first_entry")
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != "new_first_entry" || moved.Name != "new_first_entry" {
		t.Fatalf("%+v", moved)
	}
	if _, err := db.GetEntry[any]("first_entry"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
}

func TestUpdateEntry(t *testing.T) {
	db := openDB(t)
	_, err := db.UpdateEntry[any]("", nil)
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.UpdateEntry[any]("missing", nil)
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	created, err := db.CreateEntry("some_entry", 5)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := db.UpdateEntry("some_entry", 15)
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

func TestDuplicateEntry(t *testing.T) {
	db := openDB(t)
	src, err := db.CreateEntry("orig", map[string]int{"n": 3})
	if err != nil {
		t.Fatal(err)
	}
	dup, err := db.DuplicateEntry[map[string]int]("orig", "copy")
	if err != nil {
		t.Fatal(err)
	}
	if dup.ID != "copy" || dup.Name != "copy" || dup.Data["n"] != 3 {
		t.Fatalf("%+v", dup)
	}
	if dup.CreatedAt.Equal(src.CreatedAt) && dup.ID == src.ID {
		t.Fatal("expected a new object")
	}
	still, err := db.GetEntry[map[string]int]("orig")
	if err != nil {
		t.Fatal(err)
	}
	if still.Data["n"] != 3 {
		t.Fatal("source changed")
	}
	_, err = db.DuplicateEntry[any]("orig", "copy")
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("dup exist: %v", err)
	}
}

func TestRemoveEntry(t *testing.T) {
	db := openDB(t)
	if err := db.RemoveEntry(""); !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	if err := db.RemoveEntry("missing"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	if _, err := db.CreateEntry[any]("some_entry", nil); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveEntry("some_entry"); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveEntry("some_entry"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
}

func TestCreateEntryPretty(t *testing.T) {
	dir := t.TempDir()
	db := fsentry.New(dir, fsentry.WithNoLockFile(), fsentry.WithPretty())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("pretty", map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "pretty.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("\t")) {
		t.Fatalf("want tab indent: %s", raw)
	}
}
