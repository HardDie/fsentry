package fsentry_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/HardDie/fsentry"
)

func openStore(t *testing.T) (*fsentry.DB, string) {
	t.Helper()
	root := t.TempDir()
	db := fsentry.New(root, fsentry.WithNoLockFile())
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	return db, root
}

func TestValidateOK(t *testing.T) {
	db, _ := openStore(t)
	if _, err := db.CreateFolder("My Notes", noteMeta{Color: "blue"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("Welcome", noteBody{Text: "hi"}, "My Notes"); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateBinary("cover", []byte("PNG"), "my_notes"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("settings", noteBody{Text: "root"}); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 0 {
		t.Fatalf("%v", probs)
	}
	probs, err = db.Validate("My Notes")
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 0 {
		t.Fatalf("%v", probs)
	}
}

func TestValidateAfterImport(t *testing.T) {
	src := openDB(t)
	if _, err := src.CreateFolder("My Notes", noteMeta{Color: "blue"}); err != nil {
		t.Fatal(err)
	}
	if _, err := src.CreateEntry("Welcome", noteBody{Text: "hi"}, "My Notes"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := src.Export(&buf); err != nil {
		t.Fatal(err)
	}
	dst := openDB(t)
	if err := dst.Import(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}
	probs, err := dst.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 0 {
		t.Fatalf("%v", probs)
	}
}

func TestValidateFolderMissingInfo(t *testing.T) {
	db, root := openStore(t)
	if err := os.Mkdir(filepath.Join(root, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, probs, "broken", fsentry.ProblemInfo)
}

func TestValidateEntryJSON(t *testing.T) {
	db, root := openStore(t)
	if err := os.WriteFile(filepath.Join(root, "note.json"), []byte("not json"), 0o666); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, probs, "note.json", fsentry.ProblemJSON)
}

func TestValidateIDMismatch(t *testing.T) {
	db, root := openStore(t)
	if _, err := db.CreateFolder("My Notes", noteMeta{}); err != nil {
		t.Fatal(err)
	}
	info := filepath.Join(root, "my_notes", ".info.json")
	raw, err := os.ReadFile(info)
	if err != nil {
		t.Fatal(err)
	}
	patched := bytes.Replace(raw, []byte(`"id":"my_notes"`), []byte(`"id":"other"`), 1)
	if bytes.Equal(patched, raw) {
		t.Fatalf("id not found in %s", raw)
	}
	if err := os.WriteFile(info, patched, 0o666); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, probs, "my_notes", fsentry.ProblemID)
}

func TestValidateUnexpectedAndTemp(t *testing.T) {
	db, root := openStore(t)
	if err := os.WriteFile(filepath.Join(root, "readme.txt"), []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "settings.json.tmp"), []byte("{}"), 0o666); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, probs, "readme.txt", fsentry.ProblemUnexpected)
	assertProblem(t, probs, "settings.json.tmp", fsentry.ProblemTemp)
}

func TestValidateBadDiskName(t *testing.T) {
	db, root := openStore(t)
	if err := os.Mkdir(filepath.Join(root, "My Game"), 0o755); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, probs, "My Game", fsentry.ProblemBadName)
	assertProblem(t, probs, "My Game", fsentry.ProblemInfo)
}

func TestValidateNested(t *testing.T) {
	db, root := openStore(t)
	if _, err := db.CreateFolder("My Notes", noteMeta{}); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "my_notes", "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	probs, err := db.Validate("My Notes")
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, probs, "my_notes/broken", fsentry.ProblemInfo)
}

func TestValidateMissing(t *testing.T) {
	db := openDB(t)
	_, err := db.Validate("missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateCorruptedPath(t *testing.T) {
	db, root := openStore(t)
	if err := os.Mkdir(filepath.Join(root, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := db.Validate("broken")
	if !errors.Is(err, fsentry.ErrFolderCorrupted) {
		t.Fatalf("got %v, want ErrFolderCorrupted", err)
	}
}

func TestValidateSkipsLock(t *testing.T) {
	root := t.TempDir()
	db := fsentry.New(root)
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Drop() })
	probs, err := db.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 0 {
		t.Fatalf("%v", probs)
	}
}

func assertProblem(t *testing.T, probs []fsentry.Problem, path string, code fsentry.ProblemCode) {
	t.Helper()
	for _, p := range probs {
		if p.Path == path && p.Code == code {
			return
		}
	}
	t.Fatalf("missing %s %q in %v", code, path, probs)
}
