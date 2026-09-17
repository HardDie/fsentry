//go:build unix

package fs

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestMapErrorUnix(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"exist", os.ErrExist, ErrExist},
		{"not exist", os.ErrNotExist, ErrNotExist},
		{"permission", os.ErrPermission, ErrPermission},
		{"eexist", syscall.EEXIST, ErrExist},
		{"enoent", syscall.ENOENT, ErrNotExist},
		{"eacces", syscall.EACCES, ErrPermission},
		{"eperm", syscall.EPERM, ErrPermission},
		{"enotdir", syscall.ENOTDIR, ErrNotDirectory},
		{"eisdir", syscall.EISDIR, ErrIsDirectory},
		{"enospc", syscall.ENOSPC, ErrNoSpace},
		{"edquot", syscall.EDQUOT, ErrNoSpace},
		{"erofs", syscall.EROFS, ErrReadOnly},
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
