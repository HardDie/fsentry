# Validate

Walk a store (or one [folder](Folders)) and list on-disk defects. This does **not** repair files. Use it after [import](Export-and-Import), a crash, or hand-editing. Related: [on-disk format](On-Disk-Format), [errors](Errors).

Prerequisites: [Getting started](Getting-Started).

## Call

```go
probs, err := db.Validate()
if err != nil {
	return err // walk failed (missing path, permission, …)
}
if len(probs) == 0 {
	return nil // tree matches the contract
}
for _, p := range probs {
	fmt.Println(p) // e.g. "info: broken"
}
```

Validate a subtree:

```go
probs, err := db.Validate("My Notes")
```

## What is checked

| Code | Meaning |
|---|---|
| `info` | Folder `.info.json` missing or not a valid envelope |
| `json` | Entry `*.json` is not a valid envelope |
| `id` | Envelope `id` does not match the directory or file ID on disk |
| `name` | Envelope `name` does not sanitize (`NameToID`) to that ID |
| `bad_name` | The filesystem name itself is not a valid ID |
| `temp` | Leftover write file (`*.tmp`) |
| `unexpected` | Not a folder, entry, binary, folder `.info.json`, or lock file |

`.fsentry.lock` is skipped. The store **root** is not a folder: it has no `.info.json`. Nested directories must have one.

`Problem.Path` is relative to the directory you validated, with `/` (example: `my_notes/welcome.json`).

Binaries: only the `.bin` file name is checked, not the bytes.

## After import

```go
if err := db.Import(r); err != nil {
	return err
}
probs, err := db.Validate()
if err != nil {
	return err
}
if len(probs) > 0 {
	return fmt.Errorf("imported store is invalid: %v", probs)
}
```

## What it does not do

- Delete corrupted folders or stray files
- Rewrite envelope `id` / `name` to match the path
- Replace `List.CorruptedFolder` — `List` still reports unreadable folders at **one** level; `Validate` walks the whole subtree
