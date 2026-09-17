package lock

import (
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkLockUnlock(b *testing.B) {
	path := filepath.Join(b.TempDir(), FileName)
	f, err := Open(path, time.Minute)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = f.Close() })
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := f.Lock(); err != nil {
			b.Fatal(err)
		}
		if err := f.Unlock(); err != nil {
			b.Fatal(err)
		}
	}
}
