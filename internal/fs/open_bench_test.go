package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkOpenRead(b *testing.B) {
	path := filepath.Join(b.TempDir(), "f.bin")
	if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		file, err := OpenRead(path)
		if err != nil {
			b.Fatal(err)
		}
		if err := Close(file); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenWrite(b *testing.B) {
	path := filepath.Join(b.TempDir(), "f.bin")
	if err := os.WriteFile(path, []byte("hello"), 0o666); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		file, err := OpenWrite(path)
		if err != nil {
			b.Fatal(err)
		}
		if err := Close(file); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRead(b *testing.B) {
	path := filepath.Join(b.TempDir(), "f.bin")
	payload := make([]byte, 256)
	if err := os.WriteFile(path, payload, 0o666); err != nil {
		b.Fatal(err)
	}
	file, err := OpenRead(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = Close(file) })
	buf := make([]byte, 0, len(payload))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := file.Seek(0, 0); err != nil {
			b.Fatal(err)
		}
		got, err := Read(file, buf)
		if err != nil {
			b.Fatal(err)
		}
		buf = got[:0]
	}
}

func BenchmarkClose(b *testing.B) {
	// Close once per open: holding b.N FDs hits the process limit.
	path := filepath.Join(b.TempDir(), "f.bin")
	file, err := CreateFile(path)
	if err != nil {
		b.Fatal(err)
	}
	if err := Close(file); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		file, err := OpenRead(path)
		if err != nil {
			b.Fatal(err)
		}
		if err := Close(file); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSync(b *testing.B) {
	path := filepath.Join(b.TempDir(), "f.bin")
	file, err := CreateFile(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = Close(file) })
	if err := Write(file, []byte("x")); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := Sync(file); err != nil {
			b.Fatal(err)
		}
	}
}
