package fsentry

import nameto "github.com/HardDie/fsentry/internal/name"

// NameToID sanitizes a user name into a filesystem ID (lowercase, spaces to
// '_', letters/digits/underscore only, max 200 bytes). Empty or a Windows
// reserved device name yields "".
func NameToID(name string) string {
	return string(nameto.Append(nil, name))
}
