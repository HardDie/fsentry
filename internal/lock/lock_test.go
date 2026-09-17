package lock

import (
	"path/filepath"
	"testing"
	"time"
)

func TestOpenDefaultTimeout(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	f, err := Open(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	if f.timeout != DefaultTimeout {
		t.Fatalf("timeout %v, want %v", f.timeout, DefaultTimeout)
	}
}

func TestLockUnlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	f, err := Open(path, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	if err := f.Lock(); err != nil {
		t.Fatal(err)
	}
	if err := f.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestLockWaitsForLiveHolder(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	a, err := Open(path, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	b, err := Open(path, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })

	if err := a.Lock(); err != nil {
		t.Fatal(err)
	}

	got := make(chan error, 1)
	go func() {
		got <- b.Lock()
	}()

	select {
	case err := <-got:
		t.Fatalf("second lock returned before unlock: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	if err := a.Unlock(); err != nil {
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

	if err := b.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestLockStealsStale(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	a, err := Open(path, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	b, err := Open(path, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })

	if err := a.Lock(); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	b.now = func() time.Time { return start.Add(time.Hour) }

	done := make(chan error, 1)
	go func() {
		done <- b.Lock()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not steal stale lock")
	}

	if err := b.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestLockStealsAfterTimeoutElapses(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	a, err := Open(path, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	b, err := Open(path, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })

	if err := a.Lock(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- b.Lock()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not steal after timeout")
	}

	if err := b.Unlock(); err != nil {
		t.Fatal(err)
	}
}
