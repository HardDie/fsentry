package fsentry

import "time"

// SetClock replaces the store clock in tests.
func SetClock(db *DB, now func() time.Time) {
	db.now = now
}
