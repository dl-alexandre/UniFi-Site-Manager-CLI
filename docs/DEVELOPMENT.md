# Development Guide

Guide for contributing to and developing UniFi Site Manager CLI.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Building](#building)
- [Debugging](#debugging)
- [Contributing](#contributing)
- [Release Process](#release-process)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- **Go**: Version 1.24 or later
- **Git**: For version control
- **Make**: For build automation (optional)
- **Docker**: For containerized testing (optional)

### Verify Installation

```bash
go version  # Should show go1.24 or later
git --version
make --version
```

## Getting Started

### Clone Repository

```bash
git clone https://github.com/dl-alexandre/UniFi-Site-Manager-CLI.git
cd UniFi-Site-Manager-CLI
```

### Install Dependencies

```bash
# Using make
make deps

# Or directly with go
go mod download
go mod tidy
```

### Build Project

```bash
# Development build
make build

# Or manually
go build -o usm ./cmd/usm
```

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/pkg/api/...
```

## Project Structure

```
.
├── cmd/
│   └── usm/
│       └── main.go           # Entry point
├── internal/
│   ├── cli/
│   │   └── update.go         # Version checking
│   ├── cache/
│   │   └── cache.go          # Caching layer
│   └── pkg/
│       ├── api/              # API client implementations
│       │   ├── client.go     # Cloud API client
│       │   ├── local_client.go # Local API client
│       │   ├── interface.go  # SiteManager interface
│       │   ├── models.go     # Data structures
│       │   ├── sites.go      # Sites endpoints
│       │   ├── devices.go    # Devices endpoints
│       │   ├── clients.go    # Clients endpoints
│       │   ├── wlans.go      # WLAN endpoints
│       │   ├── alerts.go     # Alerts endpoints
│       │   └── errors.go     # Error types
│       ├── cli/              # CLI commands
│       │   ├── cli.go        # Root command & context
│       │   ├── sites.go      # Sites subcommands
│       │   ├── devices.go    # Devices subcommands
│       │   ├── clients.go    # Clients subcommands
│       │   ├── wlans.go      # WLAN subcommands
│       │   ├── hosts.go      # Hosts subcommands
│       │   ├── alerts.go     # Alerts subcommands
│       │   ├── events.go     # Events subcommands
│       │   ├── networks.go   # Networks subcommands
│       │   ├── whoami.go     # Auth subcommand
│       │   ├── version.go    # Version subcommand
│       │   └── init.go       # Config initialization
│       ├── config/           # Configuration management
│       │   └── config.go     # Config loading
│       └── output/           # Output formatting
│           └── formatter.go  # Table/JSON formatting
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── Makefile                  # Build automation
├── README.md                 # User documentation
├── LICENSE                   # MIT License
└── docs/                     # Documentation
```

## Development Workflow

### 1. Create Branch

```bash
git checkout -b feature/my-new-feature
# or
git checkout -b fix/issue-description
```

### 2. Make Changes

Edit files, following the [Code Style](#code-style) guidelines.

### 3. Test Changes

```bash
# Run tests
make test

# Test specific functionality
./usm sites list --output json
```

### 4. Lint and Format

```bash
# Run linter
make lint

# Format code
make format

# Or use golangci-lint directly
golangci-lint run
```

### 5. Commit Changes

```bash
git add .
git commit -m "feat: add new feature description"
```

Commit message format:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Tests
- `refactor:` Code refactoring
- `chore:` Maintenance

### 6. Push and Create PR

```bash
git push origin feature/my-new-feature
```

Then create a Pull Request on GitHub.

## Testing

### Run Tests

```bash
# All tests
make test

# With verbose output
go test -v ./...

# Specific package
go test ./internal/pkg/api/...

# Race condition detection
go test -race ./...
```

### Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
go tool cover -html=coverage.out
```

### Writing Tests

```go
// Example test in internal/pkg/api/sites_test.go
func TestListSites(t *testing.T) {
    client := NewMockClient()
    sites, err := client.ListSites(50, 0)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, sites)
}
```

### Integration Tests

```bash
# Run with real API (requires API key)
USM_API_KEY=test-key go test -tags=integration ./...
```

## Code Style

### Go Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Maximum line length: 100 characters
- Use meaningful variable names

### Project Conventions

```go
// Package naming
package api    // Descriptive, not generic

// Error handling
if err != nil {
    return nil, fmt.Errorf("failed to list sites: %w", err)
}

// Context handling
func (c *Client) ListSites(ctx context.Context, ...) ([]Site, error) {
    // Use context for cancellation
}

// Interface naming
type SiteManager interface {
    ListSites(...) ([]Site, error)
}

// Mock naming
type MockSiteManager struct {
    mock.Mock
}
```

### Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Exported | PascalCase | `ListSites` |
| Unexported | camelCase | `listSites` |
| Interfaces | noun ending in "er" | `SiteManager` |
| Structs | nouns | `SiteResponse` |
| Constants | CamelCase or UPPER_SNAKE | `DefaultTimeout` |
| Acronyms | All caps | `APIKey`, `URL` |

## Building

### Development Build

```bash
make build

# Or manually
go build -o usm ./cmd/usm
```

### Production Build

```bash
# Optimized build
go build -ldflags="-s -w" -o usm ./cmd/usm

# With version info
go build -ldflags="-s -w -X main.version=v1.0.0 -X main.gitCommit=abc123" -o usm ./cmd/usm
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o usm-linux-amd64 ./cmd/usm

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o usm-darwin-amd64 ./cmd/usm

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o usm-darwin-arm64 ./cmd/usm

# Windows
GOOS=windows GOARCH=amd64 go build -o usm-windows-amd64.exe ./cmd/usm
```

### Using Makefile

```bash
# Available targets
make help

# Common targets
make build          # Build binary
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make format         # Format code
make clean          # Clean build artifacts
make deps           # Download dependencies
make release        # Create release (requires GITHUB_TOKEN)
make snapshot       # Test release without publishing
```

## Debugging

### Enable Debug Mode

```bash
# CLI flag
./usm --debug sites list

# Environment variable
USM_DEBUG=true ./usm sites list
```

### Debug Output

Debug mode shows:
- HTTP requests (URLs, methods)
- Request headers (credentials redacted)
- Request/response bodies
- API call timing

### Using a Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/pkg/api/

# Debug main
dlv debug ./cmd/usm

# Inside debugger:
# (dlv) break main.main
# (dlv) continue
# (dlv) print variable
# (dlv) step
```

### Logging

```go
// Add logging in code
if ctx.Debug {
    fmt.Fprintf(os.Stderr, "[DEBUG] Request: %s %s\n", method, url)
}
```

### Profiling

```bash
# CPU profiling
go build -o usm ./cmd/usm
./usm -cpuprofile=cpu.prof sites list
go tool pprof cpu.prof

# Memory profiling
go build -o usm ./cmd/usm
./usm -memprofile=mem.prof sites list
go tool pprof mem.prof
```

## Contributing

### Before Contributing

1. Check existing issues and PRs
2. Open an issue to discuss major changes
3. Follow the [Code of Conduct](../CODE_OF_CONDUCT.md)

### Contribution Process

1. **Fork** the repository
2. **Clone** your fork
3. **Create** a feature branch
4. **Make** your changes
5. **Test** thoroughly
6. **Commit** with clear messages
7. **Push** to your fork
8. **Create** a Pull Request

### PR Guidelines

- Provide clear description
- Reference related issues
- Include tests for new features
- Update documentation if needed
- Ensure CI passes
- Keep changes focused

### Code Review

- All PRs require review
- Address feedback promptly
- Keep discussion technical and respectful

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- e.g., `v1.2.3`

### Creating a Release

```bash
# 1. Ensure all tests pass
make ci

# 2. Update CHANGELOG.md

# 3. Commit changes
git add CHANGELOG.md
git commit -m "chore: update changelog for v1.1.0"

# 4. Create tag
git tag v1.1.0

# 5. Push tag
git push origin v1.1.0

# 6. Create release (with GoReleaser)
export GITHUB_TOKEN=ghp_your_token
make release
```

### Testing Releases

```bash
# Create snapshot without publishing
make snapshot

# Test binary
./dist/usm-linux-amd64/usm version
```

### Homebrew Tap

Releases are automatically published to the Homebrew tap:
- [dl-alexandre/homebrew-tap](https://github.com/dl-alexandre/homebrew-tap)

## Troubleshooting

### Build Errors

```bash
# Clean and rebuild
make clean
make deps
make build

# Verify Go version
go version  # Must be 1.24+

# Update dependencies
go mod tidy
go mod download
```

### Test Failures

```bash
# Run specific test
go test -v -run TestListSites ./internal/pkg/api/

# Debug test
dlv test ./internal/pkg/api/
```

### Lint Errors

```bash
# Install linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run with fix attempts
golangci-lint run --fix
```

### Import Errors

```bash
# Fix imports
goimports -w .

# Or use make
make format
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go by Example](https://gobyexample.com/)
- [Kong Documentation](https://github.com/alecthomas/kong)
- [Resty Documentation](https://github.com/go-resty/resty)

## Questions?

- Open an issue for bugs
- Start a discussion for questions
- Check existing documentation

Thank you for contributing to UniFi Site Manager CLI!
