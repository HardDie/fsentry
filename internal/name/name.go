// Package name turns a display name into a portable filesystem ID.
package name

import (
	"unicode"
	"unicode/utf8"
)

const maxID = 200

// Append lowercases name, maps spaces to '_', drops other non-letter/digit/underscore
// runes, truncates to 200 bytes, and appends into dst. An empty or reserved result
// is a zero-length slice (not an error).
func Append(dst []byte, name string) []byte {
	dst = dst[:0]
	for _, r := range name {
		if r == ' ' {
			dst = append(dst, '_')
			continue
		}
		r = unicode.ToLower(r)
		if r == '_' || (r >= '0' && r <= '9') || unicode.IsLetter(r) {
			dst = utf8.AppendRune(dst, r)
		}
	}
	if len(dst) > maxID {
		dst = dst[:maxID]
	}
	if reservedID(dst) {
		return dst[:0]
	}
	return dst
}

func reservedID(id []byte) bool {
	switch string(id) {
	case "con", "prn", "aux", "nul",
		"com0", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9",
		"lpt0", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9":
		return true
	default:
		return false
	}
}
