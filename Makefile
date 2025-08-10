# npm-console Makefile

# Variables
BINARY_NAME=npm-console
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build targets
.PHONY: all build clean test coverage lint fmt help install deps dev run

all: clean deps lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) .

# Build for multiple platforms
build-all: build-linux build-windows build-darwin

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 .

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64.exe .

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	$(GOLINT) run

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

# Development mode
dev: deps
	@echo "Starting development mode..."
	$(GOCMD) run . --verbose

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

# Run web server
web: build
	@echo "Starting web server..."
	./bin/$(BINARY_NAME) web

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all > docs/api.md

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Update dependencies
update:
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

# Vendor dependencies
vendor:
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

# Create release
release: clean deps lint test build-all
	@echo "Creating release..."
	mkdir -p dist
	cp bin/* dist/
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C bin $(BINARY_NAME)-linux-amd64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C bin $(BINARY_NAME)-linux-arm64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C bin $(BINARY_NAME)-darwin-amd64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C bin $(BINARY_NAME)-darwin-arm64
	zip -j dist/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip bin/$(BINARY_NAME)-windows-amd64.exe
	zip -j dist/$(BINARY_NAME)-$(VERSION)-windows-arm64.zip bin/$(BINARY_NAME)-windows-arm64.exe

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  build-all   - Build for all platforms"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  coverage    - Run tests with coverage"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  deps        - Install dependencies"
	@echo "  install     - Install the binary"
	@echo "  dev         - Start development mode"
	@echo "  run         - Run the application"
	@echo "  web         - Start web server"
	@echo "  docs        - Generate documentation"
	@echo "  security    - Run security scan"
	@echo "  bench       - Run benchmarks"
	@echo "  update      - Update dependencies"
	@echo "  vendor      - Vendor dependencies"
	@echo "  release     - Create release"
	@echo "  docker      - Build Docker image"
	@echo "  help        - Show this help"
