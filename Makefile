# git-mfpr Makefile

# Variables
BINARY_NAME := git-mfpr
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod

# Directories
CMD_DIR := ./cmd/git-mfpr
DIST_DIR := ./dist

.PHONY: all build test clean fmt vet tidy run help install

# Default target
all: test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: ./$(BINARY_NAME)"

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	
	@echo "All builds complete in $(DIST_DIR)/"

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(DIST_DIR)
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

# Update dependencies
tidy:
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy

# Run the binary with test arguments
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME) 123 --dry-run

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete"

# Help target
help:
	@echo "Available targets:"
	@echo "  make build      - Build the binary for current platform"
	@echo "  make build-all  - Build binaries for all supported platforms"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make fmt        - Format code"
	@echo "  make vet        - Run go vet"
	@echo "  make tidy       - Update dependencies"
	@echo "  make run        - Build and run with test arguments"
	@echo "  make install    - Install binary to GOPATH/bin"
	@echo "  make help       - Show this help message"