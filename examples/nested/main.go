// Nested folders via the path chain (empty path is the store root).
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

type meta struct {
	Level int `json:"level"`
}

func main() {
	dir, err := os.MkdirTemp("", "fsentry-nested-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fmt.Println("root:", dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	if _, err := db.CreateFolder("Games", meta{Level: 1}); err != nil {
		log.Fatal(err)
	}
	if _, err := db.CreateFolder("My Game", meta{Level: 2}, "Games"); err != nil {
		log.Fatal(err)
	}
	cards, err := db.CreateFolder("Cards", meta{Level: 3}, "games", "my_game")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deep folder", cards.ID)

	_, err = db.CreateEntry("rules", map[string]string{"text": "draw 7"}, "Games", "My Game")
	if err != nil {
		log.Fatal(err)
	}

	list, err := db.List("games", "my_game")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("under my_game folders", list.Folders, "entries", list.Entries)
}
