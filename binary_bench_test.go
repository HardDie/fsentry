package fsentry

import (
	"strconv"
	"testing"
)

func BenchmarkCreateBinary(b *testing.B) {
	db := New(b.TempDir(), WithNoLockFile())
	if err := db.Init(); err != nil {
		b.Fatal(err)
	}
	payload := []byte("x")
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		if err := db.CreateBinary(strconv.Itoa(i), payload); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkGetBinary(b *testing.B) {
	db := New(b.TempDir(), WithNoLockFile())
	if err := db.Init(); err != nil {
		b.Fatal(err)
	}
	payload := make([]byte, 256)
	if err := db.CreateBinary("buf", payload); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 0, len(payload))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := db.GetBinary("buf", buf)
		if err != nil {
			b.Fatal(err)
		}
		buf = got[:0]
	}
}

func BenchmarkUpdateBinary(b *testing.B) {
	db := New(b.TempDir(), WithNoLockFile())
	if err := db.Init(); err != nil {
		b.Fatal(err)
	}
	payload := []byte("hello")
	if err := db.CreateBinary("u", payload); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := db.UpdateBinary("u", payload); err != nil {
			b.Fatal(err)
		}
	}
}
