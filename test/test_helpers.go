// Package test provides shared testing utilities for Ship Commander 3
//
// This package contains common test helpers, fixtures, and utilities
// to ensure consistency across all test files.
//
// Reference: .spec/research-findings/GO_CODING_STANDARDS.md
package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Context returns a test context with timeout
func Context(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

// TempDir creates a temporary directory for testing
// The directory is automatically cleaned up when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sc3-test-*")
	require.NoError(t, err, "failed to create temp dir")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// TempFile creates a temporary file with the given content
// The file is automatically cleaned up when the test completes.
func TempFile(t *testing.T, content string) string {
	t.Helper()
	dir := TempDir(t)
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to write temp file")
	return path
}

// Chdir changes to a temporary directory for testing
// The original working directory is restored when the test completes.
func Chdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	err = os.Chdir(dir)
	require.NoError(t, err, "failed to change directory")

	t.Cleanup(func() {
		err := os.Chdir(original)
		assert.NoError(t, err, "failed to restore working directory")
	})
}

// RequirePanics ensures that the function panics
func RequirePanics(t *testing.T, f func()) {
	t.Helper()
	assert.Panics(t, f, "expected function to panic")
}

// RequireNoPanic ensures that the function does not panic
func RequireNoPanic(t *testing.T, f func()) {
	t.Helper()
	assert.NotPanics(t, f, "function should not panic")
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist: %s", path)
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.Error(t, err, "file should not exist: %s", path)
}

// AssertFileContent checks if a file has the expected content
func AssertFileContent(t *testing.T, path, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)
	assert.Equal(t, expectedContent, string(content), "file content mismatch")
}

// SkipIfShort skips the test if -short flag is provided
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
}

// TableDrivenTest is a helper for table-driven tests
// Example usage:
//
//	tests := []TableTest{
//	    {
//	        Name: "valid input",
//	        Input: "test",
//	        Want: "result",
//	    },
//	    {
//	        Name: "invalid input",
//	        Input: "",
//	        WantErr: true,
//	    },
//	}
//	TableDriven(t, tests, func(t *testing.T, tt TableTest) {
//	    result, err := Process(tt.Input)
//	    if tt.WantErr {
//	        assert.Error(t, err)
//	        return
//	    }
//	    assert.NoError(t, err)
//	    assert.Equal(t, tt.Want, result)
//	 })
type TableTest struct {
	Name    string
	Input   interface{}
	Want    interface{}
	WantErr bool
}

// TableDriven runs a table-driven test with the given test function
func TableDriven(t *testing.T, tests []TableTest, testFn func(*testing.T, TableTest)) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testFn(t, tt)
		})
	}
}

// AssertError checks if an error is of the expected type
func AssertError(t *testing.T, err error, expectedErr error) {
	t.Helper()
	assert.Equal(t, expectedErr, err, "error mismatch")
}

// AssertErrorIs checks if an error wraps the expected error
func AssertErrorIs(t *testing.T, err error, expectedErr error) {
	t.Helper()
	assert.ErrorIs(t, err, expectedErr, "error should wrap expected error")
}

// AssertErrorType checks if an error is of the expected type
func AssertErrorType(t *testing.T, err error, expectedType interface{}) {
	t.Helper()
	assert.ErrorAs(t, err, expectedType, "error should be of expected type")
}
