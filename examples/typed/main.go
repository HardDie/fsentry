// Typed JSON payloads: inference on create, explicit [T] on get, RawMessage, nil.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

type settings struct {
	Theme string `json:"theme"`
	N     int    `json:"n"`
}

func main() {
	dir, err := os.MkdirTemp("", "fsentry-typed-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	created, err := db.CreateEntry("settings", settings{Theme: "dark", N: 2})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("inferred", created.Data.Theme)

	got, err := db.GetEntry[settings]("settings")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("get", got.Data.Theme, got.Data.N)

	raw, err := db.GetEntry[json.RawMessage]("settings")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("raw", string(raw.Data))

	empty, err := db.CreateFolder[any]("Empty", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nil payload folder", empty.ID)
}
