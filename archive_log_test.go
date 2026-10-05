package fsentry_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/HardDie/fsentry"
)

// logLine is one JSON line of a slog handler.
type logLine struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Step  string `json:"step"`
	Entry string `json:"entry"`
	Err   string `json:"err"`
}

func readLog(t *testing.T, buf *bytes.Buffer) []logLine {
	t.Helper()
	var out []logLine
	sc := bufio.NewScanner(buf)
	for sc.Scan() {
		var l logLine
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			t.Fatalf("bad line %q: %v", sc.Text(), err)
		}
		out = append(out, l)
	}
	return out
}

func messages(lines []logLine) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, l.Msg)
	}
	return out
}

func openLoggedDB(t *testing.T) (*fsentry.DB, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	db := fsentry.New(t.TempDir(), fsentry.WithNoLockFile(), fsentry.WithLogger(log))
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	return db, &buf
}

func TestImportFolderLogsSteps(t *testing.T) {
	data := gameArchive(t)
	db, buf := openLoggedDB(t)
	if _, err := db.ImportFolder(bytes.NewReader(data), "Other Title"); err != nil {
		t.Fatal(err)
	}
	lines := readLog(t, buf)
	want := []string{
		"import folder: started",
		"import folder: locked",
		"import folder: zip read",
		"import folder: root found",
		"import folder: targets free",
		"import folder: file written",
		"import folder: file written",
		"import folder: file written",
		"import folder: finished",
	}
	got := messages(lines)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for _, l := range lines {
		if l.Level != "DEBUG" {
			t.Fatalf("%q at %s, want DEBUG", l.Msg, l.Level)
		}
		if l.Msg == "import folder: file written" && !strings.HasPrefix(l.Entry, "other_title/") {
			t.Fatalf("entry %q", l.Entry)
		}
	}
}

func TestImportLogsFailure(t *testing.T) {
	data := gameArchive(t)
	tests := []struct {
		name     string
		prepare  func(t *testing.T, db *fsentry.DB)
		zip      []byte
		wantErr  error
		wantMsg  string // the line that names the cause
		wantStep string
	}{
		{
			name:     "not a zip",
			zip:      []byte("not a zip"),
			wantErr:  fsentry.ErrBadArchive,
			wantMsg:  "import folder: not a zip",
			wantStep: "open zip",
		},
		{
			name: "game exists",
			prepare: func(t *testing.T, db *fsentry.DB) {
				if _, err := db.ImportFolder(bytes.NewReader(data), ""); err != nil {
					t.Fatal(err)
				}
			},
			zip:      data,
			wantErr:  fsentry.ErrExist,
			wantMsg:  "import folder: destination exists",
			wantStep: "check destination",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, buf := openLoggedDB(t)
			if tt.prepare != nil {
				tt.prepare(t, db)
				buf.Reset()
			}
			_, err := db.ImportFolder(bytes.NewReader(tt.zip), "")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v", err)
			}
			lines := readLog(t, buf)
			if !strings.Contains(strings.Join(messages(lines), "\n"), tt.wantMsg) {
				t.Fatalf("no %q in %v", tt.wantMsg, messages(lines))
			}
			last := lines[len(lines)-1]
			if last.Msg != "import folder: failed" || last.Step != tt.wantStep || last.Err == "" {
				t.Fatalf("last %+v", last)
			}
		})
	}
}
