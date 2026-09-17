package fs

import (
	"errors"
	"os"
	"testing"
)

func BenchmarkWrap(b *testing.B) {
	err := &os.PathError{Op: "open", Path: "x", Err: os.ErrExist}
	b.ReportAllocs()
	for b.Loop() {
		_ = wrap(ErrExist, err)
	}
}

func BenchmarkMapError(b *testing.B) {
	cases := []struct {
		name string
		err  error
	}{
		{"nil", nil},
		{"exist", &os.PathError{Op: "open", Path: "x", Err: os.ErrExist}},
		{"not_exist", &os.PathError{Op: "open", Path: "x", Err: os.ErrNotExist}},
		{"permission", &os.PathError{Op: "open", Path: "x", Err: os.ErrPermission}},
		{"unknown", &os.PathError{Op: "open", Path: "x", Err: errors.New("cosmic ray")}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			err := tc.err
			b.ReportAllocs()
			for b.Loop() {
				_ = mapError(err)
			}
		})
	}
}
