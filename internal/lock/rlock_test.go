package lock

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/HardDie/fsentry/internal/fs"
)

// openPair opens two handles on one lock file, like two processes.
func openPair(t *testing.T, timeout time.Duration) (*File, *File) {
	t.Helper()
	path := filepath.Join(t.TempDir(), FileName)
	a, err := Open(path, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	b, err := Open(path, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })
	return a, b
}

// waitBlocked fails if got delivers within 50ms, then expects it within 2s after release.
func waitBlocked(t *testing.T, got <-chan error, release func() error) {
	t.Helper()
	select {
	case err := <-got:
		t.Fatalf("returned before release: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-got:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not acquire after release")
	}
}

func TestRLockSharedAcrossHandles(t *testing.T) {
	a, b := openPair(t, time.Minute)
	if err := a.RLock(); err != nil {
		t.Fatal(err)
	}
	got := make(chan error, 1)
	go func() { got <- b.RLock() }()
	select {
	case err := <-got:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("second reader waited for the first")
	}
	if err := a.RUnlock(); err != nil {
		t.Fatal(err)
	}
	if err := b.RUnlock(); err != nil {
		t.Fatal(err)
	}
}

func TestLockWaitsForReader(t *testing.T) {
	a, b := openPair(t, time.Minute)
	if err := a.RLock(); err != nil {
		t.Fatal(err)
	}
	got := make(chan error, 1)
	go func() { got <- b.Lock() }()
	waitBlocked(t, got, a.RUnlock)
	if err := b.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestRLockWaitsForWriter(t *testing.T) {
	a, b := openPair(t, time.Minute)
	if err := a.Lock(); err != nil {
		t.Fatal(err)
	}
	got := make(chan error, 1)
	go func() { got <- b.RLock() }()
	waitBlocked(t, got, a.Unlock)
	if err := b.RUnlock(); err != nil {
		t.Fatal(err)
	}
}

func TestRLockCountsReaders(t *testing.T) {
	a, b := openPair(t, time.Minute)
	for range 2 {
		if err := a.RLock(); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.RUnlock(); err != nil {
		t.Fatal(err)
	}
	if err := fs.TryLock(b.file); !errors.Is(err, fs.ErrBusy) {
		t.Fatalf("one reader left, exclusive got %v, want ErrBusy", err)
	}
	if err := a.RUnlock(); err != nil {
		t.Fatal(err)
	}
	if err := fs.TryLock(b.file); err != nil {
		t.Fatalf("no readers left, exclusive got %v", err)
	}
	if err := fs.Unlock(b.file); err != nil {
		t.Fatal(err)
	}
}

func TestRUnlockWithoutRLock(t *testing.T) {
	a, _ := openPair(t, time.Minute)
	if err := a.RUnlock(); err != nil {
		t.Fatal(err)
	}
	if a.readers != 0 {
		t.Fatalf("readers %d, want 0", a.readers)
	}
}

func TestRLockWritesStamp(t *testing.T) {
	a, _ := openPair(t, time.Minute)
	at := time.Unix(0, 1_700_000_000_000_000_000)
	a.now = func() time.Time { return at }
	if err := a.RLock(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.RUnlock() })
	got, ok, err := readStamp(a.file)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got != at.UnixNano() {
		t.Fatalf("stamp %d ok=%v, want %d", got, ok, at.UnixNano())
	}
}

// The last writer left a stamp older than timeout.
// A live reader refreshes it, so the next writer waits instead of stealing.
func TestLockDoesNotStealFreshReader(t *testing.T) {
	a, b := openPair(t, time.Minute)
	a.now = func() time.Time { return time.Now().Add(-time.Hour) }
	if err := a.Lock(); err != nil {
		t.Fatal(err)
	}
	if err := a.Unlock(); err != nil {
		t.Fatal(err)
	}
	a.now = time.Now
	if err := a.RLock(); err != nil {
		t.Fatal(err)
	}
	got := make(chan error, 1)
	go func() { got <- b.Lock() }()
	waitBlocked(t, got, a.RUnlock)
	if err := b.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestRLockStealsStale(t *testing.T) {
	a, b := openPair(t, time.Minute)
	if err := a.Lock(); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	b.now = func() time.Time { return start.Add(time.Hour) }
	got := make(chan error, 1)
	go func() { got <- b.RLock() }()
	select {
	case err := <-got:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not steal stale lock")
	}
	if err := b.RUnlock(); err != nil {
		t.Fatal(err)
	}
}

// Many goroutines on one File: the OS lock is held while any of them reads.
func TestRLockConcurrentGoroutines(t *testing.T) {
	a, b := openPair(t, time.Minute)
	const n = 32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range n {
		wg.Go(func() {
			<-start
			for range 50 {
				if err := a.RLock(); err != nil {
					t.Error(err)
					return
				}
				if err := fs.TryLock(b.file); !errors.Is(err, fs.ErrBusy) {
					t.Errorf("exclusive during read: %v", err)
				}
				if err := a.RUnlock(); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	close(start)
	wg.Wait()
	if a.readers != 0 {
		t.Fatalf("readers %d, want 0", a.readers)
	}
	if err := b.Lock(); err != nil {
		t.Fatal(err)
	}
	if err := b.Unlock(); err != nil {
		t.Fatal(err)
	}
}
