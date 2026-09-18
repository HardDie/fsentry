//go:build unix

package fs

import "syscall"

func unlinkFile(path string) error {
	err := syscall.Unlink(path)
	if err != nil {
		return mapError(err)
	}
	return nil
}
