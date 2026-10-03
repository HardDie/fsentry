package fsentry_test

import (
	"testing"
	"time"

	"github.com/HardDie/fsentry"
)

var (
	createdAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	movedAt   = createdAt.Add(time.Hour)
)

func openDBAt(t *testing.T, now *time.Time) *fsentry.DB {
	t.Helper()
	db := openDB(t)
	fsentry.SetClock(db, func() time.Time { return *now })
	return db
}

func TestMoveFolderSameID(t *testing.T) {
	tests := []struct {
		name    string
		oldName string
		newName string
	}{
		{name: "case", oldName: "My Game", newName: "my game"},
		{name: "punctuation", oldName: "Catan", newName: "Catan!"},
		{name: "identical", oldName: "Catan", newName: "Catan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := createdAt
			db := openDBAt(t, &now)
			src, err := db.CreateFolder(tt.oldName, map[string]int{"n": 1})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.CreateEntry[any]("child", nil, src.ID); err != nil {
				t.Fatal(err)
			}
			now = movedAt
			moved, err := db.MoveFolder[map[string]int](tt.oldName, tt.newName)
			if err != nil {
				t.Fatal(err)
			}
			if moved.ID != src.ID || moved.Name != tt.newName || moved.Data["n"] != 1 {
				t.Fatalf("%+v", moved)
			}
			if !moved.CreatedAt.Equal(createdAt) || !moved.UpdatedAt.Equal(movedAt) {
				t.Fatalf("timestamps: %+v", moved)
			}
			got, err := db.GetFolder[map[string]int](src.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got.Name != tt.newName || !got.UpdatedAt.Equal(movedAt) {
				t.Fatalf("%+v", got)
			}
			if _, err := db.GetEntry[any]("child", src.ID); err != nil {
				t.Fatalf("child lost: %v", err)
			}
		})
	}
}

func TestUpdateFolderNameWithoutTimestampSameID(t *testing.T) {
	now := createdAt
	db := openDBAt(t, &now)
	src, err := db.CreateFolder[any]("My Game", nil)
	if err != nil {
		t.Fatal(err)
	}
	now = movedAt
	moved, err := db.UpdateFolderNameWithoutTimestamp[any]("My Game", "my game!")
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != src.ID || moved.Name != "my game!" {
		t.Fatalf("%+v", moved)
	}
	got, err := db.GetFolder[any](src.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "my game!" || !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(createdAt) {
		t.Fatalf("%+v", got)
	}
}

func TestMoveEntrySameID(t *testing.T) {
	now := createdAt
	db := openDBAt(t, &now)
	src, err := db.CreateEntry("Note", 7)
	if err != nil {
		t.Fatal(err)
	}
	now = movedAt
	moved, err := db.MoveEntry[int]("Note", "NOTE!")
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != src.ID || moved.Name != "NOTE!" || moved.Data != 7 {
		t.Fatalf("%+v", moved)
	}
	got, err := db.GetEntry[int](src.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "NOTE!" || !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(movedAt) {
		t.Fatalf("%+v", got)
	}
}

func TestMoveBinarySameID(t *testing.T) {
	db := openDB(t)
	if err := db.CreateBinary("Cover", []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := db.MoveBinary("Cover", "COVER"); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetBinary("cover", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "a" {
		t.Fatalf("%q", got)
	}
}
