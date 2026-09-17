package fs

import (
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkCreateFile(b *testing.B) {
	dir := b.TempDir()
	paths := make([]string, b.N)
	for i := range paths {
		paths[i] = filepath.Join(dir, strconv.Itoa(i))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := CreateFile(paths[i])
		if err != nil {
			b.Fatal(err)
		}
		if err := file.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateFolder(b *testing.B) {
	dir := b.TempDir()
	paths := make([]string, b.N)
	for i := range paths {
		paths[i] = filepath.Join(dir, strconv.Itoa(i))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := CreateFolder(paths[i]); err != nil {
			b.Fatal(err)
		}
	}
}
