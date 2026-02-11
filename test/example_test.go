// Package test_example demonstrates table-driven testing patterns
//
// This file serves as a reference for writing idiomatic Go tests
// following the Ship Commander 3 coding standards.
//
// Reference: .spec/research-findings/GO_CODING_STANDARDS.md
package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example function to test
func Multiply(a, b int) int {
	return a * b
}

// Example function that returns an error
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Example function that processes strings
func ProcessString(input string) (string, error) {
	if input == "" {
		return "", errors.New("input cannot be empty")
	}
	if len(input) > 10 {
		return input[:10], nil
	}
	return input, nil
}

// TestMultiply demonstrates table-driven testing
//
// ✅ GOOD: Table-driven test (idiomatic Go pattern)
// This pattern is recommended for testing multiple input combinations
func TestMultiply(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "positive numbers",
			a:        3,
			b:        4,
			expected: 12,
		},
		{
			name:     "negative numbers",
			a:        -2,
			b:        5,
			expected: -10,
		},
		{
			name:     "zero",
			a:        100,
			b:        0,
			expected: 0,
		},
		{
			name:     "both negative",
			a:        -3,
			b:        -4,
			expected: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Multiply(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "Multiply(%d, %d)", tt.a, tt.b)
		})
	}
}

// TestDivide demonstrates table-driven testing with errors
//
// Shows how to handle error cases in table-driven tests
func TestDivide(t *testing.T) {
	tests := []struct {
		name      string
		a         int
		b         int
		expected  int
		wantError bool
	}{
		{
			name:      "simple division",
			a:         10,
			b:         2,
			expected:  5,
			wantError: false,
		},
		{
			name:      "division with remainder",
			a:         7,
			b:         3,
			expected:  2,
			wantError: false,
		},
		{
			name:      "division by zero",
			a:         5,
			b:         0,
			expected:  0,
			wantError: true,
		},
		{
			name:      "negative division",
			a:         -10,
			b:         2,
			expected:  -5,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Divide(tt.a, tt.b)

			if tt.wantError {
				assert.Error(t, err, "expected error for Divide(%d, %d)", tt.a, tt.b)
				return
			}

			require.NoError(t, err, "unexpected error for Divide(%d, %d)", tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "unexpected result")
		})
	}
}

// TestProcessString demonstrates advanced table-driven testing
//
// Shows how to test string processing with various edge cases
func TestProcessString(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   string
		wantError  bool
		errorCheck func(*testing.T, error) // Optional custom error check
	}{
		{
			name:      "short string",
			input:     "hello",
			expected:  "hello",
			wantError: false,
		},
		{
			name:      "exact 10 characters",
			input:     "0123456789",
			expected:  "0123456789",
			wantError: false,
		},
		{
			name:      "long string gets truncated",
			input:     "this is a very long string",
			expected:  "this is a ",
			wantError: false,
		},
		{
			name:      "empty string returns error",
			input:     "",
			expected:  "",
			wantError: true,
			errorCheck: func(t *testing.T, err error) {
				t.Helper()
				assert.Contains(t, err.Error(), "input cannot be empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessString(tt.input)

			if tt.wantError {
				require.Error(t, err, "expected error for input: %q", tt.input)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
				return
			}

			require.NoError(t, err, "unexpected error for input: %q", tt.input)
			assert.Equal(t, tt.expected, result, "unexpected result for input: %q", tt.input)
		})
	}
}

// TestUsingTestHelpers demonstrates using the test helpers package
//
// Shows how to use shared test utilities from the test package
func TestUsingTestHelpers(t *testing.T) {
	t.Run("temp dir", func(t *testing.T) {
		dir := TempDir(t)
		AssertFileExists(t, dir)
	})

	t.Run("temp file", func(t *testing.T) {
		content := "test content"
		filePath := TempFile(t, content)
		AssertFileExists(t, filePath)
		AssertFileContent(t, filePath, content)
	})

	t.Run("context", func(t *testing.T) {
		ctx := Context(t)
		assert.NotNil(t, ctx)
	})
}

// TestParallel demonstrates parallel test execution
//
// Shows how to safely run tests in parallel
func TestParallel(t *testing.T) {
	tests := []struct {
		name  string
		input int
	}{
		{name: "test 1", input: 1},
		{name: "test 2", input: 2},
		{name: "test 3", input: 3},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Mark this test as safe to run in parallel

			// Perform test operation
			result := Multiply(tt.input, 2)
			assert.Equal(t, tt.input*2, result)
		})
	}
}

// TestSkipInShortMode demonstrates skipping tests in short mode
//
// Useful for tests that are slow or have external dependencies
func TestSkipInShortMode(t *testing.T) {
	SkipIfShort(t)

	// This test only runs when -short flag is NOT provided
	// Example: integration test with external services
	t.Log("Running expensive test...")
}

// BenchmarkMultiply is an example benchmark
//
// Run with: go test -bench=. -benchmem
func BenchmarkMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Multiply(123, 456)
	}
}

// ExampleExample_test demonstrates how to write examples
//
// Examples are both documentation and tests
// Run with: go test
func ExampleMultiply() {
	result := Multiply(3, 4)
	fmt.Println(result)
	// Output: 12
}
