package fsentry

import "time"

// FolderInfo is the public view of a folder's `.info.json`.
type FolderInfo[T any] struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      T
}

// Entry is the public view of an `<id>.json` document.
type Entry[T any] struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      T
}

// List is the IDs found in one directory. Names are filesystem IDs, not display names.
type List struct {
	Folders         []string
	Entries         []string
	Binaries        []string
	CorruptedFolder []string
}
