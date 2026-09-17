package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		if err := RemoveFile(path); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("file still exists: %v", err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		err := RemoveFile(filepath.Join(t.TempDir(), "missing.bin"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "d")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		err := RemoveFile(path)
		if !errors.Is(err, ErrIsDirectory) && !errors.Is(err, ErrPermission) {
			t.Fatalf("got %v, want ErrIsDirectory", err)
		}
	})
}

func TestRemoveFolder(t *testing.T) {
	t.Run("success recursive", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "root")
		if err := os.Mkdir(root, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "a.bin"), []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		inner := filepath.Join(root, "inner")
		if err := os.Mkdir(inner, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := RemoveFolder(root); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Fatalf("folder still exists: %v", err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		err := RemoveFolder(filepath.Join(t.TempDir(), "missing"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		err := RemoveFolder(path)
		if !errors.Is(err, ErrNotDirectory) {
			t.Fatalf("got %v, want ErrNotDirectory", err)
		}
	})
}
