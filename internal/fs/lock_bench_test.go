package fs

import (
	"path/filepath"
	"testing"
)

func BenchmarkLockUnlock(b *testing.B) {
	path := filepath.Join(b.TempDir(), ".fsentry.lock")
	file, err := OpenLock(path)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = Close(file) })
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := Lock(file); err != nil {
			b.Fatal(err)
		}
		if err := Unlock(file); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenLock(b *testing.B) {
	path := filepath.Join(b.TempDir(), ".fsentry.lock")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		file, err := OpenLock(path)
		if err != nil {
			b.Fatal(err)
		}
		if err := Close(file); err != nil {
			b.Fatal(err)
		}
	}
}
