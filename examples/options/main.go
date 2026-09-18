// Options: pretty JSON, lock file off vs on, lock steal timeout.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HardDie/fsentry"
)

func main() {
	prettyDir, err := os.MkdirTemp("", "fsentry-pretty-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(prettyDir)

	pretty := fsentry.New(prettyDir, fsentry.WithNoLockFile(), fsentry.WithPretty())
	if err := pretty.Init(); err != nil {
		log.Fatal(err)
	}
	if _, err := pretty.CreateEntry("settings", map[string]int{"n": 1}); err != nil {
		log.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(prettyDir, "settings.json"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("pretty indent", strings.Contains(string(raw), "\n\t"))

	lockedDir, err := os.MkdirTemp("", "fsentry-lock-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(lockedDir)

	locked := fsentry.New(lockedDir, fsentry.WithLockFile(), fsentry.WithLockTimeout(time.Minute))
	if err := locked.Init(); err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(filepath.Join(lockedDir, ".fsentry.lock"))
	fmt.Println("lock file", err == nil)

	openDir, err := os.MkdirTemp("", "fsentry-nolock-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(openDir)
	open := fsentry.New(openDir, fsentry.WithNoLockFile())
	if err := open.Init(); err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(filepath.Join(openDir, ".fsentry.lock"))
	fmt.Println("no lock file", os.IsNotExist(err))
}
