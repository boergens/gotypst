# gotypst Makefile

.PHONY: all build test test-fixtures test-verbose clean fmt lint

# Default target
all: build test

# Build the project
build:
	go build ./...

# Run all tests
test:
	go test ./...

# Run fixture tests specifically
test-fixtures:
	go test -v ./tests -run "Test.*Fixture"

# Run all tests with verbose output
test-verbose:
	go test -v ./...

# Run tests for a specific category
# Usage: make test-category CATEGORY=syntax
test-category:
	go test -v ./tests -run "TestHarness_LoadFixturesByCategory/$(CATEGORY)"

# Run benchmark tests
bench:
	go test -bench=. ./tests

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	go clean ./...

# Show test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install development dependencies
deps:
	go mod tidy
