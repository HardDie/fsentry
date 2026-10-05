package fsentry

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/HardDie/fsentry/internal/lock"
)

// Logger receives the store's log lines, shaped like log/slog: a short
// lowercase message, then key-value pairs ("err", err; "path", p).
// A *slog.Logger satisfies it. Error and Warn are unexpected failures
// (sync, unlock, rollback); Debug follows long operations step by step (import).
// Payload bytes are never logged. Without WithLogger, lines are discarded.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// discardLogger is the default Logger: it drops every line.
type discardLogger struct{}

func (discardLogger) Debug(string, ...any) {}
func (discardLogger) Info(string, ...any)  {}
func (discardLogger) Warn(string, ...any)  {}
func (discardLogger) Error(string, ...any) {}

// Option configures a *DB created by New.
type Option func(*DB)

// WithPretty indents JSON envelopes with a tab. Default is compact JSON.
func WithPretty() Option {
	return func(db *DB) {
		db.pretty = true
	}
}

// WithLogger sets the logger. A nil log is ignored and the discard default stays.
// Debug lines are many on an import: one per file. Filter them in the logger.
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
	now         func() time.Time

	lk *lock.File
	mu sync.RWMutex
}

// New returns a store handle for root. It does not create files, take the
// lock, or validate that root exists. Empty root is stored as empty (Init
// will fail with ErrBadPath). Otherwise root is filepath.Clean'd.
func New(root string, opts ...Option) *DB {
	db := &DB{root: root, log: discardLogger{}}
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
