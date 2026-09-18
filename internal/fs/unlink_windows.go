//go:build windows

package fs

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func unlinkFile(path string) error {
	info, err := os.Lstat(path)
	if err == nil && info.IsDir() {
		return ErrIsDirectory
	}
	err = syscall.Unlink(path)
	if err == nil {
		return nil
	}
	if posixUnlink(path) == nil {
		return nil
	}
	return mapError(err)
}

// posixUnlink drops the name while other handles that opened with
// FILE_SHARE_DELETE keep the inode (Unix unlink). Used to steal a stale lock.
func posixUnlink(path string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(
		p,
		windows.DELETE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(h) }()

	flags := uint32(windows.FILE_DISPOSITION_DELETE | windows.FILE_DISPOSITION_POSIX_SEMANTICS)
	err = windows.SetFileInformationByHandle(
		h,
		windows.FileDispositionInfoEx,
		(*byte)(unsafe.Pointer(&flags)),
		uint32(unsafe.Sizeof(flags)),
	)
	if err == nil {
		return nil
	}
	var info struct{ DeleteFile byte }
	info.DeleteFile = 1
	return windows.SetFileInformationByHandle(
		h,
		windows.FileDispositionInfo,
		(*byte)(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
}
