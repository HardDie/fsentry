package name

import (
	"testing"
	"unicode/utf8"
)

func TestAppend(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"My Game", "my_game"},
		{"my_game", "my_game"},
		{"", ""},
		{"   ", "___"},
		{"foo.bar", "foobar"},
		{"CON", ""},
		{"con", ""},
		{"prn", ""},
		{"aux", ""},
		{"nul", ""},
		{"com1", ""},
		{"lpt0", ""},
		{"conjson", "conjson"},
		{"Привет мир", "привет_мир"},
	}
	var buf []byte
	for _, tc := range cases {
		got := string(Append(buf[:0], tc.in))
		if got != tc.want {
			t.Fatalf("Append(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAppendTruncate(t *testing.T) {
	in := make([]byte, 0, 250)
	for range 250 {
		in = append(in, 'a')
	}
	got := Append(nil, string(in))
	if len(got) != 200 {
		t.Fatalf("len %d", len(got))
	}
}

func TestAppendReuseAllocs(t *testing.T) {
	buf := make([]byte, 0, 64)
	n := testing.AllocsPerRun(1000, func() {
		buf = Append(buf[:0], "My Game")
		if string(buf) != "my_game" {
			t.Fatal(string(buf))
		}
	})
	if n != 0 {
		t.Fatalf("allocs %v, want 0", n)
	}
}

func TestAppendValidUTF8AfterTruncate(t *testing.T) {
	// 200-byte cut may land mid-rune; match old byte truncation.
	r := 'ж'
	in := make([]byte, 0, 210)
	for len(in) < 201 {
		in = utf8.AppendRune(in, r)
	}
	got := Append(nil, string(in))
	if len(got) > 200 {
		t.Fatal(len(got))
	}
}
