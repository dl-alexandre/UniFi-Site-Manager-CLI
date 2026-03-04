# Changelog

All notable changes to UniFi Controller Home Assistant integration will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation
- Enhanced error messages
- Improved entity naming
- Additional device tracker attributes

### Changed
- Updated for Home Assistant 2024.x
- Improved WebSocket reconnection logic
- Better handling of controller disconnections

### Fixed
- Entity availability on startup
- WebSocket connection stability
- Device tracker state updates

## [1.0.0] - 2024-01-15

### Added
- Initial release
- Sensor platform with:
  - Device count sensors (online, offline, total)
  - Client count sensors (total, wired, wireless, guest)
  - Performance sensors (CPU, memory, bandwidth, latency)
  - Alert counter
- Device tracker platform:
  - Automatic discovery of connected clients
  - Real-time state updates via WebSocket
  - Client attributes (IP, MAC, hostname, signal, etc.)
- Binary sensor platform:
  - Site health status
  - WAN connectivity
  - Firmware update availability per device
- Service calls:
  - Restart device
  - Block/unblock client
  - Reconnect client
  - Enable/disable WLAN
- Configuration flow (UI setup)
- Options flow (reconfiguration)
- Support for UniFi OS (UDM/UDR/Cloud Key)
- Support for standalone controllers
- WebSocket real-time updates
- Automatic reconnection handling
- Custom events for automations

### Security
- Local API connection (no cloud required)
- Support for self-signed certificates
- No credentials stored in YAML

## [0.9.0] - 2024-01-01 (Beta)

### Added
- Beta support for device trackers
- Client blocking/unblocking
- Real-time WebSocket events
- Multi-site support

### Changed
- Improved API client reliability
- Better error handling

### Fixed
- Authentication with UniFi OS 2.x
- Entity ID generation

## [0.8.0] - 2023-12-01 (Beta)

### Added
- Beta sensor platform
- Device and client counting
- Health monitoring
- Alert notifications

### Fixed
- Connection timeout handling
- SSL certificate issues

## [0.1.0] - 2023-10-01 (Alpha)

### Added
- Initial alpha release
- Basic API client
- Proof of concept integration
- README and documentation

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

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes requiring user action
- **MINOR** (0.X.0): New features, backward compatible
- **PATCH** (0.0.X): Bug fixes, backward compatible

## Migration Guide

### From v0.x to v1.0

**Breaking Changes**:
- Entity IDs may have changed
- Some attributes renamed

**Migration Steps**:
1. Note existing entity IDs
2. Update to v1.0
3. Check new entity IDs in Developer Tools
4. Update automations/dashboards with new IDs
5. Disable old integration if migrating from other UniFi integration

### From Official UniFi Integration

**Differences**:
- Different entity ID format
- Additional entities available
- More real-time updates

**Migration Steps**:
1. Install this integration alongside official (for testing)
2. Rename entities to match or update automations
3. Disable official integration when ready
4. Update all entity references

## Support Policy

- **Latest major version**: Full support, bug fixes, features
- **Previous major version**: Security fixes only
- **Older versions**: No support

## Deprecation Notices

Deprecated features will be:
1. Marked as deprecated in release notes
2. Emit warnings when used
3. Removed in next major version

---

For the complete list of changes, see the [GitHub commits](https://github.com/dl-alexandre/Local-UniFi-CLI/commits/main).
