# Folders

A folder is a directory named by [ID](Paths-and-IDs) plus `.info.json` inside it. It may contain more folders, [entries](Entries), and [binaries](Binaries). Payload JSON lives in `.info.json` `data`. See [on-disk format](On-Disk-Format).

Prerequisites: [Getting started](Getting-Started) (`New` + `Init`).

## Create and get

```go
type Game struct {
	Description string `json:"description"`
}

created, err := db.CreateFolder("My Game", Game{Description: "coop"})
if err != nil {
	return err
}
// created.ID == "my_game"

got, err := db.GetFolder[Game]("My Game")
// same as GetFolder[Game]("my_game")
```

Create fails with [`ErrExist`](Errors) if the ID is already a directory. Missing parent path is [`ErrBadPath`](Errors). A directory without readable `.info.json` (or whose envelope id/name does not match the disk ID) is [`ErrFolderCorrupted`](Errors) on get. The same `ErrFolderCorrupted` is returned if you pass that directory as a parent `path` for a nested create, get, list, export, or validate.

## Nested folders

`path` is the parent chain from the store root. Use names or IDs; both are sanitized the same way.

```go
_, err := db.CreateFolder("Cards", map[string]int{"count": 0}, "My Game", "My Collection")
```

On disk that is `<root>/my_game/my_collection/cards/` plus `.info.json`. `my_game` and `my_collection` must both be valid folders first.

## List children

```go
list, err := db.List("my_game")
if err != nil {
	return err
}
for _, id := range list.Folders {
	fmt.Println(id)
}
fmt.Println(list.Entries, list.Binaries, list.CorruptedFolder)
```

Empty `List()` lists the root. `List` skips `.info.json` and `.fsentry.lock`.

## Update payload

`UpdateFolder` replaces `data` and sets `updatedAt` to now. It does not rename.

```go
updated, err := db.UpdateFolder("My Game", Game{Description: "solo"})
```

Missing folder: [`ErrNotExist`](Errors).

## Move (rename)

`MoveFolder` renames the directory, rewrites `id` / `name` in `.info.json`, and bumps `updatedAt`. Children stay inside the moved tree.

```go
moved, err := db.MoveFolder[Game]("My Game", "Other Game")
// directory my_game → other_game
```

Target ID already present: [`ErrExist`](Errors).

## Remove

`RemoveFolder` deletes the directory recursively. It refuses a folder whose `.info.json` cannot be read (`ErrFolderCorrupted`) so a half-written tree is not silently wiped.

```go
err := db.RemoveFolder("Other Game")
```

## Payload types

Use a struct (or `map`, `json.RawMessage`, …). Inference works on create/update from the value. Get and move take `[T]` so JSON unmarshals into the right type.

```go
raw, err := db.GetFolder[json.RawMessage]("my_game")
```
