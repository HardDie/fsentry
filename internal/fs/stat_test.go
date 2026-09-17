package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStat(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		if err := os.WriteFile(path, []byte("hi"), 0o666); err != nil {
			t.Fatal(err)
		}
		info, err := Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			t.Fatal("expected file")
		}
		if info.Size() != 2 {
			t.Fatalf("size %d", info.Size())
		}
	})

	t.Run("dir", func(t *testing.T) {
		dir := t.TempDir()
		info, err := Stat(dir)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Fatal("expected dir")
		}
	})

	t.Run("missing", func(t *testing.T) {
		_, err := Stat(filepath.Join(t.TempDir(), "missing"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})
}
