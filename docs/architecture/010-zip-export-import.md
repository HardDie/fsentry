# 10. Zip export and import of a store tree

* **Status:** Accepted
* **Date:** 2026-09-18
* **Authors:** @oleg

---

## Context

The on-disk tree is the database; a zip of the root is a backup. [DeckBuilder](https://github.com/HardDie/DeckBuilder) already exports a game folder as a zip with the folder ID as the top-level name and imports that layout. Callers need the same operations on `*DB` without HTTP or game types.

## Considered options

1. **Document “zip the directory yourself”** and keep zip logic in DeckBuilder.
2. **`Export() ([]byte, error)` / `Import([]byte)`** like DeckBuilder’s repository.
3. **`Export(io.Writer)` / `Import(io.Reader)`** with optional `path` so a subtree can be archived; zip names always `/`; skip the lock file; refuse zip-slip.

## Decision

Use option 3.

`Export(w, path…)` walks the directory at `path` (empty = root). Non-root exports prefix zip entries with the folder ID. `Import(r, path…)` extracts into that directory. Existing files are `ErrExist`. Invalid zip or escape from dest is `ErrBadArchive`. A failed extract deletes files and directories this call created. `io.Reader` that also has `ReaderAt` and `Size()` is opened without buffering the whole archive.

## Consequences

### Positive

* DeckBuilder can call `Export` / `Import` on a game folder path.
* Tests write to `bytes.Buffer` and read with `bytes.NewReader`.

### Negative and risks

* `Import` of a plain `io.Reader` still buffers the zip in memory.
* Zip bombs (huge uncompressed size) are not limited in this slice.

### Neutral

* Lock file is process state, not data; it is never archived or restored.
