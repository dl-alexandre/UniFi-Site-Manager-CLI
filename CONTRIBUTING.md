# Contributing to UniFi Site Manager CLI

Thank you for considering contributing! This document explains the architecture and patterns to ensure consistency.

## Architecture Overview

### The SiteManager Interface

All CLI commands interact with the `SiteManager` interface (`internal/pkg/api/interface.go`), not directly with HTTP clients. This abstraction allows the CLI to work with both Cloud and Local controllers without modification.

```go
type SiteManager interface {
    ListSites(pageSize int, nextToken string) (*SitesResponse, error)
    GetSite(siteID string) (*SiteResponse, error)
    // ... 33 total methods
}
```

**Key Rule**: When adding a new CLI command, you must:
1. Add the method to the `SiteManager` interface
2. Implement it in `CloudClient` (API implementation)
3. Stub it in `LocalClient` with `&NotImplementedError{Method: "MethodName"}`
4. Add a test case in `internal/pkg/cli/router_test.go`

### Testing Requirements

#### 1. Router Tests (Mandatory)
Every CLI command must have a test case in `TestCommandRouting`:

```go
{
    name: "Route: your new command",
    args: []string{"command", "subcommand", "--flag", "value"},
    mockSetup: func(m *mocks.SiteManager) {
        m.On("InterfaceMethod", expectedArgs...).Return(expectedReturn...)
    },
    expectedError: false,
}
```

#### 2. Interface Coverage (Automatic)
The `TestSiteManagerInterfaceCoverage` test automatically verifies that all 33 interface methods have corresponding CLI routes. If you add a method without a test, this will fail.

#### 3. Stub Method Tests (For LocalClient)
If your method is stubbed in `LocalClient`, add it to `TestLocalClientStubMethods`:

```go
{"Stub: YourMethod", "YourMethod", []interface{}{args...}, "not yet implemented"},
```

### The NotImplementedError Pattern

When stubbing a method for Local Controllers, always use the typed error:

```go
func (c *LocalClient) YourMethod(args) (*Response, error) {
    return nil, &NotImplementedError{Method: "YourMethod"}
}
```

This provides a helpful user message and consistent exit codes.

### Adding a New Feature

1. **Define the interface method** in `internal/pkg/api/interface.go`
2. **Implement for Cloud** in `internal/pkg/api/cloud_client.go`
3. **Stub for Local** in `internal/pkg/api/local_client.go` using `NotImplementedError`
4. **Add CLI command** in `internal/pkg/cli/<domain>.go`
5. **Add router test** in `internal/pkg/cli/router_test.go`
6. **Add stub test** (if applicable) in `TestLocalClientStubMethods`
7. **Update `TestSiteManagerInterfaceCoverage`** to include the new method
8. **Run tests**: `go test ./...`
9. **Run linter**: `go vet ./...`

### Local Controller API Mapping

When implementing a stubbed method for Local Controllers:

1. Use the debug flag to capture the JSON: `usm command --local --debug`
2. Add a custom unmarshaler if the JSON structure differs from Cloud API
3. Document the endpoint in comments
4. Add integration tests with captured responses

Example:
```go
// ListDevices retrieves all devices for a site
// Maps to: GET /proxy/network/api/s/{site}/stat/device
// Note: Local API does not support pagination
func (c *LocalClient) ListDevices(...) { ... }
```

### Code Style

- Follow Go conventions (gofmt, golint)
- Use table-driven tests
- Mock external dependencies
- Never commit credentials or tokens
- Redact sensitive data in debug output

### Commit Messages

Use conventional commits:
- `feat: add support for X`
- `fix: handle Y error case`
- `test: add coverage for Z`
- `docs: update README`

### Questions?

Open an issue with the `question` label. For bug reports, include:
- Command used
- Output with `--debug` flag
- Expected vs actual behavior
- Controller type (Cloud/Local) and firmware version

## Release Checklist

Before a new release:
- [ ] All tests pass (`go test ./...`)
- [ ] Build succeeds (`go build ./...`)
- [ ] Interface coverage at 100%
- [ ] CHANGELOG.md updated
- [ ] Version bumped in `cmd/usm/main.go`
- [ ] Release notes drafted in `.github/release-notes-vX.X.X.md`

---

**Remember**: The goal is 100% test coverage for the CLI-to-API routing layer. When in doubt, add a test!
