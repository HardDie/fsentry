// Validate reports on-disk problems; it does not repair.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/HardDie/fsentry"
)

func main() {
	dir, err := os.MkdirTemp("", "fsentry-val-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("Notes", nil); err != nil {
		log.Fatal(err)
	}

	ok, err := db.Validate()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("healthy problems", len(ok))

	if err := os.Mkdir(filepath.Join(dir, "broken"), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "stray.txt"), []byte("x"), 0o666); err != nil {
		log.Fatal(err)
	}

	bad, err := db.Validate()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range bad {
		fmt.Println(p)
	}
}
