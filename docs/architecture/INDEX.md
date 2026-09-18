# Architecture decision records

ADRs live in this folder. Number them in order (`001`, `002`, …). Status is one of: Proposed, Accepted, Deprecated, Superseded.

| ID | Title | Status |
|---|---|---|
| [001](001-filesystem-as-database.md) | Filesystem as the database; DeckBuilder on-disk compatibility | Accepted |
| [002](002-one-public-package.md) | One public package at module root; internals by object | Accepted |
| [003](003-generic-methods-no-interface.md) | Generic methods on `*DB`; no `IFSEntry` | Accepted |
| [004](004-docs-layout.md) | README, CURSOR.md, ADRs, use cases after code exists | Accepted |
| [005](005-os-adapter-sentinels.md) | `internal/fs` maps OS errors to sentinels; `wrap` is zero-alloc | Accepted |
| [006](006-internal-fs-file-layout.md) | `internal/fs` files follow package `os` | Accepted |
| [007](007-advisory-lock-and-steal.md) | Advisory lock file, stamp, steal after timeout | Accepted |
| [008](008-exported-benchmarks-zero-wrap.md) | Benches on exported helpers; `wrap` returns the sentinel only | Accepted |
| [009](009-tests-tempdir-no-lock.md) | Race tests, `t.TempDir()`, lock off unless testing the lock | Accepted |
| [010](010-zip-export-import.md) | Zip export and import of a store tree | Accepted |
| [011](011-validate-problems.md) | Validate walks the tree and reports problems; it does not repair | Accepted |
