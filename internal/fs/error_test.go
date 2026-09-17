package fs

import (
	"errors"
	"os"
	"testing"
)

func TestMapErrorNilAllocs(t *testing.T) {
	n := testing.AllocsPerRun(1000, func() {
		_ = mapError(nil)
	})
	if n != 0 {
		t.Fatalf("mapError(nil) allocs = %v, want 0", n)
	}
}

func TestWrap(t *testing.T) {
	orig := &os.PathError{Op: "open", Path: "x", Err: os.ErrExist}
	got := wrap(ErrExist, orig)
	if got != ErrExist {
		t.Fatalf("got %v, want sentinel ErrExist", got)
	}
	if errors.Is(got, orig) {
		t.Fatal("OS error must not be chained")
	}
}

func TestWrapAllocs(t *testing.T) {
	orig := &os.PathError{Op: "open", Path: "x", Err: os.ErrExist}
	n := testing.AllocsPerRun(1000, func() {
		_ = wrap(ErrExist, orig)
	})
	if n != 0 {
		t.Fatalf("wrap allocs = %v, want 0", n)
	}
}

func TestMapErrorExistAllocs(t *testing.T) {
	err := &os.PathError{Op: "open", Path: "x", Err: os.ErrExist}
	n := testing.AllocsPerRun(1000, func() {
		_ = mapError(err)
	})
	if n != 0 {
		t.Fatalf("mapError(exist) allocs = %v, want 0", n)
	}
}
