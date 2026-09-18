package fsentry

import "testing"

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if New("root", WithNoLockFile()) == nil {
			b.Fatal("nil")
		}
	}
}
