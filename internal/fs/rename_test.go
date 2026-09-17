package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRenameFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		oldpath := filepath.Join(dir, "old.bin")
		newpath := filepath.Join(dir, "new.bin")
		if err := os.WriteFile(oldpath, []byte("data"), 0o666); err != nil {
			t.Fatal(err)
		}
		if err := RenameFile(oldpath, newpath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(oldpath); !os.IsNotExist(err) {
			t.Fatalf("old path still exists: %v", err)
		}
		got, err := os.ReadFile(newpath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "data" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("source missing", func(t *testing.T) {
		dir := t.TempDir()
		err := RenameFile(filepath.Join(dir, "missing.bin"), filepath.Join(dir, "new.bin"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("dest parent missing", func(t *testing.T) {
		dir := t.TempDir()
		oldpath := filepath.Join(dir, "old.bin")
		if err := os.WriteFile(oldpath, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		err := RenameFile(oldpath, filepath.Join(dir, "missing", "new.bin"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("dest parent is file", func(t *testing.T) {
		dir := t.TempDir()
		oldpath := filepath.Join(dir, "old.bin")
		parent := filepath.Join(dir, "notdir")
		if err := os.WriteFile(oldpath, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(parent, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		err := RenameFile(oldpath, filepath.Join(parent, "new.bin"))
		assertParentNotDir(t, err)
	})
}

func TestRenameFolder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		oldpath := filepath.Join(dir, "old")
		newpath := filepath.Join(dir, "new")
		if err := os.Mkdir(oldpath, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(oldpath, "inner.bin"), []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		if err := RenameFolder(oldpath, newpath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(oldpath); !os.IsNotExist(err) {
			t.Fatalf("old path still exists: %v", err)
		}
		got, err := os.ReadFile(filepath.Join(newpath, "inner.bin"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "x" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("source missing", func(t *testing.T) {
		dir := t.TempDir()
		err := RenameFolder(filepath.Join(dir, "missing"), filepath.Join(dir, "new"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("dest parent missing", func(t *testing.T) {
		dir := t.TempDir()
		oldpath := filepath.Join(dir, "old")
		if err := os.Mkdir(oldpath, 0o755); err != nil {
			t.Fatal(err)
		}
		err := RenameFolder(oldpath, filepath.Join(dir, "missing", "new"))
		if !errors.Is(err, ErrNotExist) {
			t.Fatalf("got %v, want ErrNotExist", err)
		}
	})

	t.Run("dest exists nonempty", func(t *testing.T) {
		dir := t.TempDir()
		oldpath := filepath.Join(dir, "old")
		newpath := filepath.Join(dir, "new")
		if err := os.Mkdir(oldpath, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(newpath, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(newpath, "keep.bin"), []byte("y"), 0o666); err != nil {
			t.Fatal(err)
		}
		err := RenameFolder(oldpath, newpath)
		if !errors.Is(err, ErrExist) && !errors.Is(err, ErrInternal) {
			// POSIX: ENOTEMPTY -> ErrExist. Some systems report a different errno.
			t.Fatalf("got %v, want ErrExist", err)
		}
	})
}
