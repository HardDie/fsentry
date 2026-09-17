//go:build unix

package fs

import (
	"os"
	"syscall"
	"testing"
)

func BenchmarkMapErrorUnix(b *testing.B) {
	cases := []struct {
		name string
		err  error
	}{
		{"eexist", syscall.EEXIST},
		{"enoent", syscall.ENOENT},
		{"eacces", syscall.EACCES},
		{"eperm", syscall.EPERM},
		{"enotdir", syscall.ENOTDIR},
		{"eisdir", syscall.EISDIR},
		{"enospc", syscall.ENOSPC},
		{"edquot", syscall.EDQUOT},
		{"erofs", syscall.EROFS},
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
