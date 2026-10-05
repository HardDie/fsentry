# 14. Optional progress callback for zip import

* **Status:** Accepted
* **Date:** 2026-10-05
* **Authors:** @HardDie

---

## Context

1. A DeckBuilder game archive can hold hundreds of MB of images.
2. Import writes and syncs every file under the write lock.
3. That takes seconds; the app wants to show a progress bar.
4. `Import` and `ImportFolder` end in a variadic `path ...string`.
   1. An options argument cannot be appended without breaking callers.

## Considered options

1. **Sibling methods with a callback**
   1. Won.
   2. `ImportWithProgress` and `ImportFolderWithProgress`.
   3. Old methods call them with nil; no caller changes.
2. **A `DB` option set at `New`**
   1. Lost. One callback for every import on the handle.
   2. A caller with two imports cannot tell them apart.
3. **A counting `io.ReaderAt` in the caller**
   1. Lost. It relies on how the zip is read, not on files written.
4. **A channel of progress values**
   1. Lost. The caller must drain it, or extraction blocks.

## Decision

Use option 1.

1. `ImportProgress` has `Files`, `FilesTotal`, `Bytes`, `BytesTotal`.
2. Files only; directories are not counted.
3. Bytes are uncompressed sizes from the zip headers.
   1. A rewritten `.info.json` counts its new body.
4. Called once before the first file, then after each file.
5. Not called when the archive is refused before extraction.
6. It runs under the write lock.
   1. It must not call the same `*DB`.
7. A nil callback is the plain call.

## Consequences

### Positive

1. Callers can show files or bytes done, with exact totals up front.
2. `Import` and `ImportFolder` keep their signatures.
3. No cost without a callback: totals are not even summed.

### Negative and risks

1. Two more public methods per import kind.
2. A slow callback slows the import, since it runs inline.
3. A huge single file reports only when it is done.

### Neutral

1. Header sizes are trusted; a lying zip only skews the bar.
