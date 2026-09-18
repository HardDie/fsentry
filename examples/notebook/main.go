// A small nested app: notebooks, notes, cover images, then zip + validate.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

type notebook struct {
	Color string `json:"color"`
}

type page struct {
	Body string `json:"body"`
}

func main() {
	dir, err := os.MkdirTemp("", "fsentry-notebook-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile(), fsentry.WithPretty())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	if _, err := db.CreateEntry("settings", map[string]string{"theme": "dark"}); err != nil {
		log.Fatal(err)
	}

	recipes, err := db.CreateFolder("Recipes", notebook{Color: "green"})
	if err != nil {
		log.Fatal(err)
	}
	work, err := db.CreateFolder("Work", notebook{Color: "blue"})
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.CreateEntry("Soup", page{Body: "simmer"}, recipes.ID)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.CreateBinary("photo", []byte{0x89, 0x50, 0x4e, 0x47}, recipes.ID); err != nil {
		log.Fatal(err)
	}
	_, err = db.CreateEntry("Standup", page{Body: "blockers"}, work.ID)
	if err != nil {
		log.Fatal(err)
	}

	root, err := db.List()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("notebooks", root.Folders, "root entries", root.Entries)

	soup, err := db.GetEntry[page]("Soup", "recipes")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("soup", soup.Data.Body)

	probs, err := db.Validate()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("validate", len(probs))

	var zip bytes.Buffer
	if err := db.Export(&zip, "Recipes"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("recipes zip bytes", zip.Len())
}
