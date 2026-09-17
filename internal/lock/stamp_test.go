package lock

import (
	"path/filepath"
	"testing"

	"github.com/HardDie/fsentry/internal/fs"
)

func TestStampRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	file, err := fs.OpenLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fs.Close(file) })

	_, ok, err := readStamp(file)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("empty lock file should have no stamp")
	}

	const nano = int64(123456789)
	if err := writeStamp(file, nano); err != nil {
		t.Fatal(err)
	}
	got, ok, err := readStamp(file)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got != nano {
		t.Fatalf("got %d ok=%v, want %d", got, ok, nano)
	}
}
