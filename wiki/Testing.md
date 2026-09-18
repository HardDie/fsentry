# Testing

Use a real temp-dir store. There is no mock interface for generic methods on `*DB` (Go still forbids type parameters on interface methods). Related: [getting started](Getting-Started), [locking](Locking).

## Unit tests

Pass `WithNoLockFile()` so tests stay fast, isolated, and race-detector-friendly. Use `t.TempDir()`.

```go
package app_test

import (
	"testing"

	"github.com/HardDie/fsentry"
)

func TestNotes(t *testing.T) {
	db := fsentry.New(t.TempDir(), fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}

	_, err := db.CreateFolder("My Notes", map[string]string{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := db.GetFolder[map[string]string]("My Notes")
	if err != nil {
		t.Fatal(err)
	}
	if got.Data["k"] != "v" {
		t.Fatalf("%+v", got)
	}
}
```

Lock tests are the exception: omit `WithNoLockFile()` (or pass `WithLockFile()`) and use two `*DB` handles on the same root. Do not spawn extra processes unless that is the only way to prove the lock.

## Commands

```bash
make test                 # go test -race -count=1 ./...
make test-integration     # go test -tags=integration
make bench                # benchmarks only; Makefile keeps -run '^$'
make ci
```

Integration files use `//go:build integration`. GitHub Actions (`.github/workflows/test.yml`) runs race tests on Linux and macOS, tests without race on Windows, then the integration tag, plus fmt/vet/tidy/examples/lint.

## Benchmarks

Library benches use `WithNoLockFile()` so they measure the store, not `flock`. Keep the mutex on. `b.ReportAllocs()` is the norm for hot paths (`GetBinary` with a sized buffer, `NameToID` into a reused buffer).

```go
func BenchmarkGetBinary(b *testing.B) {
	db := fsentry.New(b.TempDir(), fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		b.Fatal(err)
	}
	payload := []byte("hello")
	if err := db.CreateBinary("x", payload); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 0, len(payload))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var err error
		buf, err = db.GetBinary("x", buf[:0])
		if err != nil {
			b.Fatal(err)
		}
	}
}
```
