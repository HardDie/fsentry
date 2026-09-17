package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestClose(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Close(file); err != nil {
			t.Fatal(err)
		}
		err = Close(file)
		if !errors.Is(err, ErrInternal) {
			t.Fatalf("second close: got %v, want ErrInternal", err)
		}
	})
}

func TestSync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = Close(file) })
		if err := Write(file, []byte("x")); err != nil {
			t.Fatal(err)
		}
		if err := Sync(file); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("closed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Close(file); err != nil {
			t.Fatal(err)
		}
		err = Sync(file)
		if !errors.Is(err, ErrInternal) {
			t.Fatalf("got %v, want ErrInternal", err)
		}
	})
}

func TestCloseDoesNotLeaveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.bin")
	file, err := CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Close(file); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
