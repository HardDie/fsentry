// Package fs wraps operating-system file and directory creation.
//
// Callers in this module must use these helpers instead of os.OpenFile or
// os.Mkdir. Every helper maps platform errors (Windows, Linux, macOS) onto
// a fixed set of sentinels so the rest of the library can errors.Is without
// inspecting syscall.Errno.
package fs
