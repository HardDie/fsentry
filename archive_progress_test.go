package fsentry_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/HardDie/fsentry"
)

// gameArchive exports a folder "My Game" with one entry and one binary.
func gameArchive(t *testing.T) []byte {
	t.Helper()
	src := openDB(t)
	if _, err := src.CreateFolder("My Game", noteMeta{Color: "red"}); err != nil {
		t.Fatal(err)
	}
	if _, err := src.CreateEntry("Rules", noteBody{Text: "draw"}, "My Game"); err != nil {
		t.Fatal(err)
	}
	if err := src.CreateBinary("cover", []byte("PNG-bytes"), "my_game"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := src.ExportFolder(&buf, "My Game"); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func checkProgress(t *testing.T, got []fsentry.ImportProgress, wantFiles int) {
	t.Helper()
	if len(got) != wantFiles+1 {
		t.Fatalf("calls %d, want %d: %+v", len(got), wantFiles+1, got)
	}
	first, last := got[0], got[len(got)-1]
	if first.Files != 0 || first.Bytes != 0 || first.FilesTotal != wantFiles || first.BytesTotal <= 0 {
		t.Fatalf("first %+v", first)
	}
	if last.Files != last.FilesTotal || last.Bytes != last.BytesTotal {
		t.Fatalf("last %+v", last)
	}
	for i := 1; i < len(got); i++ {
		if got[i].Files != got[i-1].Files+1 || got[i].Bytes < got[i-1].Bytes {
			t.Fatalf("step %d: %+v after %+v", i, got[i], got[i-1])
		}
		if got[i].FilesTotal != first.FilesTotal || got[i].BytesTotal != first.BytesTotal {
			t.Fatalf("totals changed at %d: %+v", i, got[i])
		}
	}
}

func TestImportFolderWithProgress(t *testing.T) {
	data := gameArchive(t)
	tests := []struct {
		name   string
		rename string
		wantID string
	}{
		{name: "keep id", rename: "", wantID: "my_game"},
		// A rename rewrites .info.json, so its size comes from the new body.
		{name: "rename", rename: "Other Title", wantID: "other_title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := openDB(t)
			var got []fsentry.ImportProgress
			id, err := dst.ImportFolderWithProgress(bytes.NewReader(data), tt.rename,
				func(p fsentry.ImportProgress) { got = append(got, p) })
			if err != nil {
				t.Fatal(err)
			}
			if id != tt.wantID {
				t.Fatalf("id %q", id)
			}
			// .info.json, rules.json, cover.bin.
			checkProgress(t, got, 3)
		})
	}
}

func TestImportWithProgress(t *testing.T) {
	data := gameArchive(t)
	dst := openDB(t)
	var got []fsentry.ImportProgress
	if err := dst.ImportWithProgress(bytes.NewReader(data),
		func(p fsentry.ImportProgress) { got = append(got, p) }); err != nil {
		t.Fatal(err)
	}
	checkProgress(t, got, 3)
}

func TestImportProgressNotCalledOnRefusal(t *testing.T) {
	data := gameArchive(t)
	dst := openDB(t)
	if _, err := dst.ImportFolder(bytes.NewReader(data), ""); err != nil {
		t.Fatal(err)
	}
	calls := 0
	_, err := dst.ImportFolderWithProgress(bytes.NewReader(data), "",
		func(fsentry.ImportProgress) { calls++ })
	if !errors.Is(err, fsentry.ErrExist) {
		t.Fatalf("got %v", err)
	}
	if calls != 0 {
		t.Fatalf("progress called %d times", calls)
	}
}

func TestImportFolderNilProgress(t *testing.T) {
	data := gameArchive(t)
	dst := openDB(t)
	if _, err := dst.ImportFolderWithProgress(bytes.NewReader(data), "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dst.GetEntry[noteBody]("Rules", "my_game"); err != nil {
		t.Fatal(err)
	}
}
