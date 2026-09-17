// Package fs wraps operating-system file and directory operations.
//
// Callers in this module must use these helpers instead of os.OpenFile,
// os.Mkdir, os.File.Write/WriteAt/Read/ReadAt/Close/Sync/Truncate,
// os.Rename, os.Remove, os.RemoveAll, os.ReadDir, or os.Stat, and they must
// use OpenLock/Lock/TryLock/Unlock instead of flock / LockFileEx. Every helper
// maps platform errors (Windows, Linux, macOS) onto a fixed set of sentinels
// so the rest of the library can errors.Is without inspecting syscall.Errno.
//
// Layout follows package os: file.go (bytes and names), dir.go (directories),
// lock.go plus lock_unix.go / lock_windows.go, error.go plus mapOS per GOOS.
package fs
