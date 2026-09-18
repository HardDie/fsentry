// Basic store: Init, one folder, one entry, one binary, List.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

type meta struct {
	Kind string `json:"kind"`
}

func main() {
	dir, err := os.MkdirTemp("", "fsentry-basic-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	folder, err := db.CreateFolder("My Notes", meta{Kind: "notebook"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("folder", folder.ID, folder.Name)

	ent, err := db.CreateEntry("Welcome", map[string]string{"body": "hello"}, folder.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("entry", ent.ID)

	if err := db.CreateBinary("cover", []byte("PNG"), folder.ID); err != nil {
		log.Fatal(err)
	}

	list, err := db.List()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("root folders", list.Folders)

	inner, err := db.List("My Notes")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("notes entries", inner.Entries, "binaries", inner.Binaries)
}
