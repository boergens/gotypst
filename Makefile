# GoTypst Makefile

.PHONY: all test test-fixtures test-syntax test-foundations test-scripting clean lint

# Default target
all: test

# Run all tests
test:
	go test ./...

# Run fixture tests only
test-fixtures:
	go test -v ./tests -run 'Test.*Fixtures'

# Run syntax fixture tests
test-syntax:
	go test -v ./tests -run 'TestSyntaxFixtures'

# Run foundations fixture tests
test-foundations:
	go test -v ./tests -run 'TestFoundationsFixtures'

# Run scripting fixture tests
test-scripting:
	go test -v ./tests -run 'TestScriptingFixtures'

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./tests

# Lint the code
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go vet ./...; \
	fi

# Format the code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -testcache

# Show help
help:
	@echo "Available targets:"
	@echo "  make test           - Run all tests"
	@echo "  make test-fixtures  - Run fixture tests only"
	@echo "  make test-syntax    - Run syntax fixture tests"
	@echo "  make test-foundations - Run foundations fixture tests"
	@echo "  make test-scripting - Run scripting fixture tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make bench          - Run benchmarks"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
