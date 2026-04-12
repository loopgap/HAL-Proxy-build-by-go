# BridgeOS Makefile

.PHONY: all build test clean run bridge bridgeosd install docker-build docker-run fmt lint lint-check
.PHONY: frontend frontend-install frontend-build frontend-test frontend-lint
.PHONY: cross-platform cross-platform-build ci ci-full help

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
VERSION?=dev

# Cross-platform build parameters
GOOS?=linux
GOARCH?=amd64

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
	@echo "  make build           - Build all binaries"
	@echo "  make build-bridge    - Build bridge CLI"
	@echo "  make build-bridgeosd - Build bridgeosd daemon"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run-bridgeosd   - Run bridgeosd daemon"
	@echo "  make docker-build    - Build Docker images"
	@echo "  make docker-run      - Run with Docker Compose"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Run linter"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend          - Install and build frontend"
	@echo "  make frontend-install - Install frontend dependencies"
	@echo "  make frontend-test    - Run frontend tests"
	@echo "  make frontend-lint    - Run frontend linter"
	@echo "  make frontend-typecheck - Run TypeScript check"
	@echo "  make frontend-build   - Build frontend"
	@echo ""
	@echo "Cross-platform:"
	@echo "  make cross-platform          - Build for current platform"
	@echo "  make build-all-platforms     - Build for all platforms"
	@echo "  GOOS=linux GOARCH=amd64 make cross-platform - Specify platform"
	@echo ""
	@echo "CI:"
	@echo "  make ci        - Quick CI check (fmt, lint, test)"
	@echo "  make ci-full  - Full CI including frontend"
	@echo "  make help     - Show this help"

# Development shortcuts
dev: install fmt test frontend-install

# CI pipeline (for GitHub Actions)
ci: fmt lint-check test

# ============ Frontend Commands ============
frontend: frontend-install frontend-build

frontend-install:
	@echo "Installing frontend dependencies..."
	cd ui && npm ci

frontend-test:
	@echo "Running frontend tests..."
	cd ui && npm test -- --run

frontend-lint:
	@echo "Running frontend linter..."
	cd ui && npm run lint

frontend-typecheck:
	@echo "Running TypeScript check..."
	cd ui && npx tsc --noEmit

frontend-build:
	@echo "Building frontend..."
	cd ui && npm run build

# ============ Cross-Platform Builds ============
cross-platform: cross-platform-build

cross-platform-build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	mkdir -p $(BINARY_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/bridge-$(GOOS)-$(GOARCH) ./cmd/bridge
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/bridgeosd-$(GOOS)-$(GOARCH) ./cmd/bridgeosd
	@echo "Built for $(GOOS)/$(GOARCH)"

build-all-platforms:
	@echo "Building for all platforms..."
	$(MAKE) cross-platform-build GOOS=linux GOARCH=amd64
	$(MAKE) cross-platform-build GOOS=linux GOARCH=arm64
	$(MAKE) cross-platform-build GOOS=darwin GOARCH=amd64
	$(MAKE) cross-platform-build GOOS=darwin GOARCH=arm64
	$(MAKE) cross-platform-build GOOS=windows GOARCH=amd64

# ============ Full CI ============
ci-full: fmt lint-check frontend-typecheck test frontend-lint
