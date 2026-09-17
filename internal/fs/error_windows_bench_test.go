//go:build windows

package fs

import (
	"os"
	"syscall"
	"testing"
)

func BenchmarkMapErrorWindows(b *testing.B) {
	cases := []struct {
		name string
		err  error
	}{
		{"already_exists", syscall.ERROR_ALREADY_EXISTS},
		{"file_exists", syscall.ERROR_FILE_EXISTS},
		{"file_not_found", syscall.ERROR_FILE_NOT_FOUND},
		{"path_not_found", syscall.ERROR_PATH_NOT_FOUND},
		{"access_denied", syscall.ERROR_ACCESS_DENIED},
		{"directory", syscall.ERROR_DIRECTORY},
		{"write_protect", syscall.ERROR_WRITE_PROTECT},
		{"disk_full", syscall.ERROR_DISK_FULL},
		{"handle_disk_full", syscall.ERROR_HANDLE_DISK_FULL},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			err := &os.PathError{Op: "open", Path: "x", Err: tc.err}
			b.ReportAllocs()
			for b.Loop() {
				_ = mapError(err)
			}
		})
	}
}
