# Use cases

Scenarios for **code that exists**. Do not add `uc-*.md` until that package lives in this repository. This index may list planned IDs so numbering stays stable.

**Status:** Planned (not in this repo yet) · In progress · Done

Copy [\_TEMPLATE.md](_TEMPLATE.md) when behavior lands.

| ID | Name | Module | Status | File |
|---|---|---|---|---|
| UC-01 | Create a file | `internal/fs` | Done | [fs/uc-01-create-file.md](fs/uc-01-create-file.md) |
| UC-02 | Create a folder | `internal/fs` | Done | [fs/uc-02-create-folder.md](fs/uc-02-create-folder.md) |
| UC-03 | Map an OS error to a sentinel | `internal/fs` | Done | [fs/uc-03-map-os-error.md](fs/uc-03-map-os-error.md) |
| UC-04 | Write data to a file | `internal/fs` | Done | [fs/uc-04-write-file.md](fs/uc-04-write-file.md) |
| UC-05 | Rename a file | `internal/fs` | Done | [fs/uc-05-rename-file.md](fs/uc-05-rename-file.md) |
| UC-06 | Rename a folder | `internal/fs` | Done | [fs/uc-06-rename-folder.md](fs/uc-06-rename-folder.md) |
| UC-07 | Open a file for reading | `internal/fs` | Done | [fs/uc-07-open-read.md](fs/uc-07-open-read.md) |
| UC-08 | Open a file for writing (truncate) | `internal/fs` | Done | [fs/uc-08-open-write.md](fs/uc-08-open-write.md) |
| UC-09 | Read from a file | `internal/fs` | Done | [fs/uc-09-read-file.md](fs/uc-09-read-file.md) |
| UC-10 | Close a file | `internal/fs` | Done | [fs/uc-10-close.md](fs/uc-10-close.md) |
| UC-11 | Sync a file | `internal/fs` | Done | [fs/uc-11-sync.md](fs/uc-11-sync.md) |
| UC-12 | Remove a file | `internal/fs` | Done | [fs/uc-12-remove-file.md](fs/uc-12-remove-file.md) |
| UC-13 | Remove a folder | `internal/fs` | Done | [fs/uc-13-remove-folder.md](fs/uc-13-remove-folder.md) |
| UC-14 | Read a directory listing | `internal/fs` | Done | [fs/uc-14-readdir.md](fs/uc-14-readdir.md) |
| UC-15 | Stat a path | `internal/fs` | Done | [fs/uc-15-stat.md](fs/uc-15-stat.md) |
| UC-16 | Open a lock file | `internal/fs` | Done | [fs/uc-16-open-lock.md](fs/uc-16-open-lock.md) |
| UC-17 | Exclusive lock and unlock | `internal/fs` | Done | [fs/uc-17-lock-unlock.md](fs/uc-17-lock-unlock.md) |
| UC-18 | Non-blocking try-lock | `internal/fs` | Done | [fs/uc-18-try-lock.md](fs/uc-18-try-lock.md) |
| UC-19 | Steal a stale lock | `internal/lock` | Done | [lock/uc-19-steal-stale-lock.md](lock/uc-19-steal-stale-lock.md) |
| UC-20 | Construct a store handle | `fsentry` | Done | [fsentry/uc-20-new.md](fsentry/uc-20-new.md) |
| UC-21 | Init the store root | `fsentry` | Done | [fsentry/uc-21-init.md](fsentry/uc-21-init.md) |
| UC-22 | Create a folder | `fsentry` | Done | [fsentry/uc-22-create-folder.md](fsentry/uc-22-create-folder.md) |
| UC-23 | Get a folder | `fsentry` | Done | [fsentry/uc-23-get-folder.md](fsentry/uc-23-get-folder.md) |
| UC-24 | List a directory | `fsentry` | Done | [fsentry/uc-24-list.md](fsentry/uc-24-list.md) |
| UC-25 | Move (rename) a folder | `fsentry` | Done | [fsentry/uc-25-move-folder.md](fsentry/uc-25-move-folder.md) |
| UC-26 | Update folder payload | `fsentry` | Done | [fsentry/uc-26-update-folder.md](fsentry/uc-26-update-folder.md) |
| UC-27 | Remove a folder | `fsentry` | Done | [fsentry/uc-27-remove-folder.md](fsentry/uc-27-remove-folder.md) |
