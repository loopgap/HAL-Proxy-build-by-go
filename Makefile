# BridgeOS Makefile

.PHONY: all build test clean run bridge bridgeosd install docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
BINARY_DIR=bin

# Binaries
BRIDGE_BINARY=$(BINARY_DIR)/bridge
BRIDGEOSD_BINARY=$(BINARY_DIR)/bridgeosd

# Build flags
LDFLAGS=-ldflags "-s -w"

all: test build

# Build all binaries
build: build-bridge build-bridgeosd

# Build bridge CLI
build-bridge:
	@echo "Building bridge CLI..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BRIDGE_BINARY) ./cmd/bridge
	@echo "Bridge CLI built successfully!"

# Build bridgeosd daemon
build-bridgeosd:
	@echo "Building bridgeosd daemon..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BRIDGEOSD_BINARY) ./cmd/bridgeosd
	@echo "BridgeOS daemon built successfully!"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...
	@echo "All tests passed!"

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific package tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	@echo "Cleaned!"

# Run bridge CLI
bridge: build-bridge
	@echo "Running bridge CLI..."
	./$(BRIDGE_BINARY) $(ARGS)

# Run bridgeosd daemon
bridgeosd: build-bridgeosd
	@echo "Running BridgeOS daemon..."
	./$(BRIDGEOSD_BINARY) $(ARGS)

# Development server
run-bridgeosd:
	BRIDGEOS_DB=./data/bridgeos.db BRIDGEOS_ADDR=:8080 $(GOCMD) run ./cmd/bridgeosd

# Install dependencies
install:
	$(GOMOD) download
	$(GOMOD) tidy

# Tidy modules
tidy:
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Run linter if available, otherwise skip
lint-check:
	@which golangci-lint > /dev/null && golangci-lint run ./... || echo "golangci-lint not found, skipping..."

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker build -t bridgeos/bridgeosd:latest -f Dockerfile.backend .
	docker build -t bridgeos/bridge:latest -f Dockerfile.cli .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# Generate API documentation
docs-api:
	@echo "Generating API documentation..."
	$(GOCMD) run ./cmd/apidocs

# Create release
release: test clean build
	@echo "Creating release..."
	cd $(BINARY_DIR) && \
	tar -czvf bridgeos-$(shell date +%Y%m%d).tar.gz bridge bridgeosd

# Show help
help:
	@echo "BridgeOS Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build         - Build all binaries"
	@echo "  make build-bridge - Build bridge CLI"
	@echo "  make build-bridgeosd - Build bridgeosd daemon"
	@echo "  make test         - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make run-bridgeosd - Run bridgeosd daemon"
	@echo "  make docker-build - Build Docker images"
	@echo "  make docker-run   - Run with Docker Compose"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
	@echo "  make help         - Show this help"

# Development shortcuts
dev: install fmt test

# CI pipeline (for GitHub Actions)
ci: fmt lint-check test
