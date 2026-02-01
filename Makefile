# Claude Swarm Makefile

BINARY_NAME=swarm
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build install uninstall clean test run help

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/swarm
	@echo "Done: ./$(BINARY_NAME)"

# Install globally (to GOPATH/bin or /usr/local/bin)
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ -n "$(GOPATH)" ]; then \
		cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME); \
		echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	elif [ -d "$(HOME)/go/bin" ]; then \
		cp $(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME); \
		echo "Installed to $(HOME)/go/bin/$(BINARY_NAME)"; \
	else \
		sudo cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	fi
	@echo ""
	@echo "Verify installation:"
	@echo "  swarm --help"

# Install using go install (recommended)
install-go:
	@echo "Installing via go install..."
	@go install $(LDFLAGS) ./cmd/swarm
	@echo "Done. Make sure $(GOPATH)/bin or $(HOME)/go/bin is in your PATH"

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || true
	@rm -f $(HOME)/go/bin/$(BINARY_NAME) 2>/dev/null || true
	@sudo rm -f /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@echo "Uninstalled"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf .worktrees
	@go clean
	@echo "Done"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Quick run (for development)
run: build
	@./$(BINARY_NAME) $(ARGS)

# Format code
fmt:
	@go fmt ./...

# Lint code
lint:
	@golangci-lint run ./...

# Show help
help:
	@echo "Claude Swarm - Multi-Agent Development System"
	@echo ""
	@echo "Targets:"
	@echo "  make build          Build the binary"
	@echo "  make install        Build and install globally"
	@echo "  make install-go     Install via 'go install'"
	@echo "  make uninstall      Remove installed binary"
	@echo "  make clean          Clean build artifacts"
	@echo "  make test           Run tests"
	@echo "  make test-coverage  Run tests with coverage"
	@echo "  make fmt            Format code"
	@echo "  make lint           Lint code"
	@echo ""
	@echo "Quick Start:"
	@echo "  make install"
	@echo "  cd your-project"
	@echo "  swarm init"
	@echo "  swarm run \"Create a README\""
