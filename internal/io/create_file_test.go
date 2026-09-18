package io

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func TestCreateFile(t *testing.T) {
	// Group 1: Tests using the actual filesystem.
	// These tests don't require mocking.

	t.Run("Success Case", func(t *testing.T) {
		tempDir := t.TempDir() // Creates a temporary directory that is automatically cleaned up.
		filePath := filepath.Join(tempDir, "testfile.txt")

		file, err := CreateFile(filePath)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if file == nil {
			t.Fatal("Expected a valid file handle, but got nil")
		}
		defer file.Close()

		// Verify the file was actually created.
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			t.Error("File was not created on disk")
		}
	})

	t.Run("Error When File Exists", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "existing.txt")
		// Pre-create the file.
		err := os.WriteFile(filePath, []byte("hello"), 0666)
		if err != nil {
			t.Fatalf("Can't pre-create file 'existing.txt': %v", err)
		}

		_, err = CreateFile(filePath)
		if !errors.Is(err, ErrorExist) {
			t.Errorf("Expected ErrorExist, but got: %v", err)
		}
	})

	t.Run("Error When Parent Directory Does Not Exist", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "nonexistent", "file.txt")

		_, err := CreateFile(filePath)
		if !errors.Is(err, ErrorNotExist) {
			t.Errorf("Expected ErrorNotExist, but got: %v", err)
		}
	})

	t.Run("Error When Path Component Is A File", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create a file where a directory should be in the path.
		parentFilePath := filepath.Join(tempDir, "a_file.txt")
		err := os.WriteFile(parentFilePath, []byte("i am a file"), 0666)
		if err != nil {
			t.Fatalf("Can't pre-create file 'a_file.txt': %v", err)
		}

		filePath := filepath.Join(parentFilePath, "another_file.txt")
		_, err = CreateFile(filePath)
		if !errors.Is(err, ErrorNotDirectory) {
			t.Errorf("Expected ErrorNotDirectory, but got: %v", err)
		}
	})

	t.Run("Error On Permission Denied", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows due to different permission model.")
		}
		// Create a directory with read-only permissions.
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")
		os.Mkdir(readOnlyDir, 0555) // Read and execute only

		filePath := filepath.Join(readOnlyDir, "test.txt")
		_, err := CreateFile(filePath)
		if !errors.Is(err, ErrorPermissions) {
			t.Errorf("Expected ErrorPermissions, but got: %v", err)
		}
	})

	// Group 2: Tests using a mocked os.OpenFile.
	// This allows us to simulate errors that are hard to create reliably.

	// mockOpenFileWithError is a helper that returns a function simulating os.OpenFile
	// but always returning a specific error.
	mockOpenFileWithError := func(retErr error) func(string, int, os.FileMode) (*os.File, error) {
		return func(name string, flag int, perm os.FileMode) (*os.File, error) {
			// Real os functions wrap errors in os.PathError, so we do too for accuracy.
			return nil, &os.PathError{Op: "open", Path: name, Err: retErr}
		}
	}

	mockTestCases := []struct {
		name         string
		simulatedErr error // The OS-level error we are simulating.
		expectedErr  error // The custom error we expect our function to return.
	}{
		{"Error Is Directory", syscall.EISDIR, ErrorIsDirectory},
		{"Error No Space", syscall.ENOSPC, ErrorNoSpace},
		{"Error Read Only FS", syscall.EROFS, ErrorReadOnlyFS},
		{"Error Internal (Unknown)", fmt.Errorf("a weird cosmic ray error"), ErrorInternal},
	}

	for _, tc := range mockTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- MOCKING ---
			originalOpenFile := osOpenFile                      // 1. Save the original function.
			osOpenFile = mockOpenFileWithError(tc.simulatedErr) // 2. Replace with the mock.
			defer func() { osOpenFile = originalOpenFile }()    // 3. Restore it after the test.

			// --- ACTION ---
			_, err := CreateFile("any/fake/path")

			// --- ASSERTION ---
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectedErr, err)
			}
		})
	}
}
