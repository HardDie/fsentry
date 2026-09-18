package fsentry_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HardDie/fsentry"
)

func exampleDB() (*fsentry.DB, string, func()) {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		_ = os.RemoveAll(dir)
		panic(err)
	}
	return db, dir, func() { _ = os.RemoveAll(dir) }
}

type meta struct {
	Kind string `json:"kind"`
}

type note struct {
	Body string `json:"body"`
}

type exampleLog struct{}

func (exampleLog) Debug(string, ...any) {}
func (exampleLog) Info(string, ...any)  {}
func (exampleLog) Warn(string, ...any)  {}
func (exampleLog) Error(string, ...any) {}

func Example() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	folder, err := db.CreateFolder("My Notes", meta{Kind: "notebook"})
	if err != nil {
		panic(err)
	}
	_, err = db.CreateEntry("Welcome", note{Body: "hello"}, folder.ID)
	if err != nil {
		panic(err)
	}
	if err := db.CreateBinary("cover", []byte("PNG"), folder.ID); err != nil {
		panic(err)
	}
	list, err := db.List(folder.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println(folder.ID, list.Entries, list.Binaries)
	// Output:
	// my_notes [welcome] [cover]
}

func ExampleNameToID() {
	fmt.Println(fsentry.NameToID("My Notes"))
	fmt.Printf("%q\n", fsentry.NameToID("con"))
	// Output:
	// my_notes
	// ""
}

func ExampleNew() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithPretty(), fsentry.WithNoLockFile())
	fmt.Println(db != nil)
	// Output:
	// true
}

func ExampleWithPretty() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile(), fsentry.WithPretty())
	if err := db.Init(); err != nil {
		panic(err)
	}
	if _, err := db.CreateEntry("settings", note{Body: "x"}); err != nil {
		panic(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(string(raw), "\n\t\"id\""))
	// Output:
	// true
}

func ExampleWithLogger() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile(), fsentry.WithLogger(exampleLog{}))
	if err := db.Init(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleWithNoLockFile() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		panic(err)
	}
	_, err = os.Stat(filepath.Join(dir, ".fsentry.lock"))
	fmt.Println(errors.Is(err, os.ErrNotExist))
	// Output:
	// true
}

func ExampleWithLockFile() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile(), fsentry.WithLockFile())
	if err := db.Init(); err != nil {
		panic(err)
	}
	_, err = os.Stat(filepath.Join(dir, ".fsentry.lock"))
	fmt.Println(err == nil)
	// Output:
	// true
}

func ExampleWithLockTimeout() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithLockFile(), fsentry.WithLockTimeout(time.Minute))
	if err := db.Init(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleQuotedString_String() {
	s := fsentry.QuotedString("My Notes")
	fmt.Println(s.String())
	// Output:
	// My Notes
}

func ExampleQuotedString_MarshalJSON() {
	raw, err := json.Marshal(fsentry.QuotedString("My Notes"))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(raw))
	// Output:
	// "\"My Notes\""
}

func ExampleQuotedString_UnmarshalJSON() {
	var s fsentry.QuotedString
	if err := json.Unmarshal([]byte(`"\"My Notes\""`), &s); err != nil {
		panic(err)
	}
	fmt.Println(s.String())
	// Output:
	// My Notes
}

func ExampleProblem_String() {
	p := fsentry.Problem{Path: "broken", Code: fsentry.ProblemInfo}
	fmt.Println(p.String())
	// Output:
	// info: broken
}

func ExampleDB_Init() {
	dir, err := os.MkdirTemp("", "fsentry-ex-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	db := fsentry.New(dir, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		panic(err)
	}
	info, err := os.Stat(dir)
	fmt.Println(err == nil && info.IsDir())
	// Output:
	// true
}

func ExampleDB_Drop() {
	db, dir, cleanup := exampleDB()
	defer cleanup()

	if err := db.Drop(); err != nil {
		panic(err)
	}
	_, err := os.Stat(dir)
	fmt.Println(errors.Is(err, os.ErrNotExist))
	// Output:
	// true
}

func ExampleDB_List() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateFolder("My Notes", meta{}); err != nil {
		panic(err)
	}
	list, err := db.List()
	if err != nil {
		panic(err)
	}
	fmt.Println(list.Folders)
	// Output:
	// [my_notes]
}

func ExampleDB_CreateFolder() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	folder, err := db.CreateFolder("My Notes", meta{Kind: "notebook"})
	if err != nil {
		panic(err)
	}
	fmt.Println(folder.ID, folder.Name, folder.Data.Kind)
	// Output:
	// my_notes My Notes notebook
}

func ExampleDB_GetFolder() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateFolder("My Notes", meta{Kind: "notebook"}); err != nil {
		panic(err)
	}
	got, err := db.GetFolder[meta]("My Notes")
	if err != nil {
		panic(err)
	}
	fmt.Println(got.ID, got.Data.Kind)
	// Output:
	// my_notes notebook
}

func ExampleDB_MoveFolder() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateFolder("My Notes", meta{Kind: "notebook"}); err != nil {
		panic(err)
	}
	moved, err := db.MoveFolder[meta]("My Notes", "Archive")
	if err != nil {
		panic(err)
	}
	fmt.Println(moved.ID, moved.Name)
	// Output:
	// archive Archive
}

func ExampleDB_UpdateFolder() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateFolder("My Notes", meta{Kind: "old"}); err != nil {
		panic(err)
	}
	updated, err := db.UpdateFolder("My Notes", meta{Kind: "new"})
	if err != nil {
		panic(err)
	}
	fmt.Println(updated.Data.Kind)
	// Output:
	// new
}

func ExampleDB_RemoveFolder() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateFolder("My Notes", meta{}); err != nil {
		panic(err)
	}
	if err := db.RemoveFolder("My Notes"); err != nil {
		panic(err)
	}
	_, err := db.GetFolder[meta]("My Notes")
	fmt.Println(errors.Is(err, fsentry.ErrNotExist))
	// Output:
	// true
}

func ExampleDB_CreateEntry() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	ent, err := db.CreateEntry("settings", note{Body: "hello"})
	if err != nil {
		panic(err)
	}
	fmt.Println(ent.ID, ent.Data.Body)
	// Output:
	// settings hello
}

func ExampleDB_GetEntry() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateEntry("settings", note{Body: "hello"}); err != nil {
		panic(err)
	}
	got, err := db.GetEntry[note]("settings")
	if err != nil {
		panic(err)
	}
	fmt.Println(got.Data.Body)
	// Output:
	// hello
}

func ExampleDB_MoveEntry() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateEntry("settings", note{Body: "hello"}); err != nil {
		panic(err)
	}
	moved, err := db.MoveEntry[note]("settings", "prefs")
	if err != nil {
		panic(err)
	}
	fmt.Println(moved.ID)
	// Output:
	// prefs
}

func ExampleDB_UpdateEntry() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateEntry("settings", note{Body: "old"}); err != nil {
		panic(err)
	}
	updated, err := db.UpdateEntry("settings", note{Body: "new"})
	if err != nil {
		panic(err)
	}
	fmt.Println(updated.Data.Body)
	// Output:
	// new
}

func ExampleDB_DuplicateEntry() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateEntry("settings", note{Body: "hello"}); err != nil {
		panic(err)
	}
	copy, err := db.DuplicateEntry[note]("settings", "backup")
	if err != nil {
		panic(err)
	}
	fmt.Println(copy.ID, copy.Data.Body)
	// Output:
	// backup hello
}

func ExampleDB_RemoveEntry() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateEntry("settings", note{Body: "hello"}); err != nil {
		panic(err)
	}
	if err := db.RemoveEntry("settings"); err != nil {
		panic(err)
	}
	_, err := db.GetEntry[note]("settings")
	fmt.Println(errors.Is(err, fsentry.ErrNotExist))
	// Output:
	// true
}

func ExampleDB_CreateBinary() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	err := db.CreateBinary("cover", []byte("PNG"))
	fmt.Println(err == nil)
	// Output:
	// true
}

func ExampleDB_GetBinary() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if err := db.CreateBinary("cover", []byte("PNG")); err != nil {
		panic(err)
	}
	data, err := db.GetBinary("cover", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// Output:
	// PNG
}

func ExampleDB_MoveBinary() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if err := db.CreateBinary("cover", []byte("PNG")); err != nil {
		panic(err)
	}
	if err := db.MoveBinary("cover", "art"); err != nil {
		panic(err)
	}
	data, err := db.GetBinary("art", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// Output:
	// PNG
}

func ExampleDB_UpdateBinary() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if err := db.CreateBinary("cover", []byte("old")); err != nil {
		panic(err)
	}
	if err := db.UpdateBinary("cover", []byte("new")); err != nil {
		panic(err)
	}
	data, err := db.GetBinary("cover", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// Output:
	// new
}

func ExampleDB_RemoveBinary() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if err := db.CreateBinary("cover", []byte("PNG")); err != nil {
		panic(err)
	}
	if err := db.RemoveBinary("cover"); err != nil {
		panic(err)
	}
	_, err := db.GetBinary("cover", nil)
	fmt.Println(errors.Is(err, fsentry.ErrNotExist))
	// Output:
	// true
}

func ExampleDB_Export() {
	db, _, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateEntry("settings", note{Body: "hello"}); err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := db.Export(&buf); err != nil {
		panic(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		panic(err)
	}
	fmt.Println(zr.File[0].Name)
	// Output:
	// settings.json
}

func ExampleDB_Import() {
	src, _, cleanupSrc := exampleDB()
	defer cleanupSrc()
	if _, err := src.CreateEntry("settings", note{Body: "hello"}); err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := src.Export(&buf); err != nil {
		panic(err)
	}

	dst, _, cleanupDst := exampleDB()
	defer cleanupDst()
	if err := dst.Import(bytes.NewReader(buf.Bytes())); err != nil {
		panic(err)
	}
	got, err := dst.GetEntry[note]("settings")
	if err != nil {
		panic(err)
	}
	fmt.Println(got.Data.Body)
	// Output:
	// hello
}

func ExampleDB_Validate() {
	db, dir, cleanup := exampleDB()
	defer cleanup()

	if _, err := db.CreateFolder("My Notes", meta{}); err != nil {
		panic(err)
	}
	ok, err := db.Validate()
	if err != nil {
		panic(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "broken"), 0o755); err != nil {
		panic(err)
	}
	bad, err := db.Validate()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(ok), bad[0].String())
	// Output:
	// 0 info: broken
}
