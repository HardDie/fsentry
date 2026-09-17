# 4. Documentation layout: README, CURSOR.md, ADRs, use cases

* **Status:** Accepted
* **Date:** 2026-09-17
* **Authors:** @oleg

---

## Context

The rewrite needs a short user entry, agent/implementation rules, recorded architecture decisions, and scenarios for code that actually exists. Putting all of that in `CURSOR.md` or a README makes both unusable. Use-case files written before the code exist become fiction.

## Considered options

1. **Only README.**
2. **CURSOR.md only** (agent spec doubles as user and architecture docs).
3. **README (users) + CURSOR.md (agents) + `docs/architecture` (ADRs) + `docs/use-cases` (only after the code exists).** No wiki until there is operator/product how-to beyond a library README.

## Decision

Use option 3.

Write `docs/use-cases/uc-NN-….md` from `_TEMPLATE.md` when behavior lands; set the row to Done in `docs/use-cases/INDEX.md`. Planned rows may exist so numbering stays stable. New core decisions get the next ADR number and an INDEX row. Do not leave accepted decisions only in chat.

## Consequences

### Positive

* Users start at README; agents start at CURSOR.md; “why” lives in numbered ADRs.
* Use cases stay honest.

### Negative and risks

* Several places to update when the on-disk contract or exported API changes (CURSOR.md + README if user-visible + ADR if the decision changed + UC if behavior landed).

### Neutral

* This library has no `docs/wiki/` until someone needs operator guides that do not belong in README.
