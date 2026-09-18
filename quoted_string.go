package fsentry

import (
	"encoding/json"
	"strconv"
)

// QuotedString is a display name stored as a JSON string whose contents are a
// quoted Go string (double-encoded). Existing DeckBuilder files use this form.
type QuotedString string

// MarshalJSON writes a JSON string of strconv.Quote(s).
func (s QuotedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Quote(string(s)))
}

// UnmarshalJSON reads a JSON string and strconv.Unquote's it.
func (s *QuotedString) UnmarshalJSON(data []byte) error {
	if s == nil {
		return nil
	}
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	unquoted, err := strconv.Unquote(val)
	if err != nil {
		return err
	}
	*s = QuotedString(unquoted)
	return nil
}

// String returns the display name.
func (s QuotedString) String() string {
	return string(s)
}
