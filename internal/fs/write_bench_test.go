package fs

import (
	"path/filepath"
	"testing"
)

func BenchmarkWrite(b *testing.B) {
	path := filepath.Join(b.TempDir(), "file.bin")
	file, err := CreateFile(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = file.Close() })
	buf := make([]byte, 256)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := file.Seek(0, 0); err != nil {
			b.Fatal(err)
		}
		if err := Write(file, buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteEmpty(b *testing.B) {
	path := filepath.Join(b.TempDir(), "file.bin")
	file, err := CreateFile(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = file.Close() })
	b.ReportAllocs()
	for b.Loop() {
		if err := Write(file, nil); err != nil {
			b.Fatal(err)
		}
	}
}
