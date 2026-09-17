package fs

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkRemoveFile(b *testing.B) {
	dir := b.TempDir()
	paths := make([]string, b.N)
	for i := range paths {
		paths[i] = filepath.Join(dir, strconv.Itoa(i))
		if err := os.WriteFile(paths[i], []byte("x"), 0o666); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := RemoveFile(paths[i]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRemoveFolder(b *testing.B) {
	dir := b.TempDir()
	paths := make([]string, b.N)
	for i := range paths {
		paths[i] = filepath.Join(dir, strconv.Itoa(i))
		if err := os.Mkdir(paths[i], 0o755); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := RemoveFolder(paths[i]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadDir(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 8; i++ {
		if err := os.WriteFile(filepath.Join(dir, strconv.Itoa(i)), []byte("x"), 0o666); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := ReadDir(dir); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStat(b *testing.B) {
	path := filepath.Join(b.TempDir(), "f.bin")
	if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := Stat(path); err != nil {
			b.Fatal(err)
		}
	}
}
