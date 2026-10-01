package fsentry_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/HardDie/fsentry"
)

func TestEnsurePathMissingSegment(t *testing.T) {
	db := openDB(t)

	_, err := db.CreateFolder[any]("child", nil, "missing")
	assertMissingParent(t, err, "missing")

	if _, err := db.CreateFolder[any]("Games", nil); err != nil {
		t.Fatal(err)
	}
	_, err = db.GetFolder[any]("deck", "Games", "Missing Collection", "later")
	assertMissingParent(t, err, "Games", "Missing Collection")

	_, err = db.CreateEntry("note", struct{}{}, "..")
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("%v", err)
	}
	var pe *fsentry.BadPathError
	if errors.As(err, &pe) {
		t.Fatalf("invalid segment named a path: %#v", pe)
	}
}

func assertMissingParent(t *testing.T, err error, segs ...string) {
	t.Helper()
	var pe *fsentry.BadPathError
	if !errors.As(err, &pe) {
		t.Fatalf("got %v", err)
	}
	if !errors.Is(err, fsentry.ErrBadPath) {
		t.Fatalf("sentinel: %v", err)
	}
	if pe.Unwrap() != fsentry.ErrBadPath {
		t.Fatalf("unwrap %v", pe.Unwrap())
	}
	if len(pe.Path) != len(segs) {
		t.Fatalf("path %#v", pe.Path)
	}
	for i := range segs {
		if pe.Path[i] != segs[i] {
			t.Fatalf("path %#v", pe.Path)
		}
	}
	want := "bad path: " + strings.Join(segs, "/")
	if pe.Error() != want {
		t.Fatalf("error %q", pe.Error())
	}
}
