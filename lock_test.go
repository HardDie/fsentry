package fsentry_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/HardDie/fsentry"
)

// Two handles on one root, lock file on: reads share the lock, a write is exclusive.
// A reader must see either the old or the new binary, never a torn one.
func TestReadsShareLockWritesExclusive(t *testing.T) {
	dir := t.TempDir()
	open := func() *fsentry.DB {
		db := fsentry.New(dir)
		if err := db.Init(); err != nil {
			t.Fatal(err)
		}
		return db
	}
	writer, reader := open(), open()
	t.Cleanup(func() { _ = writer.Drop() })

	if _, err := writer.CreateFolder("cards", struct{}{}); err != nil {
		t.Fatal(err)
	}
	oldBody := bytes.Repeat([]byte{'a'}, 1<<16)
	newBody := bytes.Repeat([]byte{'b'}, 1<<16)
	if err := writer.CreateBinary("1", oldBody, "cards"); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, db := range []*fsentry.DB{reader, reader, writer} {
		wg.Go(func() {
			for range 50 {
				got, err := db.GetBinary("1", nil, "cards")
				if err != nil {
					t.Error(err)
					return
				}
				if !bytes.Equal(got, oldBody) && !bytes.Equal(got, newBody) {
					t.Error("torn read")
					return
				}
				if _, err := db.List("cards"); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Go(func() {
		for i := range 20 {
			body := oldBody
			if i%2 == 0 {
				body = newBody
			}
			if err := writer.UpdateBinary("1", body, "cards"); err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Wait()
}
