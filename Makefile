# watch-now Makefile

# Variables
BINARY_NAME := watch-now
VERSION := 0.1.0
BUILD_DIR := build
COVERAGE_DIR := coverage
GO_FILES := $(shell find . -name '*.go' -type f -not -path './vendor/*')

# Build variables
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_RED := \033[31m

.PHONY: all build run clean test coverage lint fmt complexity deadcode ci install help

# Default target
all: ci build

# Build the binary
build:
	@echo "$(COLOR_BOLD)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

# Run the application
run: build
	@echo "$(COLOR_BOLD)Running $(BINARY_NAME)...$(COLOR_RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Run once
run-once: build
	@./$(BUILD_DIR)/$(BINARY_NAME) --once

# Clean build artifacts
clean:
	@echo "$(COLOR_BOLD)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	@go clean -testcache
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

# Run tests
test:
	@echo "$(COLOR_BOLD)Running tests...$(COLOR_RESET)"
	@go test -v ./...
	@echo "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)"

# Run tests with coverage
coverage:
	@echo "$(COLOR_BOLD)Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: $(COVERAGE_DIR)/coverage.html$(COLOR_RESET)"

# Format code
fmt:
	@echo "$(COLOR_BOLD)Formatting code...$(COLOR_RESET)"
	@gofmt -w $(GO_FILES)
	@echo "$(COLOR_GREEN)✓ Format complete$(COLOR_RESET)"

# Lint code
lint:
	@echo "$(COLOR_BOLD)Linting code...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Lint complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Install with:$(COLOR_RESET)"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		exit 1; \
	fi

# Check code complexity
complexity:
	@echo "$(COLOR_BOLD)Checking code complexity...$(COLOR_RESET)"
	@if command -v gocyclo >/dev/null 2>&1; then \
		gocyclo -over 10 $(GO_FILES) | tee /tmp/gocyclo.out; \
		if [ -s /tmp/gocyclo.out ]; then \
			echo "$(COLOR_RED)✗ Complexity check failed: functions with complexity > 10 found$(COLOR_RESET)"; \
			rm /tmp/gocyclo.out; \
			exit 1; \
		else \
			echo "$(COLOR_GREEN)✓ Complexity check passed$(COLOR_RESET)"; \
			rm -f /tmp/gocyclo.out; \
		fi \
	else \
		echo "$(COLOR_YELLOW)⚠ gocyclo not installed. Install with:$(COLOR_RESET)"; \
		echo "  go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"; \
		exit 1; \
	fi

# Check for dead code
deadcode:
	@echo "$(COLOR_BOLD)Checking for dead code...$(COLOR_RESET)"
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck -checks="U*" ./... | tee /tmp/deadcode.out; \
		if [ -s /tmp/deadcode.out ]; then \
			echo "$(COLOR_RED)✗ Dead code check failed: unused code found$(COLOR_RESET)"; \
			rm /tmp/deadcode.out; \
			exit 1; \
		else \
			echo "$(COLOR_GREEN)✓ Dead code check passed$(COLOR_RESET)"; \
			rm -f /tmp/deadcode.out; \
		fi \
	else \
		echo "$(COLOR_YELLOW)⚠ staticcheck not installed. Install with:$(COLOR_RESET)"; \
		echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"; \
		exit 1; \
	fi

# Run all CI checks
ci: fmt lint complexity deadcode test build
	@echo "$(COLOR_BOLD)========================================$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ All CI checks passed!$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)========================================$(COLOR_RESET)"

# Install development dependencies
install-deps:
	@echo "$(COLOR_BOLD)Installing development dependencies...$(COLOR_RESET)"
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "$(COLOR_GREEN)✓ Dependencies installed$(COLOR_RESET)"

# Install binary to system
install: build
	@echo "$(COLOR_BOLD)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	@install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $$(go env GOPATH)/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ $(BINARY_NAME) installed to $$(go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Make sure $$(go env GOPATH)/bin is in your PATH$(COLOR_RESET)"

# Initialize go module if needed
init:
	@if [ ! -f go.mod ]; then \
		echo "$(COLOR_BOLD)Initializing Go module...$(COLOR_RESET)"; \
		go mod init github.com/orchard9/watch-now; \
		go mod tidy; \
		echo "$(COLOR_GREEN)✓ Go module initialized$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)Go module already initialized$(COLOR_RESET)"; \
	fi

# Help target
help:
	@echo "$(COLOR_BOLD)watch-now - Universal Development Monitor$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@echo "  make [target]"
	@echo ""
	@echo "$(COLOR_BOLD)Targets:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)build$(COLOR_RESET)        Build the watch-now binary"
	@echo "  $(COLOR_GREEN)run$(COLOR_RESET)          Build and run watch-now"
	@echo "  $(COLOR_GREEN)run-once$(COLOR_RESET)     Build and run watch-now once"
	@echo "  $(COLOR_GREEN)clean$(COLOR_RESET)        Remove build artifacts"
	@echo "  $(COLOR_GREEN)test$(COLOR_RESET)         Run tests"
	@echo "  $(COLOR_GREEN)coverage$(COLOR_RESET)     Run tests with coverage report"
	@echo "  $(COLOR_GREEN)fmt$(COLOR_RESET)          Format code using gofmt"
	@echo "  $(COLOR_GREEN)lint$(COLOR_RESET)         Lint code using golangci-lint"
	@echo "  $(COLOR_GREEN)complexity$(COLOR_RESET)   Check code complexity (max: 10)"
	@echo "  $(COLOR_GREEN)deadcode$(COLOR_RESET)     Check for dead code"
	@echo "  $(COLOR_GREEN)ci$(COLOR_RESET)           Run all CI checks"
	@echo "  $(COLOR_GREEN)install$(COLOR_RESET)      Install binary to \$$GOPATH/bin"
	@echo "  $(COLOR_GREEN)install-deps$(COLOR_RESET) Install development dependencies"
	@echo "  $(COLOR_GREEN)init$(COLOR_RESET)         Initialize Go module"
	@echo "  $(COLOR_GREEN)help$(COLOR_RESET)         Show this help message"
	@echo ""
	@echo "$(COLOR_BOLD)CI Pipeline:$(COLOR_RESET)"
	@echo "  The 'ci' target runs: fmt → lint → complexity → deadcode → test → build"
	@echo ""
	@echo "$(COLOR_BOLD)Examples:$(COLOR_RESET)"
	@echo "  make build        # Build the binary"
	@echo "  make run          # Run continuous monitoring"
	@echo "  make ci           # Run all checks before committing"