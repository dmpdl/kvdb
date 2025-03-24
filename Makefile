# Go command to use
GO := go

# Default target
all: test lint run

# Run all tests
test:
	@echo "Running all tests..."
	$(GO) test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GO) test ./... -race

# Run linter (golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Build and run the application
run:
	@echo "Building and running the application..."
	$(GO) run cmd/server/main.go

# Clean up generated files
clean:
	@echo "Cleaning up..."
	rm -f coverage.out coverage.html
	$(GO) clean -testcache

# Phony targets
.PHONY: all test test-coverage test-race lint run clean