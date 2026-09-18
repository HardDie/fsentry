package jsonutil

import (
	"bytes"
	"testing"
)

func TestEncodeCompact(t *testing.T) {
	got, err := Encode(map[string]int{"a": 1}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(got, []byte("\n")) {
		t.Fatalf("missing newline: %q", got)
	}
	if bytes.Contains(got, []byte("\t")) {
		t.Fatalf("unexpected tab: %q", got)
	}
}

func TestEncodePretty(t *testing.T) {
	got, err := Encode(map[string]int{"a": 1}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("\t")) {
		t.Fatalf("want tab indent: %q", got)
	}
}
