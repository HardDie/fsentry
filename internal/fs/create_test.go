package fs

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCreateFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = file.Close() })
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.bin")
		if err := os.WriteFile(path, nil, 0o666); err != nil {
			t.Fatal(err)
		}
		file, err := CreateFile(path)
		if file != nil {
			t.Fatal("expected nil file")
		}
		if !errors.Is(err, ErrExist) {
			t.Fatalf("got %v, want ErrExist", err)
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "file.bin")
		_, err := CreateFile(path)
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
		_, err := CreateFile(filepath.Join(parent, "file.bin"))
		assertParentNotDir(t, err)
	})

	t.Run("path is directory", func(t *testing.T) {
		dir := t.TempDir()
		_, err := CreateFile(dir)
		if errors.Is(err, ErrIsDirectory) || errors.Is(err, ErrExist) || errors.Is(err, ErrPermission) {
			return
		}
		t.Fatalf("got %v, want ErrIsDirectory, ErrExist, or ErrPermission", err)
	})

	t.Run("permission", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows ACL model is not POSIX chmod")
		}
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory write bits")
		}
		parent := filepath.Join(t.TempDir(), "ro")
		if err := os.Mkdir(parent, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
		_, err := CreateFile(filepath.Join(parent, "file.bin"))
		if !errors.Is(err, ErrPermission) {
			t.Fatalf("got %v, want ErrPermission", err)
		}
	})
}

func TestCreateFolder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "dir")
		if err := CreateFolder(path); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Fatal("expected directory")
		}
	})

	t.Run("exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "dir")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		err := CreateFolder(path)
		if !errors.Is(err, ErrExist) {
			t.Fatalf("got %v, want ErrExist", err)
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "dir")
		err := CreateFolder(path)
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
		err := CreateFolder(filepath.Join(parent, "dir"))
		assertParentNotDir(t, err)
	})

	t.Run("path is file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file")
		if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		err := CreateFolder(path)
		if !errors.Is(err, ErrExist) {
			t.Fatalf("got %v, want ErrExist", err)
		}
	})

	t.Run("permission", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows ACL model is not POSIX chmod")
		}
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory write bits")
		}
		parent := filepath.Join(t.TempDir(), "ro")
		if err := os.Mkdir(parent, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
		err := CreateFolder(filepath.Join(parent, "dir"))
		if !errors.Is(err, ErrPermission) {
			t.Fatalf("got %v, want ErrPermission", err)
		}
	})
}

func TestCreateFolderAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c")
	if err := CreateFolderAll(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatalf("%v", err)
	}
	if err := CreateFolderAll(path); err != nil {
		t.Fatal(err)
	}
}

func TestCreateFileDoesNotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.bin")
	file, err := CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("new file has %d bytes", info.Size())
	}
}

func assertParentNotDir(t *testing.T, err error) {
	t.Helper()
	if errors.Is(err, ErrNotDirectory) {
		return
	}
	// Windows reports ERROR_PATH_NOT_FOUND when a path component is a file.
	if runtime.GOOS == "windows" && errors.Is(err, ErrNotExist) {
		return
	}
	t.Fatalf("got %v, want ErrNotDirectory", err)
}
