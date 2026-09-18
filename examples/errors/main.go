// Sentinel errors: match with errors.Is, do not inspect OS errnos.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

func main() {
	dir, err := os.MkdirTemp("", "fsentry-errors-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	_, err = db.CreateFolder[any]("!!!", nil)
	fmt.Println("bad name", errors.Is(err, fsentry.ErrBadName))

	if _, err := db.CreateFolder[any]("Notes", nil); err != nil {
		log.Fatal(err)
	}
	_, err = db.CreateFolder[any]("Notes", nil)
	fmt.Println("exist", errors.Is(err, fsentry.ErrExist))

	_, err = db.GetFolder[any]("Missing")
	fmt.Println("not exist", errors.Is(err, fsentry.ErrNotExist))

	_, err = db.CreateEntry("x", struct{}{}, "no_parent")
	fmt.Println("bad path", errors.Is(err, fsentry.ErrBadPath))
}
