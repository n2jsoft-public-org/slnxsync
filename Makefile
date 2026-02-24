# Makefile for slnxsync

.PHONY: all build install test test-verbose test-coverage lint fmt vet clean help

# Variables
BINARY_NAME=slnxsync
BIN_DIR=./bin
MAIN_PATH=.
GO=go
GOFLAGS=
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS?=-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
all: lint test build

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make install        - Install the binary to GOPATH/bin"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make all            - Run lint, test, and build"
	@echo "  make ci             - Run CI checks (lint, test, build)"

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Build complete: $(BIN_DIR)/$(BINARY_NAME)"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)
	@echo "✓ Installation complete"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test ./...
	@echo "✓ Tests passed"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "✓ Coverage report generated: $(COVERAGE_FILE)"
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✓ HTML coverage report: $(COVERAGE_HTML)"

## lint: Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "Error: golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "✓ Lint checks passed"

## fmt: Format code with gofmt
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@echo "✓ Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "✓ Vet checks passed"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -f $(COVERAGE_HTML)
	@echo "✓ Clean complete"

## ci: Run CI checks (lint, test, build)
ci: lint test build
	@echo "✓ All CI checks passed"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "✓ Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	@echo "✓ Dependencies tidied"

# Run the CLI (example usage)
run:
	$(GO) run $(MAIN_PATH) $(ARGS)

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ Cross-platform builds complete"
