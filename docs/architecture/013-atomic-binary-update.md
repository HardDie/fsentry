# 13. Atomic binary update

* **Status:** Accepted
* **Date:** 2026-10-05
* **Authors:** @oleg

---

## Context

1. `UpdateBinary` truncated `<id>.bin` and wrote the new bytes into it.
2. A crash or power loss mid-write left a partial file.
3. The old bytes were already gone.
4. JSON writes already went through a temporary file and a rename.
5. DeckBuilder is about to rewrite every image of a game in one run (its ADR 027).

## Considered options

1. **Keep the in-place write**
   1. Lost. One crash loses an image.
2. **A new `ReplaceBinary` next to `UpdateBinary`**
   1. Lost. Two update methods; callers would pick the unsafe one.
3. **Make `UpdateBinary` atomic**
   1. Won. Same API; every caller gets the safe write.

## Decision

1. `UpdateBinary` writes `<id>.bin.tmp` in the same folder.
2. It fsyncs the temporary file, closes it, and renames it over `<id>.bin`.
3. JSON and binary replaces share one helper (`writeFileReplace`).
4. A stale `.tmp` from a crash is removed by the next replace.
5. `Validate` already reports a leftover `.tmp` (`ProblemTemp`).
6. `CreateBinary` is unchanged: there is no old content to lose.
7. A temp file is never visible as an object.
   1. IDs never hold a dot (`NameToID`), so no `Create*` or `Get*` name reaches `*.tmp`.
   2. `List` counts only `.json` and `.bin` files.
   3. `Export`, `ExportFolder`, and `Import` skip `*.tmp`, like the lock file.

## Consequences

### Positive

1. After a crash a binary holds its old bytes or its new bytes, never part of them.
2. JSON and binaries follow one write rule.

### Negative and risks

1. An update is about 10% slower (4.1 → 4.6 ms; fsync dominates).
2. Six more allocations per update: five in OS calls, one for the `.tmp` path.
3. On Windows, the rename fails if another process holds the file open without delete sharing.
   1. The update then returns an error and the old file stays.
4. The folder itself is not fsynced after the rename, the same as for JSON.

### Neutral

1. The file gets a new inode on every update.
