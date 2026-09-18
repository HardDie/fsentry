package fsentry

import (
	"encoding/json"
	"testing"
)

func TestQuotedStringRoundTrip(t *testing.T) {
	in := QuotedString("My Game")
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `"\"My Game\""` {
		t.Fatalf("got %s", raw)
	}
	var out QuotedString
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "My Game" {
		t.Fatalf("got %q", out)
	}
}
