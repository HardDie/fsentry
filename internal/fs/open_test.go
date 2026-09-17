package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		if err := os.WriteFile(path, []byte("hi"), 0o666); err != nil {
			t.Fatal(err)
		}
		file, err := OpenRead(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = Close(file) })
	})

	t.Run("missing", func(t *testing.T) {
		_, err := OpenRead(filepath.Join(t.TempDir(), "missing.bin"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("parent is file", func(t *testing.T) {
		dir := t.TempDir()
		parent := filepath.Join(dir, "notdir")
		if err := os.WriteFile(parent, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		_, err := OpenRead(filepath.Join(parent, "f.bin"))
		assertParentNotDir(t, err)
	})
}

func TestOpenWrite(t *testing.T) {
	t.Run("success truncates", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		if err := os.WriteFile(path, []byte("hello"), 0o666); err != nil {
			t.Fatal(err)
		}
		file, err := OpenWrite(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Write(file, []byte("x")); err != nil {
			t.Fatal(err)
		}
		if err := Close(file); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "x" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("missing", func(t *testing.T) {
		_, err := OpenWrite(filepath.Join(t.TempDir(), "missing.bin"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("path is directory", func(t *testing.T) {
		dir := t.TempDir()
		_, err := OpenWrite(dir)
		if !errors.Is(err, ErrIsDirectory) && !errors.Is(err, ErrPermission) {
			t.Fatalf("got %v, want ErrIsDirectory", err)
		}
	})
}
