# git-mfpr Makefile

.PHONY: help build test lint clean install-tools release

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the git-mfpr binary
	go build -o bin/git-mfpr ./cmd/git-mfpr/

build-test: ## Build the test binary
	go build -o bin/test ./cmd/test/

build-all: ## Build all binaries
	mkdir -p bin
	go build -o bin/git-mfpr ./cmd/git-mfpr/
	go build -o bin/test ./cmd/test/

# Test targets
test: ## Run tests
	go test -v ./...

test-race: ## Run tests with race detection
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Linting and formatting
lint: ## Run golangci-lint
	golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix

fmt: ## Format code with gofmt
	go fmt ./...

fmt-check: ## Check if code is formatted
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		exit 1; \
	fi

# Development tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html

# Release targets
release-build: ## Build release binaries
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/git-mfpr-linux-amd64 ./cmd/git-mfpr/
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/git-mfpr-linux-arm64 ./cmd/git-mfpr/
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/git-mfpr-darwin-amd64 ./cmd/git-mfpr/
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/git-mfpr-darwin-arm64 ./cmd/git-mfpr/
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/git-mfpr-windows-amd64.exe ./cmd/git-mfpr/

# GoReleaser targets
goreleaser-snapshot: ## Build snapshot release locally
	goreleaser release --snapshot --rm-dist --skip-publish

goreleaser-build: ## Build release locally
	goreleaser build --snapshot --rm-dist

goreleaser-check: ## Check GoReleaser configuration
	goreleaser check

goreleaser-release: ## Create a new release (requires tag)
	goreleaser release --rm-dist

# Security
security: ## Run security scan
	gosec ./...

# Dependencies
deps: ## Download dependencies
	go mod download

deps-tidy: ## Tidy dependencies
	go mod tidy

# Development workflow
dev: deps fmt lint test ## Run full development workflow

# CI simulation
ci: deps fmt-check lint test-race test-coverage security ## Run CI checks locally