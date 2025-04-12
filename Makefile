# Go command to use
GO := go

# Default target
all: test lint run-database

# Run all tests
test:
	@echo "Running all tests..."
	$(GO) clean -testcache
	$(GO) test ./... -race

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run linter (golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Build and run the database
run-database:
	@echo "Building and running the database..."
	$(GO) run cmd/server/main.go

# Build and run the database client
run-client:
	@echo "Building and running database client..."
	$(GO) run cmd/client/main.go

# Clean up generated files
clean:
	@echo "Cleaning up..."
	rm -f coverage.out coverage.html
	$(GO) clean -testcache

# Phony targets
.PHONY: all test test-coverage lint run-database clean