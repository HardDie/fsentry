//go:build windows

package fs

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestMapErrorWindows(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"exist", os.ErrExist, ErrExist},
		{"not exist", os.ErrNotExist, ErrNotExist},
		{"permission", os.ErrPermission, ErrPermission},
		{"already exists", syscall.ERROR_ALREADY_EXISTS, ErrExist},
		{"file exists", syscall.ERROR_FILE_EXISTS, ErrExist},
		{"file not found", syscall.ERROR_FILE_NOT_FOUND, ErrNotExist},
		{"path not found", syscall.ERROR_PATH_NOT_FOUND, ErrNotExist},
		{"access denied", syscall.ERROR_ACCESS_DENIED, ErrPermission},
		{"directory", syscall.ERROR_DIRECTORY, ErrNotDirectory},
		{"write protect", syscall.ERROR_WRITE_PROTECT, ErrReadOnly},
		{"disk full", syscall.ERROR_DISK_FULL, ErrNoSpace},
		{"handle disk full", syscall.ERROR_HANDLE_DISK_FULL, ErrNoSpace},
		{"unknown", errors.New("cosmic ray"), ErrInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapError(&os.PathError{Op: "open", Path: "x", Err: tc.err})
			if !errors.Is(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if got != tc.want {
				t.Fatalf("want the sentinel itself, got %v", got)
			}
		})
	}
}

func TestMapErrorNil(t *testing.T) {
	if mapError(nil) != nil {
		t.Fatal("nil should map to nil")
	}
}
