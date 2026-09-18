# UC-42: Full public-API flow (export identity, isolated delete)

**Module:** `integration` (`package integration_test`)  
**Status:** Done  
**Actors:** CI / `make test-integration`  
**Goal:** One external test calls every exported store method; writes match reads; deletes do not remove siblings; zip export/import is byte-identical  
**Preconditions:** Go 1.27+; temp directory  

## Main scenario (happy path)

1. Construct `*DB` with options (`WithPretty`, `WithLogger`, `WithLockTimeout`, `WithNoLockFile`). `Init` twice.
2. Create nested folders, entries, binaries. Create return values equal `Get*`. `Validate` is empty.
3. Update / move / duplicate. Source of duplicate unchanged. Unrelated objects unchanged.
4. `Export` the store; `Import` into a new root. File trees (except `.fsentry.lock`) are byte-identical. API snapshot (IDs, names, times, payloads) matches.
5. Folder `Export`/`Import` matches the `games/` subtree files.
6. Remove one entry, one binary, one folder. Siblings still get and `Validate` is empty.
7. `Drop` removes the root. A second store with `WithLockFile` creates `.fsentry.lock` then `Drop`.

## Alternative scenarios and errors

* **6a. Get under a removed folder:** `ErrBadPath` or `ErrNotExist`.

## Postconditions

* Default `go test ./...` does not run this file (`//go:build integration`).
