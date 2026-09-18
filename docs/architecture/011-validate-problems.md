# 11. Validate walks the tree and reports problems; it does not repair

* **Status:** Accepted
* **Date:** 2026-09-18
* **Authors:** @oleg

---

## Context

Import, crashes, and hand-edited files can leave a store that `List` only partially describes (`CorruptedFolder` is one directory, one level). Callers need a full pass: envelopes, ID mismatches, leftover `*.tmp`, unexpected names.

## Considered options

1. **Only `List.CorruptedFolder`** — too local; ignores bad entry JSON and ID drift.
2. **`Validate() error`** — first failure only; bad after a large import.
3. **`Validate(path…) ([]Problem, error)`** — collect every defect; `err` is only walk failure. Read-only.

## Decision

Use option 3.

`Problem.Path` is relative with `/`. Codes: `info`, `json`, `id`, `name`, `bad_name`, `temp`, `unexpected`. Nil slice means OK. Do not auto-delete or rewrite files.

## Consequences

### Positive

* After `Import`, callers can `Validate` and show a report.
* Repair stays a separate, explicit feature.

### Negative and risks

* Binary payloads are not inspected (only the `.bin` ID).
* A permission error on `ReadDir` aborts the walk (`err` non-nil, list maybe partial).

### Neutral

* Root is not a folder object; `.info.json` at the store root is `unexpected`.
