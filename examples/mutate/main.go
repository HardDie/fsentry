// Update, move, duplicate, and remove folders and entries.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

type meta struct {
	N int `json:"n"`
}

type body struct {
	Text string `json:"text"`
}

func main() {
	dir, err := os.MkdirTemp("", "fsentry-mutate-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	if _, err := db.CreateFolder("Drafts", meta{N: 1}); err != nil {
		log.Fatal(err)
	}
	updated, err := db.UpdateFolder("Drafts", meta{N: 2})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("updated folder", updated.Data.N)

	moved, err := db.MoveFolder[meta]("Drafts", "Archive")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("moved folder", moved.ID)

	copy, err := db.DuplicateFolder[meta]("Archive", "Archive Copy")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("duplicate folder", copy.ID)

	if _, err := db.CreateEntry("note", body{Text: "v1"}, "archive"); err != nil {
		log.Fatal(err)
	}
	ent, err := db.UpdateEntry("note", body{Text: "v2"}, "archive")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("updated entry", ent.Data.Text)

	dup, err := db.DuplicateEntry[body]("note", "note copy", "archive")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("duplicate", dup.ID)

	renamed, err := db.MoveEntry[body]("note copy", "snapshot", "archive")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("moved entry", renamed.ID)

	if err := db.RemoveEntry("snapshot", "archive"); err != nil {
		log.Fatal(err)
	}
	_, err = db.GetEntry[body]("snapshot", "archive")
	fmt.Println("removed", errors.Is(err, fsentry.ErrNotExist))
}
