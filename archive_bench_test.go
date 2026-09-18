package fsentry_test

import (
	"bytes"
	"testing"

	"github.com/HardDie/fsentry"
)

func BenchmarkExport(b *testing.B) {
	db := fsentry.New(b.TempDir(), fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		b.Fatal(err)
	}
	if _, err := db.CreateFolder("n", struct{}{}); err != nil {
		b.Fatal(err)
	}
	if err := db.CreateBinary("b", []byte("hello"), "n"); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var buf bytes.Buffer
		if err := db.Export(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImport(b *testing.B) {
	src := fsentry.New(b.TempDir(), fsentry.WithNoLockFile())
	if err := src.Init(); err != nil {
		b.Fatal(err)
	}
	if _, err := src.CreateEntry("settings", struct{}{}); err != nil {
		b.Fatal(err)
	}
	var zipBuf bytes.Buffer
	if err := src.Export(&zipBuf); err != nil {
		b.Fatal(err)
	}
	data := zipBuf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		db := fsentry.New(b.TempDir(), fsentry.WithNoLockFile())
		if err := db.Init(); err != nil {
			b.Fatal(err)
		}
		if err := db.Import(bytes.NewReader(data)); err != nil {
			b.Fatal(err)
		}
	}
}
