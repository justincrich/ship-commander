# Test Package - Agent Guidance

## ⚠️ MANDATORY READING BEFORE WRITING TESTS

**ALL agents MUST read**: `TESTING.md` (root directory)

This document contains:
- **5 Test Integrity Rules** - Prevent vanity/self-validating tests
- **Go-specific anti-patterns** - What NOT to do in Go tests
- **Mocking strategy** - When to mock vs. use real implementations
- **Test quality checklist** - Verify tests are valuable
- **Given-When-Then template** - Structure tests for clarity

### Quick Anti-Pattern Reference

| ❌ VANITY (Don't Do) | ✅ VALUABLE (Do This) |
|---------------------|---------------------|
| Create struct, assert same fields | Verify DB state after operation |
| `mock.AssertCalled()` | Use test DB, verify real persistence |
| Test struct field exists | Test operation that uses the field |
| Arbitrary test data | Realistic production-like data |

## Purpose

This package provides shared testing utilities, fixtures, and examples for Ship Commander 3. It ensures consistency across all test files and demonstrates idiomatic Go testing patterns.

## Context in Ship Commander 3

The test package is **NOT for production code**—it's a collection of:
- Shared test helpers (functions used across multiple test files)
- Test fixtures (sample data for testing)
- Example tests demonstrating best practices
- Integration test utilities

## Package Organization

```
test/
├── AGENTS.md          # This file
├── test_helpers.go    # Shared test utility functions
├── example_test.go    # Example tests demonstrating patterns
├── fixtures/          # Test data and fixtures
│   ├── commissions/   # Sample commission files
│   ├── prds/          # Sample PRD files
│   └── configs/       # Sample configuration files
└── integration/       # Integration tests
    └── README.md      # Integration test guidelines
```

## Key Testing Principles

### 1. Table-Driven Tests (Preferred Pattern)

```go
// ✅ GOOD: Table-driven test (idiomatic Go)
func TestSplitHostPort(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantHost string
        wantPort string
        wantErr  bool
    }{
        {
            name:     "valid IPv4 with port",
            input:    "192.0.2.0:8000",
            wantHost: "192.0.2.0",
            wantPort: "8000",
            wantErr:  false,
        },
        {
            name:    "missing port",
            input:   "192.0.2.0",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            host, port, err := net.SplitHostPort(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.wantHost, host)
            assert.Equal(t, tt.wantPort, port)
        })
    }
}
```

### 2. Test File Organization

| Pattern | File Name | Package | Usage |
|---------|-----------|---------|-------|
| **Internal tests** | `xxx_test.go` | `package mypackage` | Tests that access unexported code |
| **External tests** | `xxx_external_test.go` | `package mypackage_test` | Tests black-box behavior only |
| **Example tests** | `example_test.go` | `package mypackage` | Executable examples for Godoc |

### 3. Test Package Naming

```go
// ✅ GOOD: Internal test (access unexported)
package handler

func TestHandler(t *testing.T) {
    h := &Handler{} // Can access unexported fields
    // ...
}

// ✅ GOOD: External test (black-box only)
package handler_test

func TestHandler(t *testing.T) {
    h := NewHandler() // Can only call exported functions
    // ...
}
```

### 4. Testing Libraries

Use `github.com/stretchr/testify` for assertions and mocking:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"  // Non-fatal assertions
    "github.com/stretchr/testify/require" // Fatal assertions (stop test)
    "github.com/stretchr/testify/mock"    // Mock objects
)

// Use assert for non-fatal checks
assert.Equal(t, expected, actual, "values should match")
assert.NoError(t, err, "should not return error")

// Use require for fatal checks
require.NoError(t, err, "setup should not fail")
require.NotNil(t, obj, "object should not be nil")
```

## Shared Test Helpers

The test package provides these utilities:

### Context Management
```go
ctx := test.Context(t) // Returns context with auto-cancel
```

### Temporary Files/Dirs
```go
dir := test.TempDir(t)          // Auto-cleanup temp dir
file := test.TempFile(t, "data") // Auto-cleanup temp file
```

### File Assertions
```go
test.AssertFileExists(t, path)           // Check file exists
test.AssertFileNotExists(t, path)        // Check file missing
test.AssertFileContent(t, path, content) // Check file content
```

### Table-Driven Helper
```go
tests := []test.TableTest{
    {Name: "valid", Input: "test", Want: "result"},
    {Name: "invalid", Input: "", WantErr: true},
}
test.TableDriven(t, tests, func(t *testing.T, tt test.TableTest) {
    // Test logic here
})
```

## Testing Standards

### Test Coverage Requirements

| Test Type | Coverage Target | Race Detector |
|-----------|----------------|---------------|
| **Unit tests** | >80% per package | Required |
| **Integration tests** | >60% critical paths | Required |
| **End-to-end tests** | Cover main workflows | Optional (slow) |

### Running Tests

```bash
# Run all tests
make test

# Run tests without race detector (faster)
make test-fast

# Run tests with coverage report
make test-coverage

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run benchmarks
make benchmark
```

### Test Organization

```
internal/beads/
├── client.go
├── client_test.go          # Internal tests (package beads)
├── client_external_test.go  # External tests (package beads_test)
└── mocks/
    └── mock_client.go       # Generated mocks (if using gomock)
```

## Anti-Patterns to Avoid

### ❌ DON'T: Ignore errors in tests
```go
// BAD! Ignoring error
data, _ := readFile(path)
```

### ✅ DO: Handle errors properly
```go
// GOOD! Fatal assertion for setup errors
data, err := readFile(path)
require.NoError(t, err, "setup should not fail")
```

### ❌ DON'T: Use goroutines without synchronization
```go
// BAD! Race condition
go func() {
    result := doWork()
    results = append(results, result)
}()
```

### ✅ DO: Use t.Parallel() with proper synchronization
```go
// GOOD! Parallel test execution
for _, tt := range tests {
    tt := tt // Capture range variable
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        result := doWork(tt.input)
        assert.Equal(t, tt.expected, result)
    })
}
```

### ❌ DON'T: Test unexported functions from external tests
```go
// BAD! External test trying to access unexported
package mypackage_test

func TestHelper(t *testing.T) {
    helper() // Compile error! helper is unexported
}
```

### ✅ DO: Test exported behavior
```go
// GOOD! Test through public API
func TestPublicAPI(t *testing.T) {
    result := PublicFunction()
    assert.Equal(t, expected, result)
}
```

## Mocking Guidelines

### When to Mock
- External dependencies (databases, APIs, file system)
- Slow operations (network calls, disk I/O)
- Non-deterministic behavior (time, randomness)

### When NOT to Mock
- Simple functions with no external dependencies
- Value objects (structs with only data)
- Fast in-memory operations

### Mock Example (using testify/mock)

```go
// Mock interface
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Get(id string) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Usage in test
func TestService(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRepo.On("Get", "123").Return(&User{ID: "123"}, nil)

    svc := NewService(mockRepo)
    user, err := svc.GetUser("123")

    require.NoError(t, err)
    assert.Equal(t, "123", user.ID)
    mockRepo.AssertExpectations(t)
}
```

## Testing Workflow

### During Development

1. **Write test first** (TDD approach)
   ```bash
   # Create test file
   touch internal/mypackage/mypackage_test.go
   ```

2. **Run test to see it fail**
   ```bash
   go test -run TestMyFunction ./internal/mypackage/
   ```

3. **Implement minimum code to pass**
   ```go
   // Write just enough to make test pass
   ```

4. **Refactor while keeping tests green**
   ```bash
   # Continuously run tests while refactoring
   make test-fast
   ```

### Before Committing

```bash
# Run full test suite
make test

# Check coverage
make test-coverage

# Run pre-commit checks
make ci
```

## Integration Tests

Integration tests go in `test/integration/`:

```go
// test/integration/beads_test.go
package integration_test

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestBeadsIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Actual integration test with Beads CLI
    // ...
}
```

Run integration tests:
```bash
make test-integration
```

## References

- `.spec/research-findings/GO_CODING_STANDARDS.md` - Section 2: Testing Best Practices
- `test/test_helpers.go` - Shared test utilities
- `test/example_test.go` - Table-driven test examples
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Guidelines](https://go.dev/doc/tutorial/add-a-test)

---

**Version**: 1.0
**Last Updated**: 2025-02-10
