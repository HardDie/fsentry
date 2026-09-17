package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenLock(t *testing.T) {
	t.Run("creates", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".fsentry.lock")
		file, err := OpenLock(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = Close(file) })
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("opens existing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".fsentry.lock")
		first, err := OpenLock(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Close(first); err != nil {
			t.Fatal(err)
		}
		second, err := OpenLock(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Close(second); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("parent missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", ".fsentry.lock")
		_, err := OpenLock(path)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestLockUnlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".fsentry.lock")
	file, err := OpenLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(file) })
	if err := Lock(file); err != nil {
		t.Fatal(err)
	}
	if err := Unlock(file); err != nil {
		t.Fatal(err)
	}
	if err := Lock(file); err != nil {
		t.Fatal(err)
	}
	if err := Unlock(file); err != nil {
		t.Fatal(err)
	}
}

func TestLockBlocksSecondHandle(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".fsentry.lock")
	a, err := OpenLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(a) })
	b, err := OpenLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(b) })

	if err := Lock(a); err != nil {
		t.Fatal(err)
	}

	got := make(chan error, 1)
	go func() {
		got <- Lock(b)
	}()

	select {
	case err := <-got:
		t.Fatalf("second lock returned before unlock: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	if err := Unlock(a); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-got:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second lock did not acquire after unlock")
	}

	if err := Unlock(b); err != nil {
		t.Fatal(err)
	}
}

func TestTryLockBusy(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".fsentry.lock")
	a, err := OpenLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(a) })
	b, err := OpenLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(b) })

	if err := TryLock(a); err != nil {
		t.Fatal(err)
	}
	err = TryLock(b)
	if !errors.Is(err, ErrBusy) {
		t.Fatalf("got %v, want ErrBusy", err)
	}
	if err := Unlock(a); err != nil {
		t.Fatal(err)
	}
	if err := TryLock(b); err != nil {
		t.Fatal(err)
	}
	if err := Unlock(b); err != nil {
		t.Fatal(err)
	}
}
