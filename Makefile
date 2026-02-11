# Ship Commander 3 - Makefile
# Provides convenient targets for building, testing, linting, and development

# Variables
BINARY_NAME=sc3
MAIN_PATH=./cmd/root
BUILD_DIR=./build
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
GO_BIN_PATH=$(shell if [ -n "$$(go env GOBIN)" ]; then printf "%s" "$$(go env GOBIN)"; else printf "%s/bin" "$$(go env GOPATH)"; fi)
GOLANGCI_LINT=$(GO_BIN_PATH)/golangci-lint

export PATH := $(PATH):$(GO_BIN_PATH)

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"
GOFLAGS=-race

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: help
help: ## Display this help message
	@echo "$(COLOR_BOLD)Ship Commander 3 - Development Commands$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

## Development (Hot Reload)

.PHONY: dev
dev: ## Run with hot reload during development (requires air)
	@echo "$(COLOR_GREEN)🚀 Starting development server with hot reload...$(COLOR_RESET)"
	@air

.PHONY: dev-install
dev-install: ## Install air for hot reload
	@echo "$(COLOR_YELLOW)📦 Installing air for hot reload...$(COLOR_RESET)"
	@go install github.com/cosmtrek/air@latest
	@echo "$(COLOR_GREEN)✅ air installed to $(shell go env GOPATH)/bin/air$(COLOR_RESET)"

## Building

.PHONY: build
build: clean ## Build the binary
	@echo "$(COLOR_GREEN)🔨 Building $(BINARY_NAME)...$(COLOR_RESET)"
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(COLOR_GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-fast
build-fast: ## Build without race detector (faster for development)
	@echo "$(COLOR_GREEN)🔨 Building $(BINARY_NAME) (fast)...$(COLOR_RESET)"
	@go build -ldflags "-X main.Version=dev" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(COLOR_GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_YELLOW)🧹 Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✅ Clean complete$(COLOR_RESET)"

## Testing

.PHONY: test
test: ## Run all tests
	@echo "$(COLOR_GREEN)🧪 Running tests...$(COLOR_RESET)"
	@go test -v -race -cover ./...

.PHONY: test-fast
test-fast: ## Run tests without race detector (faster)
	@echo "$(COLOR_GREEN)🧪 Running tests (fast)...$(COLOR_RESET)"
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_GREEN)🧪 Running tests with coverage...$(COLOR_RESET)"
	@go test -race -covermode=atomic -coverprofile=$(COVERAGE_FILE) ./...
	@echo ""
	@echo "$(COLOR_GREEN)📊 Overall coverage:$(COLOR_RESET)"
	@go tool cover -func=$(COVERAGE_FILE) | grep total

.PHONY: test-coverage-html
test-coverage-html: ## Generate HTML coverage report
	@echo "$(COLOR_GREEN)🧪 Running tests with coverage (HTML)...$(COLOR_RESET)"
	@go test -race -covermode=atomic -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✅ Coverage report: $(COVERAGE_HTML)$(COLOR_RESET)"

.PHONY: test-unit
test-unit: ## Run unit tests only (no integration)
	@echo "$(COLOR_GREEN)🧪 Running unit tests...$(COLOR_RESET)"
	@go test -v -race -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "$(COLOR_GREEN)🧪 Running integration tests...$(COLOR_RESET)"
	@go test -v -race ./test/integration/...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(COLOR_GREEN)⚡ Running benchmarks...$(COLOR_RESET)"
	@go test -bench=. -benchmem ./...

## Linting & Formatting

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(COLOR_GREEN)🔍 Running linters...$(COLOR_RESET)"
	@$(GOLANGCI_LINT) run --timeout=5m

.PHONY: lint-fast
lint-fast: ## Run linters with fast checks only
	@echo "$(COLOR_GREEN)🔍 Running linters (fast)...$(COLOR_RESET)"
	@$(GOLANGCI_LINT) run --fast --timeout=2m

.PHONY: fmt
fmt: ## Format code with goimports and gofmt
	@echo "$(COLOR_GREEN)✨ Formatting code...$(COLOR_RESET)"
	@goimports -w .
	@gofmt -s -w .
	@echo "$(COLOR_GREEN)✅ Formatting complete$(COLOR_RESET)"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(COLOR_GREEN)✨ Checking formatting...$(COLOR_RESET)"
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@test -z "$$(goimports -l . | tee /dev/stderr)" || (echo "Imports are not formatted. Run 'make fmt'" && exit 1)
	@echo "$(COLOR_GREEN)✅ Code is properly formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_GREEN)🔍 Running go vet...$(COLOR_RESET)"
	@go vet ./...

.PHONY: check
check: fmt-check vet lint ## Run all checks (format, vet, lint)
	@echo "$(COLOR_GREEN)✅ All checks passed$(COLOR_RESET)"

## Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_GREEN)📦 Downloading dependencies...$(COLOR_RESET)"
	@go mod download

.PHONY: deps-tidy
deps-tidy: ## Tidy go.mod and go.sum
	@echo "$(COLOR_GREEN)📦 Tidying dependencies...$(COLOR_RESET)"
	@go mod tidy
	@echo "$(COLOR_GREEN)✅ Dependencies tidy$(COLOR_RESET)"

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@echo "$(COLOR_GREEN)🔒 Verifying dependencies...$(COLOR_RESET)"
	@go mod verify

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "$(COLOR_GREEN)📦 Updating dependencies...$(COLOR_RESET)"
	@go get -u ./...
	@go mod tidy

## Tools Installation

.PHONY: tools-install
tools-install: ## Install development tools
	@echo "$(COLOR_GREEN)🔧 Installing development tools...$(COLOR_RESET)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(COLOR_GREEN)✅ Tools installed$(COLOR_RESET)"

## Git Hooks

.PHONY: hooks-install
hooks-install: ## Install pre-commit hooks
	@echo "$(COLOR_GREEN)🪝 Installing pre-commit hooks...$(COLOR_RESET)"
	@chmod +x scripts/pre-commit
	@ln -sf ../../scripts/pre-commit .git/hooks/pre-commit
	@echo "$(COLOR_GREEN)✅ Pre-commit hooks installed$(COLOR_RESET)"

## CI/CD (for automation)

.PHONY: ci
ci: deps-tidy fmt-check vet lint test ## Run full CI pipeline
	@echo "$(COLOR_GREEN)✅ CI pipeline passed$(COLOR_RESET)"

.PHONY: ci-fast
ci-fast: fmt-check vet lint-fast test-fast ## Run fast CI pipeline (for dev)

## Utilities

.PHONY: run
run: build-fast ## Build and run the binary
	@echo "$(COLOR_GREEN)▶️  Running $(BINARY_NAME)...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

.PHONY: install
install: ## Install binary to $GOPATH/bin
	@echo "$(COLOR_GREEN)📦 Installing $(BINARY_NAME)...$(COLOR_RESET)"
	@go install $(LDFLAGS) $(MAIN_PATH)
	@echo "$(COLOR_GREEN)✅ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: version
version: ## Print version information
	@echo "$(COLOR_BOLD)Ship Commander 3$(COLOR_RESET)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

.PHONY: clean-all
clean-all: clean ## Clean everything including test cache
	@echo "$(COLOR_YELLOW)🧹 Cleaning all artifacts...$(COLOR_RESET)"
	@go clean -cache -testcache
	@echo "$(COLOR_GREEN)✅ Complete clean finished$(COLOR_RESET)"

# Default variables
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
ARGS?=
