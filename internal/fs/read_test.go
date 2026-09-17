package fs

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	payload := []byte("hello world")

	t.Run("sized buffer", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Write(file, payload); err != nil {
			t.Fatal(err)
		}
		if err := Close(file); err != nil {
			t.Fatal(err)
		}
		in, err := OpenRead(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = Close(in) })
		got, err := Read(in, make([]byte, 0, len(payload)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("grows", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Write(file, payload); err != nil {
			t.Fatal(err)
		}
		if err := Close(file); err != nil {
			t.Fatal(err)
		}
		in, err := OpenRead(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = Close(in) })
		got, err := Read(in, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.bin")
		file, err := CreateFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Close(file); err != nil {
			t.Fatal(err)
		}
		in, err := OpenRead(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = Close(in) })
		got, err := Read(in, make([]byte, 0, 8))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("got %q", got)
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
		_, err = Read(file, nil)
		if !errors.Is(err, ErrInternal) {
			t.Fatalf("got %v, want ErrInternal", err)
		}
	})
}

func TestReadSizedAllocs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.bin")
	payload := bytes.Repeat([]byte("a"), 256)
	file, err := CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Write(file, payload); err != nil {
		t.Fatal(err)
	}
	if err := Close(file); err != nil {
		t.Fatal(err)
	}
	in, err := OpenRead(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(in) })
	buf := make([]byte, 0, len(payload))
	n := testing.AllocsPerRun(100, func() {
		if _, err := in.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		got, err := Read(in, buf)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(payload) {
			t.Fatal(len(got))
		}
		buf = got[:0]
	})
	if n != 0 {
		t.Fatalf("Read sized allocs = %v, want 0", n)
	}
}
