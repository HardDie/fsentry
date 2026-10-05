package fsentry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDoesNotCreateRoot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	db := New(path)
	if db == nil {
		t.Fatal("New returned nil")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("New must not create root: %v", err)
	}
}

func TestNewCleansRoot(t *testing.T) {
	db := New("a/./b")
	want := filepath.Clean("a/./b")
	if db.root != want {
		t.Fatalf("root %q, want %q", db.root, want)
	}
}

func TestNewEmptyRoot(t *testing.T) {
	db := New("")
	if db.root != "" {
		t.Fatalf("empty root became %q", db.root)
	}
}

func TestNewDefaults(t *testing.T) {
	db := New("root")
	if db.pretty {
		t.Fatal("pretty default")
	}
	if db.noLock {
		t.Fatal("lock should be on by default")
	}
	if db.lockTimeout != 0 {
		t.Fatalf("timeout %v, want 0 (Init uses DefaultTimeout)", db.lockTimeout)
	}
	if db.log != (discardLogger{}) {
		t.Fatal("logger default is discard")
	}
}

func TestNewOptions(t *testing.T) {
	log := namedLogger{name: "test"}
	timeout := 5 * time.Minute
	db := New("root", WithPretty(), WithLogger(log), WithNoLockFile(), WithLockTimeout(timeout))
	if !db.pretty {
		t.Fatal("WithPretty")
	}
	if db.log != log {
		t.Fatal("WithLogger")
	}
	if !db.noLock {
		t.Fatal("WithNoLockFile")
	}
	if db.lockTimeout != timeout {
		t.Fatalf("timeout %v, want %v", db.lockTimeout, timeout)
	}
}

func TestWithLockFileOverridesNoLock(t *testing.T) {
	db := New("root", WithNoLockFile(), WithLockFile())
	if db.noLock {
		t.Fatal("WithLockFile should enable the lock")
	}
}

func TestWithLoggerNilIgnored(t *testing.T) {
	log := namedLogger{name: "test"}
	db := New("root", WithLogger(log), WithLogger(nil))
	if db.log != log {
		t.Fatal("nil logger must not clear a previous logger")
	}
}

func TestWithLockTimeoutNonPositive(t *testing.T) {
	db := New("root", WithLockTimeout(0), WithLockTimeout(-time.Second))
	if db.lockTimeout != 0 {
		t.Fatalf("non-positive timeout set %v", db.lockTimeout)
	}
}

func TestNewNilOption(t *testing.T) {
	db := New("root", nil, WithPretty())
	if !db.pretty {
		t.Fatal("nil option should be skipped")
	}
}

// namedLogger is a Logger that compares unequal to the default.
type namedLogger struct{ name string }

func (namedLogger) Debug(string, ...any) {}
func (namedLogger) Info(string, ...any)  {}
func (namedLogger) Warn(string, ...any)  {}
func (namedLogger) Error(string, ...any) {}
