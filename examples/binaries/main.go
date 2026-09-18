// Opaque binaries: create, GetBinary into a buffer, update, move, remove.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

func main() {
	dir, err := os.MkdirTemp("", "fsentry-bin-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	if _, err := db.CreateFolder[any]("Art", nil); err != nil {
		log.Fatal(err)
	}

	payload := []byte("PNG-BYTES")
	if err := db.CreateBinary("cover", payload, "Art"); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 0, 64)
	got, err := db.GetBinary("cover", buf, "art")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("get", string(got))

	if err := db.UpdateBinary("cover", []byte("JPEG"), "art"); err != nil {
		log.Fatal(err)
	}
	if err := db.MoveBinary("cover", "hero", "art"); err != nil {
		log.Fatal(err)
	}
	got, err = db.GetBinary("hero", got[:0], "art")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("moved", string(got))

	if err := db.RemoveBinary("hero", "art"); err != nil {
		log.Fatal(err)
	}
	_, err = db.GetBinary("hero", nil, "art")
	fmt.Println("removed", errors.Is(err, fsentry.ErrNotExist))
}
