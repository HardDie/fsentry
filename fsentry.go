package fsentry

import (
	"path/filepath"
	"sync"
	"time"
)

// Logger receives unexpected sync or close failures. Payload bytes are never logged.
// A nil Logger on *DB discards all messages.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Option configures a *DB created by New.
type Option func(*DB)

// WithPretty indents JSON envelopes with a tab. Default is compact JSON.
func WithPretty() Option {
	return func(db *DB) {
		db.pretty = true
	}
}

// WithLogger sets the logger. A nil log is ignored and the discard default stays.
func WithLogger(log Logger) Option {
	return func(db *DB) {
		if log == nil {
			return
		}
		db.log = log
	}
}

// WithNoLockFile skips the inter-process lock file. Production callers omit this.
// Tests and benchmarks pass it so they stay fast and race-detector-friendly.
func WithNoLockFile() Option {
	return func(db *DB) {
		db.noLock = true
	}
}

// WithLockFile turns the inter-process lock file on (the default).
func WithLockFile() Option {
	return func(db *DB) {
		db.noLock = false
	}
}

// WithLockTimeout sets how long a waiter waits before stealing a stale lock.
// d <= 0 keeps the default (10 minutes).
func WithLockTimeout(d time.Duration) Option {
	return func(db *DB) {
		if d <= 0 {
			return
		}
		db.lockTimeout = d
	}
}

// DB is a handle to a store root. New does not create the directory or take
// the lock; call Init before other methods.
type DB struct {
	root        string
	pretty      bool
	noLock      bool
	lockTimeout time.Duration
	log         Logger

	mu sync.RWMutex
}

// New returns a store handle for root. It does not create files, take the
// lock, or validate that root exists. Empty root is stored as empty (Init
// will fail with ErrBadPath). Otherwise root is filepath.Clean'd.
func New(root string, opts ...Option) *DB {
	db := &DB{root: root}
	if root != "" {
		db.root = filepath.Clean(root)
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(db)
	}
	return db
}
