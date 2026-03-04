# Changelog

All notable changes to UniFi Site Manager CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Enhanced documentation with comprehensive guides
- Docker support with official images
- New examples directory with automation scripts
- Improved error messages with suggestions
- Debug mode with credential redaction

### Changed
- Updated to Go 1.24
- Improved CLI help text
- Better formatting for table output

## [1.0.0] - 2024-01-15

### Added
- Full Site Manager API coverage
- Dual mode support (Cloud API and Local Controller)
- Complete site management (CRUD operations)
- Host/Console management
- Device management (list, get, restart, upgrade, adopt)
- Client management (list, stats, block/unblock)
- WLAN management (CRUD operations)
- Alert management (list, ack, archive)
- Event viewing
- Network listing
- JSON and table output formats
- Environment variable configuration
- Configuration file support
- Comprehensive test suite
- CI/CD with GitHub Actions
- Automated releases with GoReleaser
- Homebrew tap support
- Shell completion support
- Rate limiting with automatic retry

### Changed
- Improved error handling with specific error types
- Better HTTP client with retry logic
- Enhanced table formatting

### Fixed
- Rate limit handling
- TLS certificate verification for local controllers
- CSRF token handling for UniFi OS
- Memory leaks in long-running processes

## [0.9.0] - 2024-01-01

### Added
- Beta support for local UniFi OS controllers
- Device adoption support
- Client blocking/unblocking
- Host restart functionality
- Alert management
- Event viewing
- Network listing

### Changed
- Refactored API client architecture
- Improved error messages
- Better debug output

## [0.8.0] - 2023-12-01

### Added
- WLAN management (create, update, delete)
- Client statistics
- Site statistics (day/week/month periods)
- Host statistics
- Improved filtering for devices

### Fixed
- Pagination handling for large sites
- JSON output formatting
- Color output on Windows

## [0.7.0] - 2023-11-15

### Added
- Device restart functionality
- Device upgrade support
- Client filtering (wired/wireless)
- Search functionality

### Changed
- Improved CLI structure with Kong
- Better help documentation

## [0.6.0] - 2023-11-01

### Added
- Site health monitoring
- Host health monitoring
- Site creation and deletion
- Configuration initialization (`usm init`)

### Fixed
- Authentication error handling
- Network timeout issues

## [0.5.0] - 2023-10-15

### Added
- Basic site management (list, get, update)
- Host management (list, get)
- Device listing
- Client listing
- WLAN listing
- Alert listing
- Event listing
- Network listing
- Whoami command
- Version command

### Changed
- Initial CLI framework with Kong
- HTTP client with Resty
- Basic table output formatting

## [0.1.0] - 2023-10-01

### Added
- Initial project structure
- Basic API client
- Configuration management
- Makefile with common tasks
- README documentation
- MIT License

---

## Release Notes Template

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Now removed features

### Fixed
- Bug fixes

### Security
- Security improvements
```

## Versioning Guide

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Migration Guide

### From v0.x to v1.0

Breaking changes:
- Configuration file format changed
- Some environment variable names changed
- CLI flags reorganized

Migration steps:
1. Backup your config: `cp ~/.config/usm/config.yaml ~/.config/usm/config.yaml.bak`
2. Run `usm init` to create new config
3. Update scripts using old flag names
4. Test thoroughly before production use
