# 15. slog-shaped log lines; Debug steps for import

* **Status:** Accepted
* **Date:** 2026-10-06
* **Authors:** @HardDie

---

## Context

1. The store logged only "sync failed" and "unlock failed".
   1. No path, no error: a line could not be traced.
2. Every call site checked `db.log != nil`.
3. DeckBuilder logs with `log/slog`.
   1. Short lowercase message plus fields; errors as `"err", err`.
   2. It never passed a logger, so store lines were lost.
4. A large import is slow, and its failures hide behind one sentinel.
   1. `ErrBadArchive` has several causes.

## Considered options

1. **slog-shaped lines, Debug steps for import**
   1. Won. A `*slog.Logger` plugs in; the level filters Debug.
2. **Import `log/slog` directly**
   1. Lost. Callers with another logger would need an adapter.
   2. The `Logger` interface already matches `*slog.Logger`.
3. **Return wrapped errors with the cause**
   1. Lost. `wrap` stays zero-alloc and returns the sentinel only (ADR 008).

## Decision

Use option 1.

1. Message: short, lowercase; fields after it.
2. Errors go as `"err", err`; paths as `"path"` or `"entry"`.
3. Error: unexpected sync, unlock, rollback-remove failures.
4. Warn: an import rolled back.
5. Debug: each import step, one line per file.
   1. Prefix `import folder:` or `import:`.
   2. `failed` names the `step` where it stopped.
   3. A refusal also logs the cause behind the sentinel.
6. The default logger discards; call sites do not check for nil.
7. Payload bytes are never logged.

## Consequences

### Positive

1. A slow import shows lock wait, zip read time, and time per file.
2. A refused archive says why.
3. DeckBuilder passes `slog.Default()`; lines join its log file.

### Negative and risks

1. A 1000-file import writes 1000 Debug lines when Debug is on.
2. Debug arguments are built even when discarded: a few allocs per file.
   1. Next to an fsync per file, it does not show.

### Neutral

1. Hot paths (`GetBinary`, `NameToID`) log nothing; their alloc rules stand.
