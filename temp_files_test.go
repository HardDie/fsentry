package fsentry_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/HardDie/fsentry"
)

// staleTemps is a store with one folder, binary, and entry,
// plus the temp files a crash during their update would leave.
func staleTemps(t *testing.T) (*fsentry.DB, string, []string) {
	t.Helper()
	db, dir := openDir(t)
	if _, err := db.CreateFolder("cards", map[string]string{}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateBinary("1", []byte("card"), "cards"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("notes", map[string]string{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	temps := []string{
		filepath.Join(dir, "cards", "1.bin.tmp"),
		filepath.Join(dir, "cards", ".info.json.tmp"),
		filepath.Join(dir, "notes.json.tmp"),
	}
	for _, p := range temps {
		if err := os.WriteFile(p, []byte("half written"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return db, dir, temps
}

func TestTempFilesNotListed(t *testing.T) {
	db, _, _ := staleTemps(t)
	root, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(root.Entries, []string{"notes"}) || !slices.Equal(root.Folders, []string{"cards"}) || len(root.Binaries) != 0 {
		t.Fatalf("root list %+v", root)
	}
	cards, err := db.List("cards")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cards.Binaries, []string{"1"}) || len(cards.Entries) != 0 || len(cards.Folders) != 0 {
		t.Fatalf("cards list %+v", cards)
	}
}

// Names map to IDs without dots, so no Get can reach a temp file.
func TestTempFilesNotReadable(t *testing.T) {
	db, _, _ := staleTemps(t)
	for _, name := range []string{"1.bin.tmp", "1.bin", "1.tmp", ".tmp"} {
		if _, err := db.GetBinary(name, nil, "cards"); !errors.Is(err, fsentry.ErrNotExist) {
			t.Errorf("GetBinary(%q) = %v, want ErrNotExist", name, err)
		}
	}
	for _, name := range []string{"notes.json.tmp", "notes.json"} {
		if _, err := db.GetEntry[map[string]string](name); !errors.Is(err, fsentry.ErrNotExist) {
			t.Errorf("GetEntry(%q) = %v, want ErrNotExist", name, err)
		}
	}
	if got, err := db.GetBinary("1", nil, "cards"); err != nil || string(got) != "card" {
		t.Fatalf("real binary: %q, %v", got, err)
	}
}

// A create with a temp-like name makes its own dot-free file and leaves the temp alone.
func TestCreateCannotTouchTempFiles(t *testing.T) {
	tests := []struct {
		name   string
		create func(*fsentry.DB) error
		file   string // what the create must write, relative to the root
	}{
		{"binary 1.bin.tmp", func(db *fsentry.DB) error { return db.CreateBinary("1.bin.tmp", []byte("x"), "cards") }, "cards/1bintmp.bin"},
		{"binary 1.bin", func(db *fsentry.DB) error { return db.CreateBinary("1.bin", []byte("x"), "cards") }, "cards/1bin.bin"},
		{"binary .tmp", func(db *fsentry.DB) error { return db.CreateBinary(".tmp", []byte("x"), "cards") }, "cards/tmp.bin"},
		{"entry notes.json.tmp", func(db *fsentry.DB) error {
			_, err := db.CreateEntry("notes.json.tmp", map[string]string{})
			return err
		}, "notesjsontmp.json"},
		{"folder cards.tmp", func(db *fsentry.DB) error {
			_, err := db.CreateFolder("cards.tmp", map[string]string{})
			return err
		}, "cardstmp/.info.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dir, temps := staleTemps(t)
			if err := tt.create(db); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(tt.file))); err != nil {
				t.Fatalf("want %s: %v", tt.file, err)
			}
			for _, p := range temps {
				if body, err := os.ReadFile(p); err != nil || string(body) != "half written" {
					t.Fatalf("temp %s changed: %q, %v", p, body, err)
				}
			}
		})
	}
}

func zipNames(t *testing.T, data []byte) []string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}

func TestExportSkipsTempFiles(t *testing.T) {
	db, _, _ := staleTemps(t)
	var root, folder bytes.Buffer
	if err := db.Export(&root); err != nil {
		t.Fatal(err)
	}
	if err := db.ExportFolder(&folder, "cards"); err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{root.Bytes(), folder.Bytes()} {
		names := zipNames(t, data)
		for _, n := range names {
			if strings.HasSuffix(n, ".tmp") {
				t.Fatalf("temp file in archive: %q (all: %v)", n, names)
			}
		}
		if !slices.ContainsFunc(names, func(n string) bool { return strings.HasSuffix(n, "cards/1.bin") }) {
			t.Fatalf("real binary missing from archive: %v", names)
		}
	}
}

func TestImportSkipsTempFiles(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range map[string]string{"keep.bin": "keep", "keep.bin.tmp": "half written"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	db, dir := openDir(t)
	if err := db.Import(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.bin.tmp")); !os.IsNotExist(err) {
		t.Fatalf("temp file imported: %v", err)
	}
	if got, err := db.GetBinary("keep", nil); err != nil || string(got) != "keep" {
		t.Fatalf("real binary: %q, %v", got, err)
	}
}
