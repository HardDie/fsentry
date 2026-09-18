package fsentry_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/HardDie/fsentry"
)

func TestCreateGetBinary(t *testing.T) {
	db := openDB(t)

	err := db.CreateBinary("", []byte("data"))
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("empty name: %v", err)
	}
	err = db.CreateBinary("bad_path", []byte("data"), "not_exist_folder")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("missing parent: %v", err)
	}

	payload := []byte("hello bin")
	if err := db.CreateBinary("My Image", payload); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetBinary("My Image", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("got %q", got)
	}
	byID, err := db.GetBinary("my_image", make([]byte, 0, len(payload)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(byID, payload) {
		t.Fatalf("got %q", byID)
	}

	if _, err := db.CreateFolder[any]("games", nil); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateBinary("inner", []byte("x"), "games"); err != nil {
		t.Fatal(err)
	}

	err = db.CreateBinary("My Image", payload)
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("dup: %v", err)
	}

	if err := db.CreateBinary("empty", nil); err != nil {
		t.Fatal(err)
	}
	empty, err := db.GetBinary("empty", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty %q", empty)
	}

	list, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Binaries) != 2 {
		t.Fatalf("binaries %v", list.Binaries)
	}
}

func TestGetBinaryErrors(t *testing.T) {
	db := openDB(t)
	_, err := db.GetBinary("", nil)
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	_, err = db.GetBinary("missing", nil)
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	_, err = db.GetBinary("x", nil, "missing_parent")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
}

func TestGetBinarySizedBuffer(t *testing.T) {
	db := openDB(t)
	payload := bytes.Repeat([]byte("a"), 256)
	if err := db.CreateBinary("buf", payload); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 0, len(payload))
	got, err := db.GetBinary("buf", buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("%q", got)
	}
}

func TestMoveBinary(t *testing.T) {
	db := openDB(t)
	err := db.MoveBinary("not_exist", "new")
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	err = db.MoveBinary("a", "b", "missing")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
	if err := db.CreateBinary("first_binary", []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateBinary("second_binary", []byte("b")); err != nil {
		t.Fatal(err)
	}
	err = db.MoveBinary("", "x")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	err = db.MoveBinary("first_binary", "")
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	err = db.MoveBinary("first_binary", "second_binary")
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("%v", err)
	}
	if err := db.MoveBinary("first_binary", "new_first_binary"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetBinary("first_binary", nil); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	got, err := db.GetBinary("new_first_binary", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "a" {
		t.Fatalf("%q", got)
	}
}

func TestUpdateBinary(t *testing.T) {
	db := openDB(t)
	err := db.UpdateBinary("", []byte("data"))
	if !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	err = db.UpdateBinary("missing", []byte("data"))
	if !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	if err := db.CreateBinary("some_binary", []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateBinary("some_binary", []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetBinary("some_binary", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("%q", got)
	}
	if err := db.UpdateBinary("some_binary", nil); err != nil {
		t.Fatal(err)
	}
	got, err = db.GetBinary("some_binary", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("%q", got)
	}
}

func TestRemoveBinary(t *testing.T) {
	db := openDB(t)
	if err := db.RemoveBinary(""); !errors.Is(err, fsentry.ErrBadName) {
		t.Fatalf("%v", err)
	}
	if err := db.RemoveBinary("missing"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
	if err := db.CreateBinary("some_binary", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveBinary("some_binary"); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveBinary("some_binary"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("%v", err)
	}
}

func TestBinaryOnDisk(t *testing.T) {
	dir := t.TempDir()
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateBinary("pic", []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "pic.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, []byte{1, 2, 3}) {
		t.Fatalf("%v", raw)
	}
}
