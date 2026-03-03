.PHONY: build build-all build-linux build-darwin build-windows test test-coverage lint release snapshot clean format install install-hooks mocks deps ci help check-release uninstall

BINARY_NAME=usm
MAIN_PATH=./cmd/usm
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME) -s -w"

# Default target
all: test build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

# Build for all platforms (local cross-compilation)
build-all: build-linux build-darwin build-windows

# Linux builds
build-linux:
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Linux builds complete"

# macOS builds
build-darwin:
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "macOS builds complete"

# Windows builds
build-windows:
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Windows builds complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Install goreleaser
install-goreleaser:
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser/v2@latest)

# Create a snapshot release (local build, no publishing)
snapshot: install-goreleaser test
	@echo "Creating snapshot build..."
	goreleaser release --snapshot --clean

# Check release configuration
check-release: install-goreleaser
	@echo "Checking release configuration..."
	goreleaser check

# Create a full release (requires GITHUB_TOKEN and git tag)
release: install-goreleaser test check-release
	@echo "Creating release..."
	@if [ -z "$${GITHUB_TOKEN}" ]; then \
		echo "Error: GITHUB_TOKEN environment variable is required"; \
		echo "Set it with: export GITHUB_TOKEN=your_token_here"; \
		exit 1; \
	fi
	@if [ -z "$(shell git describe --tags --exact-match 2>/dev/null)" ]; then \
		echo "Error: No git tag found on current commit"; \
		echo "Create a tag first: git tag v1.0.0"; \
		echo "Then push it: git push origin v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating release for tag: $(shell git describe --tags --exact-match)"
	goreleaser release --clean

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
format:
	@echo "Formatting code..."
	@gofmt -w -s .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@which mockery > /dev/null || (echo "Installing mockery..." && go install github.com/vektra/mockery/v2@latest)
	mockery --name=SiteManager --dir=internal/pkg/api --output=internal/pkg/mocks --outpkg=mocks
	@echo "Mocks generated in internal/pkg/mocks/"

# Install locally
install: build
	@echo "Installing to GOPATH/bin..."
	go install $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

# Uninstall
uninstall:
	@echo "Uninstalling from GOPATH/bin..."
	rm -f $$(go env GOPATH)/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Run CI pipeline locally
ci: deps format lint test build
	@echo "✓ CI pipeline complete"

# Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@git config core.hooksPath .githooks
	@echo "Hooks installed from .githooks/"

# Show help
help:
	@echo "UniFi Site Manager CLI - Makefile"
	@echo ""
	@echo "BUILD TARGETS:"
	@echo "  make build          - Build binary for current platform"
	@echo "  make build-all      - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make build-linux    - Build for Linux (amd64, arm64)"
	@echo "  make build-darwin   - Build for macOS (amd64, arm64)"
	@echo "  make build-windows  - Build for Windows (amd64)"
	@echo ""
	@echo "TEST TARGETS:"
	@echo "  make test           - Run all tests with race detection"
	@echo "  make test-coverage  - Run tests and generate coverage report"
	@echo ""
	@echo "RELEASE TARGETS:"
	@echo "  make snapshot       - Create snapshot release (goreleaser, local only)"
	@echo "  make release        - Create official release (requires GITHUB_TOKEN)"
	@echo "  make check-release  - Validate .goreleaser.yaml configuration"
	@echo ""
	@echo "QUALITY TARGETS:"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make format         - Format code with gofmt and goimports"
	@echo "  make mocks          - Generate test mocks with mockery"
	@echo ""
	@echo "MAINTENANCE TARGETS:"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make ci             - Run full CI pipeline locally"
	@echo "  make install        - Install binary to GOPATH/bin"
	@echo "  make uninstall      - Remove binary from GOPATH/bin"
	@echo "  make install-hooks  - Install git hooks"
	@echo ""
	@echo "RELEASE PROCESS:"
	@echo "  1. make ci          - Ensure everything passes locally"
	@echo "  2. Update CHANGELOG.md with new version"
	@echo "  3. git commit -am 'Release vX.Y.Z'"
	@echo "  4. git tag vX.Y.Z"
	@echo "  5. git push origin vX.Y.Z"
	@echo "  6. export GITHUB_TOKEN=ghp_xxxxxxxx"
	@echo "  7. make release     - GoReleaser handles the rest!"
	@echo ""
	@echo "QUICK START:"
	@echo "  make all            - Run tests and build (default)"
	@echo "  make help           - Show this help message"
