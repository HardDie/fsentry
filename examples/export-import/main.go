// Export a store (or one folder) as zip and Import into another root.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/HardDie/fsentry"
)

type meta struct {
	Kind string `json:"kind"`
}

func main() {
	srcDir, err := os.MkdirTemp("", "fsentry-src-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(srcDir) }()
	dstDir, err := os.MkdirTemp("", "fsentry-dst-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dstDir) }()

	src := fsentry.New(srcDir, fsentry.WithNoLockFile())
	if err := src.Init(); err != nil {
		log.Fatal(err)
	}
	if _, err := src.CreateFolder("My Notes", meta{Kind: "notebook"}); err != nil {
		log.Fatal(err)
	}
	if _, err := src.CreateEntry("Welcome", map[string]string{"body": "hi"}, "My Notes"); err != nil {
		log.Fatal(err)
	}

	var whole bytes.Buffer
	if err := src.Export(&whole); err != nil {
		log.Fatal(err)
	}
	fmt.Println("store zip bytes", whole.Len())

	dst := fsentry.New(dstDir, fsentry.WithNoLockFile())
	if err := dst.Init(); err != nil {
		log.Fatal(err)
	}
	if err := dst.Import(bytes.NewReader(whole.Bytes())); err != nil {
		log.Fatal(err)
	}
	got, err := dst.GetEntry[map[string]string]("Welcome", "my_notes")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("imported", got.Data["body"])

	var folderZip bytes.Buffer
	if err := src.Export(&folderZip, "My Notes"); err != nil {
		log.Fatal(err)
	}
	inbox, err := dst.CreateFolder[any]("Inbox", nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := dst.Import(bytes.NewReader(folderZip.Bytes()), inbox.ID); err != nil {
		log.Fatal(err)
	}
	list, err := dst.List("inbox")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("inbox folders", list.Folders)
}
