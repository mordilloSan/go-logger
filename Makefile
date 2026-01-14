.PHONY: test fmt vet all clean help test-concurrency test-progress

# Default target
all: fmt vet test

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run concurrency tests with real-time progress display
test-concurrency:
	@echo "Running concurrency test with real-time progress..."
	@echo "Watch the mutex prevent garbled output from 100+ concurrent goroutines!"
	@echo ""
	@go test -v -run TestConcurrency_RealTimeProgress ./logger

# Run only concurrency tests (without real-time display)
test-concurrency-all:
	@echo "Running all concurrency tests..."
	@go test -v -run TestConcurrency ./logger

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run static analysis
vet:
	@echo "Running go vet..."
	@go vet ./...

# Clean build artifacts and cache
clean:
	@echo "Cleaning..."
	@go clean -cache -testcache

# Pre-release check: format, vet, and test
pre-release: fmt vet test
	@echo "✓ All pre-release checks passed!"
	@echo "Ready to create release tag"

# Help target
help:
	@echo "Available targets:"
	@echo "  make test              - Run all tests"
	@echo "  make test-concurrency  - Demo real-time concurrent logging (100 goroutines)"
	@echo "  make fmt               - Format code"
	@echo "  make vet               - Run static analysis"
	@echo "  make all               - Run fmt, vet, and test (default)"
	@echo "  make pre-release       - Run all checks before release"
	@echo "  make clean             - Clean build cache"
	@echo "  make help              - Show this help message"
