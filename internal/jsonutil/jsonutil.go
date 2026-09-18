// Package jsonutil encodes JSON envelopes into a buffer.
package jsonutil

import (
	"bytes"
	"encoding/json"
)

// Encode marshals v. Pretty uses a tab indent. A trailing newline is always
// written (encoding/json Encoder), matching existing on-disk files.
func Encode(v any, pretty bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if pretty {
		enc.SetIndent("", "\t")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
