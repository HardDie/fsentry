// Package fsentry treats a directory tree as a small database.
//
// Callers store folders, JSON entries, and opaque binaries on disk under a
// single root. There is no SQL server and no DeckBuilder domain types.
// Open a store with New, then Init before other operations.
//
// Godoc includes a runnable example for every exported function and method.
//
// License: GNU General Public License v3.0. See the LICENSE file in the
// module root.
package fsentry
