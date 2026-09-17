package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkRenameFile(b *testing.B) {
	dir := b.TempDir()
	oldpath := filepath.Join(dir, "a.bin")
	newpath := filepath.Join(dir, "b.bin")
	if err := os.WriteFile(oldpath, []byte("x"), 0o666); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	src, dst := oldpath, newpath
	for b.Loop() {
		if err := RenameFile(src, dst); err != nil {
			b.Fatal(err)
		}
		src, dst = dst, src
	}
}

func BenchmarkRenameFolder(b *testing.B) {
	dir := b.TempDir()
	oldpath := filepath.Join(dir, "a")
	newpath := filepath.Join(dir, "b")
	if err := os.Mkdir(oldpath, 0o755); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	src, dst := oldpath, newpath
	for b.Loop() {
		if err := RenameFolder(src, dst); err != nil {
			b.Fatal(err)
		}
		src, dst = dst, src
	}
}
