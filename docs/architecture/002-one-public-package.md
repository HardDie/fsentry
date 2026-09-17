# 2. One public package at module root; internals by object

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

Old fsentry split `package main`, `pkg/fsentry`, `pkg/fsentry_error`, and `internal/service` + `internal/folder/service` + `internal/repository/folder` (+ entity types that duplicated the public structs). That is hard to test and easy to fork accidentally. This repo is a **library**, not an application.

## Considered options

1. **Keep the old layer sandwich** and only fix bugs.
2. **`pkg/fsentry` + `internal/…`** as in some community layouts.
3. **One exported package `fsentry` at the module root.** Internals grouped by object (`fs`, `lock`, `name`, `jsonutil`), not by layer.

## Decision

Use option 3. Module path is `github.com/HardDie/fsentry`. Callers import only `fsentry`. No `pkg/`. No `cmd/` until someone asks for a CLI. No `internal/entity`. No `package main`.

Import direction: public `fsentry` → `internal/*`. Internals do not import the public API.

## Consequences

### Positive

* `go get` / `go doc` match a normal library.
* One implementation per object kind (`folder.go`, `entry.go`, `binary.go` at the root when those land).

### Negative and risks

* The root package will grow several files; that is preferable to a fake hexagonal layout.

### Neutral

* [ADR 006](006-internal-fs-file-layout.md) covers how `internal/fs` is split on disk.
