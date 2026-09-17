# 1. Filesystem as the database; DeckBuilder on-disk compatibility

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

This module replaces [HardDie/fsentry](https://github.com/HardDie/fsentry). That tree is unfinished (duplicate packages, stub `List`, public API in `main.go`). [HardDie/DeckBuilder](https://github.com/HardDie/DeckBuilder) already stores games, collections, and settings as a directory of folders, `.json` entries, and `.bin` files. Callers need a library that reads and writes that tree, not a SQL engine or a POSIX convenience wrapper.

## Considered options

1. **Copy the old fsentry layout and finish it** — keep `pkg/fsentry`, services, and repositories.
2. **New backend (SQLite or similar)** with an import from the folder tree.
3. **Clean rewrite:** same on-disk contract as DeckBuilder; new Go packages; no DeckBuilder domain types in this module.

## Decision

Use option 3.

The filesystem **is** the database. A zip of the root is a backup. Object kinds are folder (directory + `.info.json`), entry (`<id>.json` envelope), and binary (`<id>.bin` raw bytes). IDs come from `NameToID`. Existing DeckBuilder trees must keep reading. Changing that format needs an explicit request and a new ADR.

Out of scope unless asked: HTTP, GUI, games/cards as types, replication, encryption, a daemon.

## Consequences

### Positive

* DeckBuilder stays a compatibility test for bytes on disk; call sites can move to this module without migrating data.
* Agents have a hard rule: do not invent a parallel layout.

### Negative and risks

* POSIX limits (name length, case-insensitive volumes, locking on network FS) are the product’s limits.
* QuotedString and envelope fields are compatibility debt; they stay because old files use them.

### Neutral

* Old fsentry is a behavior reference, not a file to copy wholesale.
* Permissions stay `0755` / `0666` masked by umask unless a Windows test proves we need ACL helpers again.
