# Binaries

A binary is raw bytes in `<id>.bin`. There is **no** metadata file and no JSON envelope. Names still go through [NameToID](Paths-and-IDs). Related: [folders](Folders), [entries](Entries).

Prerequisites: [Getting started](Getting-Started).

## Create, update, move, remove

```go
png := []byte{0x89, 0x50, 0x4e, 0x47}

if err := db.CreateBinary("cover", png, "My Notes"); err != nil {
	return err
}

if err := db.UpdateBinary("cover", append(png, 0x00), "my_notes"); err != nil {
	return err
}

if err := db.MoveBinary("cover", "art", "my_notes"); err != nil {
	return err
}

if err := db.RemoveBinary("art", "my_notes"); err != nil {
	return err
}
```

Create of an existing ID is [`ErrExist`](Errors). Update/remove of a missing file is [`ErrNotExist`](Errors).

`List` includes binary IDs (without `.bin`):

```go
list, err := db.List("my_notes")
fmt.Println(list.Binaries)
```

## Get into a buffer

`GetBinary` reads into `buf` and returns that slice, or a grown one if `buf` is too small. `buf == nil` is allowed and allocates.

```go
data, err := db.GetBinary("cover", nil, "my_notes")
```

Reuse a sized buffer when you care about allocations (hot path: `cap(buf)` ≥ file size):

```go
buf := make([]byte, 0, 64*1024)
data, err := db.GetBinary("cover", buf, "my_notes")
// next call can reuse data[:0] if cap is still enough
data, err = db.GetBinary("cover", data[:0], "my_notes")
```

There is no `DuplicateBinary`; copy with `GetBinary` then `CreateBinary` under a new name.
