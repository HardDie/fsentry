package fsentry_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/fsentry"
)

type noteMeta struct {
	Color string `json:"color"`
}

type noteBody struct {
	Text string `json:"text"`
}

func TestExportImportRoundTrip(t *testing.T) {
	src := openDB(t)
	if _, err := src.CreateFolder("My Notes", noteMeta{Color: "blue"}); err != nil {
		t.Fatal(err)
	}
	if _, err := src.CreateEntry("Welcome", noteBody{Text: "hello"}, "My Notes"); err != nil {
		t.Fatal(err)
	}
	if err := src.CreateBinary("cover", []byte("PNG"), "my_notes"); err != nil {
		t.Fatal(err)
	}
	if _, err := src.CreateEntry("settings", noteBody{Text: "root"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := src.Export(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty zip")
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if strings.Contains(f.Name, "\\") {
			t.Fatalf("zip name has backslash: %q", f.Name)
		}
		if filepath.Base(f.Name) == ".fsentry.lock" {
			t.Fatal("lock file in zip")
		}
	}

	dst := openDB(t)
	if err := dst.Import(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}

	folder, err := dst.GetFolder[noteMeta]("My Notes")
	if err != nil {
		t.Fatal(err)
	}
	if folder.Data.Color != "blue" {
		t.Fatalf("%+v", folder)
	}
	ent, err := dst.GetEntry[noteBody]("Welcome", "my_notes")
	if err != nil {
		t.Fatal(err)
	}
	if ent.Data.Text != "hello" {
		t.Fatalf("%+v", ent)
	}
	bin, err := dst.GetBinary("cover", nil, "my_notes")
	if err != nil {
		t.Fatal(err)
	}
	if string(bin) != "PNG" {
		t.Fatalf("%q", bin)
	}
	rootEnt, err := dst.GetEntry[noteBody]("settings")
	if err != nil {
		t.Fatal(err)
	}
	if rootEnt.Data.Text != "root" {
		t.Fatalf("%+v", rootEnt)
	}
}

func TestExportFolderPrefix(t *testing.T) {
	db := openDB(t)
	if _, err := db.CreateFolder("My Notes", noteMeta{Color: "blue"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("Welcome", noteBody{Text: "hello"}, "My Notes"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := db.Export(&buf, "My Notes"); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) == 0 {
		t.Fatal("empty")
	}
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "my_notes/") {
			t.Fatalf("want my_notes/ prefix, got %q", f.Name)
		}
	}

	parent := openDB(t)
	if _, err := parent.CreateFolder("Inbox", noteMeta{}); err != nil {
		t.Fatal(err)
	}
	if err := parent.Import(bytes.NewReader(buf.Bytes()), "Inbox"); err != nil {
		t.Fatal(err)
	}
	got, err := parent.GetFolder[noteMeta]("My Notes", "inbox")
	if err != nil {
		t.Fatal(err)
	}
	if got.Data.Color != "blue" {
		t.Fatalf("%+v", got)
	}
}

func TestImportExist(t *testing.T) {
	src := openDB(t)
	if _, err := src.CreateEntry("settings", noteBody{Text: "a"}); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := src.Export(&buf); err != nil {
		t.Fatal(err)
	}
	dst := openDB(t)
	if _, err := dst.CreateEntry("settings", noteBody{Text: "b"}); err != nil {
		t.Fatal(err)
	}
	err := dst.Import(bytes.NewReader(buf.Bytes()))
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("got %v", err)
	}
	got, err := dst.GetEntry[noteBody]("settings")
	if err != nil {
		t.Fatal(err)
	}
	if got.Data.Text != "b" {
		t.Fatalf("overwrote: %+v", got)
	}
}

func TestImportZipSlip(t *testing.T) {
	db := openDB(t)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("../outside.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(w, "{}"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	err = db.Import(bytes.NewReader(buf.Bytes()))
	if !errors.Is(err, fsentry.ErrBadArchive) {
		t.Fatalf("got %v", err)
	}
}

func TestImportBadZip(t *testing.T) {
	db := openDB(t)
	err := db.Import(bytes.NewReader([]byte("not a zip")))
	if !errors.Is(err, fsentry.ErrBadArchive) {
		t.Fatalf("got %v", err)
	}
}

func TestImportNil(t *testing.T) {
	db := openDB(t)
	if err := db.Import(nil); !errors.Is(err, fsentry.ErrBadArchive) {
		t.Fatalf("got %v", err)
	}
}

func TestImportRollback(t *testing.T) {
	db := openDB(t)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("ok.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(w, `{"id":"ok","name":"\"ok\"","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","data":{}}`); err != nil {
		t.Fatal(err)
	}
	w, err = zw.Create("ok.json/nested.bin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	err = db.Import(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("expected error")
	}
	if _, err := db.GetEntry[noteBody]("ok"); !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("ok.json should be rolled back: %v", err)
	}
}

func TestExportMissingFolder(t *testing.T) {
	db := openDB(t)
	var buf bytes.Buffer
	err := db.Export(&buf, "missing")
	if !errors.Is(err, fsentry.ErrBadPath) && !errors.Is(err, fsentry.ErrNotExist) {
		t.Fatalf("got %v", err)
	}
}

func TestExportSkipsLockFile(t *testing.T) {
	dir := t.TempDir()
	db := fsentry.New(dir)
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateEntry("settings", noteBody{Text: "x"}); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := db.Export(&buf); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if strings.Contains(f.Name, ".fsentry.lock") {
			t.Fatalf("lock in zip: %q", f.Name)
		}
	}
}
