# Go Coding Standards for AGENTS.md

*Based on comprehensive research of Go best practices, official documentation, and community standards (2025)*

---

## Overview

Go follows a philosophy of **simplicity, readability, and productivity** over cleverness or performance at all costs. These standards align with Go's idioms while maintaining consistency with general engineering principles.

---

## 1. Linting & Code Style

### Essential Tools (Must Use)

```yaml
Required Tooling:
  - gofmt:          Non-negotiable canonical formatting
  - goimports:       Import management (superset of gofmt)
  - golangci-lint:   Meta-linter (50+ linters in one)
  - staticcheck:     Advanced static analysis
```

### Recommended golangci-lint Configuration

```yaml
linters:
  enable:
    - errcheck      # Check for unchecked errors
    - goimports     # Manage imports automatically
    - revive        # Style checker (modern golint replacement)
    - govet         # Official Go vet analyzer
    - staticcheck   # Bug detection and performance issues
    - ineffassign   # Detect ineffectual assignments
    - misspell      # Find commonly misspelled words
    - gocyclo       # Compute cyclometric complexities
    - gosec         # Security problems

linters-settings:
  govet:
    enable-all: true
  staticcheck:
    checks: ["all"]
```

### Code Style Rules

- **Line Length**: Soft limit of 99 characters (Uber standard)
- **Naming**:
  - Packages: `lowercase`, no underscores, no plurals (e.g., `net/http`, not `net/urls`)
  - Exported: `MixedCaps` (e.g., `MyFunction`)
  - Unexported: `MixedCaps` but internal (e.g., `myHelper`)
  - Constants: `MixedCaps` for exported, `camelCase` for unexported
  - Interfaces: `-er` suffix (e.g., `Reader`, `Writer`)
  - Errors: `Err` prefix for exported error variables (e.g., `ErrNotFound`)

---

## 2. Testing Best Practices

### Go Testing Philosophy

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
            name:     "IPv6 with port",
            input:    "[2001:db8::1]:80",
            wantHost: "2001:db8::1",
            wantPort: "80",
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

### Testing Standards

| Aspect | Standard |
|--------|----------|
| **Test File Location** | Same package as code (`xxx_test.go`) |
| **Test Package** | Internal tests: `package mypackage` <br> External tests: `package mypackage_test` |
| **Test Organization** | Use `t.Run()` for subtests (table-driven) |
| **Parallel Tests** | Use `t.Parallel()` when safe |
| **Assertion Library** | `stretchr/testify` (recommended) |
| **Mocking Framework** | `gomock` (interface-based) or testify/mock |
| **Test Coverage** | Run `go test -cover ./...` (aim for >80%) |
| **Race Detection** | Run `go test -race ./...` in CI |

### Preferred Testing Libraries

```go
// Testing Libraries
import (
    "testing"                    // Standard library
    "github.com/stretchr/testify"        // Assertions & mocks
    "github.com/stretchr/testify/assert"  // Assertions
    "github.com/stretchr/testify/require" // Fatal assertions (stop test)
    "github.com/stretchr/testify/mock"    // Mock objects
    "github.com/stretchr/testify/suite"   // Test suites
)

// Mocking (interface-based)
//go:generate mockgen -source=mypackage.go -destination=mocks/mypackage_mock.go
import "github.com/golang/mock/gomock"
```

### Test File Organization

```
myproject/
├── handler.go
├── handler_test.go          # Internal tests (package handler)
├── handler_external_test.go  # External tests (package handler_test)
└── mocks/
    └── mock_service.go       # Generated mocks (if using gomock)
```

---

## 3. Error Handling Standards

### Go's Error Philosophy

> "Errors are values." - Go Proverb
> "Don't just check errors, handle them gracefully." - Go Proverb

### Error Handling Patterns

```go
// ✅ GOOD: Early return with guard clause
func processFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }

    return data, nil
}

// ❌ BAD: Deep nesting
func processFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err == nil {
        defer f.Close()
        data, err := io.ReadAll(f)
        if err == nil {
            return data, nil
        } else {
            return nil, err
        }
    } else {
        return nil, err
    }
}
```

### Error Wrapping (Go 1.13+)

```go
// Use %w to wrap errors (preserves type for errors.Is/As)
if err != nil {
    return fmt.Errorf("context: %w", err)
}

// Match wrapped errors
if errors.Is(err, os.ErrNotExist) {
    // Handle not found
}

// Extract wrapped error type
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    // Handle specific error type
    fmt.Println(pathErr.Path)
}
```

### Error Naming

```go
// ✅ Exported error variables (for matching)
var (
    ErrNotFound   = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)

// ✅ Custom error types (for structured errors)
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s %q not found", e.Resource, e.ID)
}

// Usage
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

var notFound *NotFoundError
if errors.As(err, &notFound) {
    fmt.Printf("Resource: %s, ID: %s\n", notFound.Resource, notFound.ID)
}
```

---

## 4. Concurrency Patterns

### Go Concurrency Philosophy

> "Don't communicate by sharing memory; share memory by communicating." - Go Proverb
> "Concurrency is not parallelism." - Go Proverb
> "Channels orchestrate; mutexes serialize." - Go Proverb

### Channel vs Mutex

```go
// ✅ Use channels for orchestration and communication
func worker(jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        results <- process(job)
    }
}

// ✅ Use mutexes for serialization
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

### Context for Cancellation

```go
// ✅ ALWAYS pass context as first parameter
func fetchData(ctx context.Context, id string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

// ✅ Usage with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

data, err := fetchData(ctx, "user-123")
if err != nil {
    // Handle timeout or cancellation
    log.Printf("Request failed: %v", err)
}
```

### Goroutine Leaks Prevention

```go
// ✅ GOOD: Goroutine with controlled lifecycle
func startWorker() (stop func(), done <-chan struct{}) {
    stopCh := make(chan struct{})
    doneCh := make(chan struct{})

    go func() {
        defer close(doneCh)
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                doWork()
            case <-stopCh:
                return
            }
        }
    }()

    // Return function to stop goroutine
    stop := func() {
        close(stopCh)
    }

    return stop, doneCh
}

// ❌ BAD: Fire-and-forget goroutine
func badWorker() {
    go func() {
        for {
            doWork() // Never stops!
        }
    }()
}
```

---

## 5. Project Structure Standards

### Recommended Layout (Pragmatic Approach)

```
myproject/
├── cmd/                    # Main applications
│   ├── myapp/
│   │   └── main.go
│   └── myadmin/
│       └── main.go
├── internal/               # Private application code
│   ├── handler/
│   ├── service/
│   └── repository/
├── pkg/                    # Public library code (optional)
│   └── publiclib/
├── api/                    # API definitions (OpenAPI, protobuf)
├── configs/                # Configuration templates
├── scripts/                # Build and install scripts
├── test/                   # Additional test data and tools
├── docs/                   # Documentation
├── go.mod
├── go.sum
└── README.md
```

### Package Organization

| Principle | Guideline |
|-----------|-----------|
| **Package Names** | Single lowercase word, descriptive (`http`, `json`, `user`) |
| **Avoid** | `util`, `common`, `base`, `helper` |
| **Package Size** | Prefer fewer, larger packages over many small ones |
| **Internal Code** | Use `internal/` for code you don't want imported |
| **Package Visibility** | Unexported = private, Exported = public API |

### Module Structure

```go
// ✅ GOOD: Flat package structure
// github.com/user/myproject/internal/user
package user

type User struct {
    ID    string
    Name  string
    Email string
}

func New(id, name, email string) *User {
    return &User{
        ID:    id,
        Name:  name,
        Email: email,
    }
}
```

---

## 6. Go-Specific Standards (vs. General Coding Standards)

### Similarities with General Standards

| Aspect | Go Practice | General Standard |
|--------|------------|------------------|
| **Readable Code** | ✅ High priority | ✅ High priority |
| **Small Functions** | ✅ Preferred | ✅ Preferred |
| **DRY Principle** | ✅ Applicable | ✅ Applicable |
| **Code Reviews** | ✅ Essential | ✅ Essential |
| **Testing** | ✅ Table-driven | ✅ Unit tests |
| **Documentation** | ✅ Godoc comments | ✅ API docs |

### Unique Go Standards

| Aspect | Go Standard | Why Different |
|--------|------------|---------------|
| **Error Handling** | Explicit, no exceptions | Errors are values, not control flow |
| **Concurrency** | Goroutines + channels | Built-in primitives vs. threads |
| **Interfaces** | Implicit, structural typing | Duck typing vs. explicit implementation |
| **Nullability** | Nil is valid for slices/maps | No null vs. nil distinction |
| **Initialization** | Zero values are useful | No constructors needed |
| **Encapsulation** | Exported = Capitalized | No public/private keywords |
| **Inheritance** | Composition via embedding | No extends keyword |
| **Generics** | Go 1.18+ (type parameters) | Minimal, focused usage |

---

## 7. Common Patterns & Idioms

### Interface Design

```go
// ✅ GOOD: Small, focused interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// ✅ GOOD: Interfaces accept interfaces
func StoreData(w io.Writer, data []byte) error {
    // Can accept any Writer (files, buffers, networks)
}

// ❌ BAD: Large interfaces ("interface bloat")
type Database interface {
    Query(sql string) Rows
    Exec(sql string) Result
    Begin() Tx
    Close() error
    Ping() error
    Stats() Stats
    // Too many methods!
}
```

### Functional Options Pattern

```go
// ✅ GOOD: Functional options for configurable APIs
type Server struct {
    addr    string
    timeout time.Duration
    logger  *zap.Logger
}

type Option interface {
    apply(*Server)
}

func WithTimeout(timeout time.Duration) Option {
    return timeoutOption(timeout)
}

func WithLogger(logger *zap.Logger) Option {
    return loggerOption(logger)
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second,
        logger:  zap.NewNop(),
    }
    for _, opt := range opts {
        opt.apply(s)
    }
    return s
}

// Usage
srv := NewServer(":8080",
    WithTimeout(10*time.Second),
    WithLogger(logger),
)
```

### Context Propagation Pattern

```go
// ✅ GOOD: Context flows through the call stack
func HandleRequest(ctx context.Context, req Request) error {
    // Add values to context
    ctx = context.WithValue(ctx, "requestID", generateID())

    // Pass context down
    if err := processRequest(ctx, req); err != nil {
        return fmt.Errorf("process request: %w", err)
    }

    if err := persistResult(ctx, req.Result); err != nil {
        return fmt.Errorf("persist result: %w", err)
    }

    return nil
}
```

---

## 8. Performance Guidelines

### Pre-allocating Capacity

```go
// ✅ GOOD: Pre-allocate when size is known
func buildResults(items []Item) []Result {
    results := make([]Result, 0, len(items)) // Capacity hint

    for _, item := range items {
        results = append(results, process(item))
    }

    return results
}

// ✅ GOOD: Map with capacity hint
m := make(map[string]int, 100) // Avoids resizing
```

### String Conversions

```go
// ❌ BAD: Repeated conversion in loop
for i := 0; i < 1000; i++ {
    s := fmt.Sprint(i) // Slow!
    writeString(s)
}

// ✅ GOOD: Use strconv for primitives
for i := 0; i < 1000; i++ {
    s := strconv.Itoa(i) // Fast!
    writeString(s)
}
```

### Avoid allocations in hot paths

```go
// ❌ BAD: Allocates on each iteration
for _, item := range items {
    data := []byte(item.Name)
    process(data)
}

// ✅ GOOD: Reuse buffer
var buf bytes.Buffer
for _, item := range items {
    buf.Reset()
    buf.WriteString(item.Name)
    process(buf.Bytes())
}
```

---

## 9. Documentation Standards

### Godoc Comments

```go
// Package user provides user management functionality.
//
// This package handles user registration, authentication,
// and profile management following our security guidelines.
package user

// UserService handles user business logic.
//
// UserService provides methods for creating, updating, and
// retrieving user information while enforcing business rules.
type UserService struct {
    repo Repository
}

// NewUserService creates a new UserService with the given repository.
//
// It validates that the repository is not nil before returning.
// Returns an error if validation fails.
func NewUserService(repo Repository) (*UserService, error) {
    if repo == nil {
        return nil, errors.New("repository cannot be nil")
    }
    return &UserService{repo: repo}, nil
}

// CreateUser creates a new user with the given details.
//
// It validates the email format and ensures the email is not already
// registered. Returns the created user ID or an error.
//
// Example:
//
//	id, err := svc.CreateUser(ctx, CreateUserRequest{
//	    Name:  "John Doe",
//	    Email: "john@example.com",
//	})
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (string, error) {
    // Implementation...
}
```

### Comment Rules

| Type | Guideline |
|------|-----------|
| **Package comments** | Block comment before package clause |
| **Exported functions** | Full sentence, mentions what (not how) |
| **Exported types** | Description of the type and usage |
| **Exported constants** | Meaning of the value |
| **Unexported code** | Comments explain non-obvious logic |
| **Interface implementations** | No comment needed (obvious from interface) |

---

## 10. CI/CD Integration

### Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Format code
go fmt ./...
goimports -w .

# Run linters
golangci-lint run

# Run tests
go test -race -cover ./...

# Check for go.sum inconsistencies
go mod tidy
go mod verify
```

### GitHub Actions Example

```yaml
name: Go

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Install dependencies
      run: go mod download

    - name: Run go vet
      run: go vet ./...

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v6
      with:
        version: latest

    - name: Run tests
      run: go test -race -covermode=atomic -coverprofile=coverage.out ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.out
```

---

## 11. Quick Reference Checklist

### Code Review Checklist

- [ ] Code is formatted with `gofmt` / `goimports`
- [ ] `golangci-lint` passes with no issues
- [ ] `go vet` passes with no warnings
- [ ] Tests pass with `go test ./...`
- [ ] Tests pass with race detector: `go test -race ./...`
- [ ] Test coverage is acceptable (>80%)
- [ ] Public symbols have godoc comments
- [ ] Errors are wrapped with context using `%w`
- [ ] No `panic()` in production code (return errors instead)
- [ ] Goroutines have controlled lifecycles
- [ ] Context is passed as first parameter
- [ ] Zero values are useful where appropriate
- [ ] No package-level mutable state

### Common Anti-Patterns to Avoid

```go
// ❌ DON'T: Panic in production code
func process(data string) {
    if data == "" {
        panic("data is empty") // BAD!
    }
}

// ✅ DO: Return errors
func process(data string) error {
    if data == "" {
        return errors.New("data is empty")
    }
    return nil
}

// ❌ DON'T: Ignore errors
data, _ := ioutil.ReadFile(path) // BAD!

// ✅ DO: Always handle errors
data, err := ioutil.ReadFile(path)
if err != nil {
    return fmt.Errorf("read file: %w", err)
}

// ❌ DON'T: Use goroutines without lifecycle control
go func() {
    for { doWork() } // Never stops!
}()

// ✅ DO: Control goroutine lifecycles
ctx, cancel := context.WithCancel(context.Background())
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
        }
    }
}()
// Remember to call cancel() when done!
```

---

## 12. Migration Guide from Other Languages

### For JavaScript/TypeScript Developers

| JavaScript | Go | Notes |
|------------|-----|-------|
| `try/catch` | Explicit error checks | Go doesn't use exceptions |
| `async/await` | Goroutines + channels | Concurrency is built-in |
| `class extends` | Struct embedding | Composition over inheritance |
| `undefined` | `nil` | Zero values are useful |
| `import/export` | Uppercase for exported | Capitalization = visibility |
| `===` | `==` (no coercion in Go) | Strong typing |
| `map/filter` | `for` loops with `range` | Iteration is explicit |

### For Python Developers

| Python | Go | Notes |
|--------|-----|-------|
| Indentation | Braces | Go uses `{}` |
| Dynamic typing | Static typing | Type safety at compile time |
| List comprehensions | `for` loops | Explicit is better than clever |
| `__init__` | No constructors | Zero values are useful |
| `with` statement | `defer` | Deferred cleanup |
| Exceptions | Error returns | No exception stack unwinding |

### For Java Developers

| Java | Go | Notes |
|-----|-----|-------|
| `public/private` | Capitalization | Exported = Capitalized |
| Exceptions | Error values | Explicit error handling |
| `extends` | Embedding | Composition, not inheritance |
| `@Override` | No syntax needed | Implicit interface satisfaction |
| `null` | `nil` | Zero values for everything |
| Generics | Type parameters (Go 1.18+) | Simpler, fewer features |
| Constructors | `New` functions | No constructors, just functions |

---

## References & Further Reading

### Official Documentation
1. [Effective Go](https://go.dev/doc/effective_go) - Must-read for all Go developers
2. [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - Common review feedback
3. [Go Modules Reference](https://go.dev/ref/mod) - Module management
4. [Go Blog](https://go.dev/blog/) - Official blog with best practices

### Style Guides
1. [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) - Comprehensive production standards
2. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) - Official review guidelines
3. [Go Proverbs](https://go-proverbs.github.io/) - Simple, poetic, pithy guidelines

### Testing & Tools
1. [Testify Documentation](https://github.com/stretchr/testify) - Assertion library
2. [GoMock User Guide](https://github.com/golang/mock) - Mocking framework
3. [golangci-lint](https://golangci-lint.run/) - Meta-linter documentation
4. [Staticcheck](https://staticcheck.dev/docs/) - Advanced static analysis

### Books
1. "The Go Programming Language" by Donovan & Kernighan
2. "Go in Action" by Manning
3. "100 Go Mistakes and How to Avoid Them" - Compact reference

### Videos & Talks
1. [GopherCon Videos](https://www.youtube.com/@GopherCon)
2. Rob Pike: "Go Proverbs" - Essential philosophy
3. Dave Cheney: "Practical Go" - Real-world advice
4. Brad Fitzpatrick: "Let's Talk About Go" - Design decisions

---

## Appendix: Sample .golangci.yml

```yaml
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - bodyclose       # Check whether HTTP response body is closed
    - deadcode        # Find unused code
    - depguard        # Check for import restrictions
    - errcheck        # Check for unchecked errors
    - gofmt           # Check that code is gofmt-ed
    - goimports       # Check import order and missing imports
    - gosec           # Security problems
    - gosimple        # Simplify code
    - govet           # Reports suspicious constructs
    - ineffassign     # Detect ineffectual assignments
    - misspell        # Find commonly misspelled words
    - revive          # Fast, configurable linter
    - staticcheck     # Go 1.13+ linter (bugs, performance)
    - typecheck       # Parse and type-check code
    - unconvert       # Remove unnecessary type conversions
    - unused          # Check for unused constants, variables, functions

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true

  revive:
    rules:
      - name: exported-functions-have-comments
        severity: warning
      - name: var-naming
        severity: warning
      - name: exported-var-have-comments
        severity: warning

  staticcheck:
    checks: ["all"]

  govet:
    enable-all: true
    disable:
      - shadow

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

---

*Document Version: 1.0*
*Last Updated: 2025-02-10*
*Standards aligned with Go 1.21+ and Go 1.22+*
