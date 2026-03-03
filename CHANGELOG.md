# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.2] - 2026-03-02

### Added
- **Local Controller Support (Beta)**: Direct connection to UniFi OS controllers (UDM, UDM-Pro, UDR)
  - Session-based authentication with cookie jar and CSRF token handling
  - Self-signed certificate support for local controllers
  - Endpoint mapping from Cloud API to Local API proxy paths
  - Site resolution helper (defaults to "default" site)
- Debug logging with credential redaction
  - `--debug` flag for verbose API logging
  - Automatic redaction of passwords, CSRF tokens, cookies, and API keys
  - Safe for sharing debug output in bug reports
- Custom JSON unmarshaler for Device and NetworkClient
  - Handles field mapping differences between Cloud and Local APIs
  - Cloud uses: `id` (string), `status` (string)
  - Local uses: `_id` (MongoDB ObjectID), `state` (int: 1=ONLINE, 0=OFFLINE)
- WLAN CRUD operations for local controllers
  - `ListWLANs()`, `CreateWLAN()`, `GetWLAN()`, `UpdateWLAN()`, `DeleteWLAN()`
  - CSRF token auto-injection for mutating operations
- Device and Client observability for local controllers
  - `ListDevices()`, `GetDevice()`, `RestartDevice()`
  - `ListClients()` with wired/wireless filtering

### Infrastructure
- Enhanced CLI routing logic with mode detection (Cloud vs Local)
- Kong flag groups for better `--help` organization
- Environment variable support for all authentication parameters
  - `USM_LOCAL` - Enable local mode
  - `USM_HOST` - Local controller IP/hostname
  - `USM_USERNAME` - Local controller username
  - `USM_PASSWORD` - Local controller password (use env var for security)
- Comprehensive beta testing documentation in README

### Changed
- Migrated from 3,500-line monolith to domain-driven architecture
- Split API client into Cloud and Local implementations
- Both implement common `SiteManager` interface

## [0.0.1] - 2026-03-02

### Added
- Project initialization
- Basic CLI structure with Kong
- API client foundation with resty
- Configuration management with Viper
- Initial command structure

[Unreleased]: https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/compare/v0.0.2...HEAD
[0.0.2]: https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/tag/v0.0.2
[0.0.1]: https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/tag/v0.0.1
