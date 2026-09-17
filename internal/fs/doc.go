// Package fs wraps operating-system file and directory operations.
//
// Callers in this module must use these helpers instead of os.OpenFile,
// os.Mkdir, os.File.Write, os.File.Read, os.File.Close, os.File.Sync,
// os.Rename, os.Remove, os.RemoveAll, os.ReadDir, or os.Stat, and they must
// use OpenLock/Lock/TryLock/Unlock instead of flock / LockFileEx. Every helper
// maps platform errors (Windows, Linux, macOS) onto a fixed set of sentinels
// so the rest of the library can errors.Is without inspecting syscall.Errno.
package fs
