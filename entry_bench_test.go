package fsentry

import (
	"strconv"
	"testing"
)

func BenchmarkCreateEntry(b *testing.B) {
	db := New(b.TempDir(), WithNoLockFile())
	if err := db.Init(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		if _, err := db.CreateEntry[any](strconv.Itoa(i), nil); err != nil {
			b.Fatal(err)
		}
		i++
	}
}
