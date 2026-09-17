package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWrite(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		payload := []byte("hello")
		if err := Write(file, payload); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(payload) {
			t.Fatalf("got %q, want %q", got, payload)
		}
	})

	t.Run("empty", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = file.Close() })
		if err := Write(file, nil); err != nil {
			t.Fatal(err)
		}
		info, err := file.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != 0 {
			t.Fatalf("size %d", info.Size())
		}
	})

	t.Run("closed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "closed.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		err = Write(file, []byte("x"))
		if !errors.Is(err, ErrInternal) {
			t.Fatalf("got %v, want ErrInternal", err)
		}
	})
}

func TestWriteEmptyAllocs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.bin")
	file, err := CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	n := testing.AllocsPerRun(1000, func() {
		if err := Write(file, nil); err != nil {
			t.Fatal(err)
		}
	})
	if n != 0 {
		t.Fatalf("Write(nil) allocs = %v, want 0", n)
	}
}
