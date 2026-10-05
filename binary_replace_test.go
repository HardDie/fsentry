package fsentry_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/HardDie/fsentry"
)

// openDir is a store whose root path the test needs.
func openDir(t *testing.T) (*fsentry.DB, string) {
	t.Helper()
	dir := t.TempDir()
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	return db, dir
}

// UpdateBinary writes a new file and renames it into place; it never rewrites the old one.
func TestUpdateBinaryRenamesIntoPlace(t *testing.T) {
	db, dir := openDir(t)
	if err := db.CreateBinary("cover", []byte("old")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "cover.bin")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// On Windows os.Stat loads the file ID lazily, on the first SameFile call.
	// Load it now, or it would be read from the new file after the rename.
	os.SameFile(before, before)
	if err := db.UpdateBinary("cover", []byte("new")); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(before, after) {
		t.Fatal("file was rewritten in place, want a new file renamed over it")
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file left behind: %v", err)
	}
	got, err := db.GetBinary("cover", nil)
	if err != nil || string(got) != "new" {
		t.Fatalf("got %q, %v", got, err)
	}
}

// A temp file left by a crash does not block the next update and is cleared.
func TestUpdateBinaryClearsStaleTemp(t *testing.T) {
	db, dir := openDir(t)
	if err := db.CreateBinary("cover", []byte("old")); err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(dir, "cover.bin.tmp")
	if err := os.WriteFile(tmp, []byte("half written"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateBinary("cover", []byte("new")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatalf("stale temp file still there: %v", err)
	}
	got, err := db.GetBinary("cover", nil)
	if err != nil || string(got) != "new" {
		t.Fatalf("got %q, %v", got, err)
	}
}

// If the new bytes cannot be written, the old file stays as it was.
func TestUpdateBinaryFailureKeepsOld(t *testing.T) {
	db, dir := openDir(t)
	if err := db.CreateBinary("cover", []byte("old")); err != nil {
		t.Fatal(err)
	}
	// A non-empty folder where the temp file goes: it can be neither removed nor created.
	tmp := filepath.Join(dir, "cover.bin.tmp")
	if err := os.MkdirAll(filepath.Join(tmp, "blocker"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := db.UpdateBinary("cover", []byte("new"))
	if err == nil {
		t.Fatal("want an error when the temp file cannot be created")
	}
	if errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("got %v, want a write error", err)
	}
	got, err := db.GetBinary("cover", nil)
	if err != nil || string(got) != "old" {
		t.Fatalf("old content lost: got %q, %v", got, err)
	}
}
