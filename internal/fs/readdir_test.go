package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "a.bin"), []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		entries, err := ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]bool{}
		for _, e := range entries {
			got[e.Name()] = e.IsDir()
		}
		isDir, ok := got["a.bin"]
		if !ok {
			t.Fatalf("missing a.bin: %v", got)
		}
		if isDir {
			t.Fatal("a.bin should be a file")
		}
		if !got["sub"] {
			t.Fatalf("missing sub dir: %v", got)
		}
	})

	t.Run("missing", func(t *testing.T) {
		_, err := ReadDir(filepath.Join(t.TempDir(), "missing"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		_, err := ReadDir(path)
		if !errors.Is(err, ErrNotDirectory) {
			t.Fatalf("got %v, want ErrNotDirectory", err)
		}
	})
}
