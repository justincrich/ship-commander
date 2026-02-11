# Ship Commander 3 - Development Guide

## Quick Start

### Prerequisites

```bash
# Install Go 1.22+
go version

# Install development tools
make tools-install

# Install pre-commit hooks (optional but recommended)
make hooks-install
```

### Development Workflow

```bash
# 1. Start development server with hot reload
make dev

# 2. Make changes - auto-reload on save

# 3. Run tests
make test

# 4. Run linting
make lint

# 5. Format code
make fmt

# 6. Commit (pre-commit hooks will run checks)
git commit -m "Your commit message"
```

## Common Commands

### Building

| Command | Purpose |
|---------|---------|
| `make build` | Build binary with race detector |
| `make build-fast` | Build without race detector (faster) |
| `make run` | Build and run immediately |
| `make install` | Install to `$GOPATH/bin` |

### Testing

| Command | Purpose |
|---------|---------|
| `make test` | Run all tests with race detector |
| `make test-fast` | Run tests without race detector |
| `make test-coverage` | Run tests with coverage report |
| `make test-coverage-html` | Generate HTML coverage report |
| `make test-unit` | Run unit tests only (no integration) |
| `make test-integration` | Run integration tests only |
| `make benchmark` | Run benchmarks |

### Linting & Formatting

| Command | Purpose |
|---------|---------|
| `make lint` | Run golangci-lint (full) |
| `make lint-fast` | Run fast linters only |
| `make fmt` | Format code with goimports |
| `make fmt-check` | Check if code is formatted |
| `make vet` | Run go vet |
| `make check` | Run all checks (fmt, vet, lint) |

### CI/CD

| Command | Purpose |
|---------|---------|
| `make ci` | Run full CI pipeline |
| `make ci-fast` | Run fast CI (for development) |

### Dependencies

| Command | Purpose |
|---------|---------|
| `make deps` | Download dependencies |
| `make deps-tidy` | Tidy go.mod and go.sum |
| `make deps-verify` | Verify dependencies |
| `make deps-update` | Update all dependencies |

## Hot Reload Development

Ship Commander 3 uses [`air`](https://github.com/cosmtrek/air) for hot reload during development.

### How It Works

1. Monitor Go, TOML, and YAML files for changes
2. Automatically rebuild on file save
3. Restart the binary
4. Minimal downtime (~1-2 seconds)

### Configuration

Hot reload is configured in `.air.toml`:
- **Include files**: `.go`, `.toml`, `.yaml`, `.yml`
- **Exclude directories**: `build/`, `test/`, `.spec/`, `design/`, `docs/`
- **Build delay**: 1 second (debounce rapid changes)
- **Graceful shutdown**: Sends interrupt signal before kill

### Starting Dev Server

```bash
# Method 1: Using make (recommended)
make dev

# Method 2: Using air directly
air

# Method 3: Custom binary name
BINARY_NAME=my-app air
```

### Dev Server Commands

While the dev server is running:
- **Ctrl+C**: Stop the server
- File changes trigger automatic rebuild
- Check `build/tmp/` for build artifacts

## Testing Best Practices

### Table-Driven Tests

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {name: "valid", input: "test", expected: "result", wantErr: false},
        {name: "invalid", input: "", expected: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Test Organization

```
internal/mypackage/
├── mypackage.go          # Implementation
├── mypackage_test.go     # Internal tests (package mypackage)
└── mypackage_external_test.go  # External tests (package mypackage_test)
```

### Running Specific Tests

```bash
# Run specific test
go test -run TestMyFunction ./internal/mypackage/

# Run specific test with verbose output
go test -v -run TestMyFunction ./internal/mypackage/

# Run tests matching a pattern
go test -run TestMultiply ./...

# Skip integration tests
go test -short ./...
```

## Linting

### golangci-lint Configuration

The project uses `.golangci.yml` with these linters enabled:

**Core Linters** (non-negotiable):
- `errcheck` - Check for unchecked errors
- `gofmt` - Code formatting
- `goimports` - Import management
- `govet` - Official Go vet analyzer
- `staticcheck` - Bug detection
- `revive` - Style checker

**Additional Linters**:
- `ineffassign` - Ineffectual assignments
- `misspell` - Spelling mistakes
- `gocyclo` - Cyclomatic complexity
- `gosec` - Security issues
- `bodyclose` - HTTP response body closed
- `deadcode` - Unused code
- And more...

### Fixing Linter Issues

```bash
# Auto-fix formatting issues
make fmt

# See all linter issues
golangci-lint run

# Fix specific issues (if auto-fix available)
golangci-lint run --fix
```

## Code Style

### Formatting

```bash
# Format all code
goimports -w .
gofmt -s -w .

# Or use make
make fmt
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| **Packages** | `lowercase`, no underscores | `net/http`, `user` |
| **Exported** | `MixedCaps` | `MyFunction`, `UserType` |
| **Unexported** | `mixedCaps` | `myHelper`, `internalState` |
| **Constants (exported)** | `MixedCaps` | `MaxRetries`, `DefaultPort` |
| **Interfaces** | `-er` suffix | `Reader`, `Writer`, `Closer` |
| **Errors** | `Err` prefix | `ErrNotFound`, `ErrInvalidInput` |

### Error Handling

```go
// ✅ GOOD: Early return with error wrapping
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
        }
    }
    return nil, err
}
```

## Pre-commit Hooks

The project includes pre-commit hooks (`.git/hooks/pre-commit`) that run:

1. **Format check** - Ensures code is formatted
2. **go vet** - Runs Go vet analyzer
3. **golangci-lint** - Runs linters on changed files only
4. **TODO/FIXME check** - Warns about unfinished work
5. **Unit tests** - Runs tests for changed packages

### Bypassing Hooks (Not Recommended)

```bash
# Skip hooks (use sparingly!)
git commit --no-verify -m "WIP: work in progress"
```

## Development Tips

### 1. Use `make help` for Commands

```bash
make help
# Displays all available commands with descriptions
```

### 2. Run Tests in Parallel

```bash
# Run all tests in parallel (faster on multi-core)
go test -parallel 4 ./...
```

### 3. Test Coverage by Package

```bash
# Check coverage for specific package
go test -cover ./internal/beads/

# Generate coverage profile
go test -coverprofile=coverage.out ./internal/beads/
go tool cover -func=coverage.out
```

### 4. Race Detector

Always run tests with race detector before committing:
```bash
make test  # Includes -race flag
```

### 5. Benchmarking

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkMultiply -benchmem ./...
```

### 6. Memory Profiling

```bash
# Profile memory usage
go test -memprofile=mem.prof ./...

# Analyze profile
go tool pprof mem.prof
```

## Common Issues

### Issue: "go: cannot find main module"

**Solution**: Run from project root directory
```bash
cd /path/to/ship-commander-3
make build
```

### Issue: "golangci-lint not found"

**Solution**: Install development tools
```bash
make tools-install
```

### Issue: Pre-commit hooks failing

**Solution**: Run checks manually first
```bash
make ci
```

Then fix issues and commit again.

### Issue: Hot reload not working

**Solution**: Check file watcher limits
```bash
# macOS
brew install watchman

# Linux
echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

## Best Practices Summary

1. **Write tests first** (TDD approach)
2. **Run tests frequently** (save → test cycle)
3. **Format code before committing** (auto-formatted)
4. **Never ignore linter warnings** (fix or document why)
5. **Use table-driven tests** (idiomatic Go)
6. **Handle errors explicitly** (early returns)
7. **Pass context as first parameter** (for cancellable operations)
8. **Control goroutine lifecycles** (use context cancellation)
9. **Document exported symbols** (Godoc comments)
10. **Keep packages focused** (single responsibility)

## References

- [Go Coding Standards](.spec/research-findings/GO_CODING_STANDARDS.md) - Full Go standards guide
- [AGENTS.md](AGENTS.md) - Root-level agent coordination
- [Test Package Guide](test/AGENTS.md) - Testing patterns and utilities
- [Makefile](Makefile) - All available commands

---

**Version**: 1.0
**Last Updated**: 2025-02-10
