# usm - UniFi Site Manager CLI

A full-featured command-line interface for the [UniFi Site Manager API](https://developer.ui.com/site-manager/v1.0.0/gettingstarted).

## Overview

`usm` provides a fast, intuitive interface for managing UniFi sites, hosts, devices, clients, WLANs, alerts, and more via the official Site Manager API at `api.ui.com`. This tool focuses on cloud-based site management - for local controller APIs (Network, Protect), see the separate `unifi` CLI.

## Features

- **Sites**: Create, update, delete, list, get details, health, statistics
- **Hosts/Consoles**: List, get details, health, statistics, restart
- **Devices**: List, get details, restart, upgrade firmware, adopt new devices
- **Clients**: List (wired/wireless), view statistics, block/unblock
- **WLANs**: Create, update, delete, list, get details
- **Alerts**: List, acknowledge, archive
- **Events**: View system events
- **Networks**: List configured networks
- **Full API Coverage**: Health monitoring, performance statistics, firmware management

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-darwin-arm64 -o usm

# macOS (Intel)
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-darwin-amd64 -o usm

# Linux (amd64)
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm

# Linux (arm64)
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-arm64 -o usm

chmod +x usm
sudo mv usm /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/dl-alexandre/UniFi-Site-Manager-CLI.git
cd UniFi-Site-Manager-CLI
make build
```

## Quick Start

1. **Get an API Key** from [unifi.ui.com](https://unifi.ui.com) → Settings → Control Plane → Integrations → API Keys

2. **Configure**:
   ```bash
   usm init
   # Or set environment variable
   export USM_API_KEY="your-api-key"
   ```

3. **Verify**:
   ```bash
   usm whoami
   ```

4. **List sites**:
   ```bash
   usm sites list
   ```

## Release Process

This project uses [GoReleaser](https://goreleaser.com/) for automated releases.

### For Maintainers

To create a new release:

1. **Ensure CI passes**:
   ```bash
   make ci
   ```

2. **Update CHANGELOG.md** with new version details

3. **Commit and tag**:
   ```bash
   git commit -am "Release v1.0.0"
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **Set GitHub token** (requires repo write access):
   ```bash
   export GITHUB_TOKEN=ghp_your_token_here
   ```

5. **Create release**:
   ```bash
   make release
   ```

GoReleaser will:
- Run tests
- Build binaries for all platforms (Linux, macOS, Windows; AMD64, ARM64)
- Create archives and checksums
- Publish to GitHub Releases
- Update Homebrew tap (if configured)

### Testing Releases Locally

Test without publishing:
```bash
make snapshot
```

This creates binaries in `dist/` for local testing.

## Commands

### `usm init`
Interactive configuration setup. Creates `~/.config/usm/config.yaml`.

```bash
usm init          # First-time setup
usm init --force # Overwrite existing config
```

### `usm whoami`
Show authenticated user account information.

```bash
usm whoami
usm whoami --output json
```

### `usm version`
Show version information.

```bash
usm version
usm version --check  # Check for updates (not yet implemented)
```

## Site Management

### `usm sites list`
List all sites with pagination and filtering.

```bash
usm sites list                          # Default: 50 sites per page
usm sites list --page-size 100          # Get 100 sites
usm sites list --page-size 0            # Fetch all sites
usm sites list --search "office"        # Filter by name/description
usm sites list --output json            # JSON output
usm sites list --no-headers             # Table without headers (for scripts)
```

### `usm sites get <site-id>`
Get detailed information about a specific site.

```bash
usm sites get 60abcdef1234567890abcdef
```

### `usm sites create <name>`
Create a new site.

```bash
usm sites create "New Office" --description "Main office location"
```

### `usm sites update <site-id>`
Update an existing site.

```bash
usm sites update 60abcdef1234567890abcdef --name "Updated Name" --description "New description"
```

### `usm sites delete <site-id>`
Delete a site (with confirmation).

```bash
usm sites delete 60abcdef1234567890abcdef
usm sites delete 60abcdef1234567890abcdef --force  # Skip confirmation
```

### `usm sites health <site-id>`
Get site health information.

```bash
usm sites health 60abcdef1234567890abcdef
```

### `usm sites stats <site-id>`
Get site performance statistics.

```bash
usm sites stats 60abcdef1234567890abcdef              # Current stats
usm sites stats 60abcdef1234567890abcdef --period day   # Daily stats
usm sites stats 60abcdef1234567890abcdef --period week  # Weekly stats
usm sites stats 60abcdef1234567890abcdef --period month # Monthly stats
```

## Host/Console Management

### `usm hosts list`
List all UniFi hosts/consoles.

```bash
usm hosts list                          # Default: 50 hosts per page
usm hosts list --page-size 0            # Fetch all hosts
usm hosts list --search "UDM"         # Filter by name
```

### `usm hosts get <host-id>`
Get detailed information about a specific host.

```bash
usm hosts get 60abcdef1234567890abcdef
```

### `usm hosts health <host-id>`
Get host health information.

```bash
usm hosts health 60abcdef1234567890abcdef
```

### `usm hosts stats <host-id>`
Get host performance statistics.

```bash
usm hosts stats 60abcdef1234567890abcdef --period day
```

### `usm hosts restart <host-id>`
Restart a host/console.

```bash
usm hosts restart 60abcdef1234567890abcdef
usm hosts restart 60abcdef1234567890abcdef --force  # Skip confirmation
```

## Device Management

### `usm devices list <site-id>`
List all devices for a site.

```bash
usm devices list 60abcdef1234567890abcdef
usm devices list 60abcdef1234567890abcdef --page-size 100
usm devices list 60abcdef1234567890abcdef --status online    # Filter by status
usm devices list 60abcdef1234567890abcdef --type ap          # Filter by type (ap, switch, gateway)
```

### `usm devices get <site-id> <device-id>`
Get detailed information about a specific device.

```bash
usm devices get 60abcdef1234567890abcdef 60fedcba0987654321fedcba
```

### `usm devices restart <site-id> <device-id>`
Restart a device.

```bash
usm devices restart 60abcdef1234567890abcdef 60fedcba0987654321fedcba
```

### `usm devices upgrade <site-id> <device-id>`
Upgrade device firmware.

```bash
usm devices upgrade 60abcdef1234567890abcdef 60fedcba0987654321fedcba
```

### `usm devices adopt <site-id> <mac-address>`
Adopt a new device to a site.

```bash
usm devices adopt 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

## Client Management

### `usm clients list <site-id>`
List all clients for a site.

```bash
usm clients list 60abcdef1234567890abcdef
usm clients list 60abcdef1234567890abcdef --wired-only      # Show only wired clients
usm clients list 60abcdef1234567890abcdef --wireless-only     # Show only wireless clients
usm clients list 60abcdef1234567890abcdef --search "iPhone"  # Filter by hostname/name
```

### `usm clients stats <site-id> <mac-address>`
Get statistics for a specific client.

```bash
usm clients stats 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

### `usm clients block <site-id> <mac-address>`
Block a client from the network.

```bash
usm clients block 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

### `usm clients unblock <site-id> <mac-address>`
Unblock a previously blocked client.

```bash
usm clients unblock 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

## WLAN Management

### `usm wlans list <site-id>`
List all wireless networks for a site.

```bash
usm wlans list 60abcdef1234567890abcdef
```

### `usm wlans get <site-id> <wlan-id>`
Get detailed information about a specific WLAN.

```bash
usm wlans get 60abcdef1234567890abcdef 60wlan0987654321abcdef
```

### `usm wlans create <site-id> <name> <ssid>`
Create a new wireless network.

```bash
usm wlans create 60abcdef1234567890abcdef "Office WiFi" "Office-5G" \
  --password "secure-password" \
  --security wpapsk \
  --vlan 10 \
  --band both \
  --wpa3
```

### `usm wlans update <site-id> <wlan-id>`
Update an existing WLAN.

```bash
usm wlans update 60abcdef1234567890abcdef 60wlan0987654321abcdef \
  --password "new-password" \
  --enabled=false
```

### `usm wlans delete <site-id> <wlan-id>`
Delete a WLAN.

```bash
usm wlans delete 60abcdef1234567890abcdef 60wlan0987654321abcdef
```

## Alert Management

### `usm alerts list`
List all alerts (optionally filtered by site).

```bash
usm alerts list                        # All alerts
usm alerts list --site-id 60abcdef1234567890abcdef  # Site-specific alerts
usm alerts list --archived             # Show archived alerts
```

### `usm alerts ack <site-id> <alert-id>`
Acknowledge an alert.

```bash
usm alerts ack 60abcdef1234567890abcdef 60alert0987654321abcdef
```

### `usm alerts archive <site-id> <alert-id>`
Archive an alert.

```bash
usm alerts archive 60abcdef1234567890abcdef 60alert0987654321abcdef
```

## Event Management

### `usm events list`
List system events.

```bash
usm events list                        # All events
usm events list --site-id 60abcdef1234567890abcdef  # Site-specific events
```

## Network Management

### `usm networks list <site-id>`
List all configured networks for a site.

```bash
usm networks list 60abcdef1234567890abcdef
```

## Configuration

Configuration file: `~/.config/usm/config.yaml`

**Note**: API keys are NOT stored in the config file. Use environment variables or flags.

```yaml
api:
  base_url: https://api.ui.com
  timeout: 30

output:
  format: table      # table, json
  color: auto        # auto, always, never
  no_headers: false
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `USM_API_KEY` | API key for authentication (required) |
| `USM_BASE_URL` | API base URL (default: `https://api.ui.com`) |
| `USM_TIMEOUT` | Request timeout in seconds (default: `30`) |
| `USM_FORMAT` | Default output format: `table`, `json` |
| `USM_COLOR` | Color mode: `auto`, `always`, `never` |
| `USM_NO_HEADERS` | Disable table headers: `true`, `false` |
| `USM_CONFIG` | Path to config file |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication failure |
| 3 | Permission denied |
| 4 | Validation error |
| 5 | Rate limited (429) |
| 6 | Network error |

## Development

```bash
# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run linter
make lint

# Format code
make format

# Install git hooks
make install-hooks
```

## API Reference

- [Site Manager API v1.0.0 Documentation](https://developer.ui.com/site-manager/v1.0.0/gettingstarted)

## Related Projects

- `unifi` CLI - For local UniFi Network and Protect APIs

## Beta Testing & Local Controller Support (v0.0.2)

This release includes experimental support for direct connection to UniFi OS local controllers (UDM, UDM-Pro, UDR). This feature is in beta and requires community testing.

### Local Controller Mode

Connect directly to your UniFi Dream Machine or UniFi OS console without using the cloud API:

```bash
# List devices from local controller
usm --local --host=192.168.1.1 --username=admin --password=yourpassword devices list

# Or use environment variables
export USM_LOCAL=true
export USM_HOST=192.168.1.1
export USM_USERNAME=admin
export USM_PASSWORD=yourpassword

usm devices list
usm clients list
```

**Security Note**: Never pass `--password` as a CLI flag in production - use the `USM_PASSWORD` environment variable to prevent credentials from appearing in process lists.

### Debug Mode

If local controller commands fail, enable debug mode to see the exact API requests and responses:

```bash
usm --local --host=192.168.1.1 --username=admin --debug devices list
```

Debug output is automatically sanitized - passwords, CSRF tokens, cookies, and API keys are redacted.

### Reporting Issues

When reporting local controller issues:

1. Run the command with `--debug` flag
2. Copy the `[DEBUG]` output (credentials are already redacted)
3. Open a GitHub issue with the debug output

### Chrome Dev Tools Debugging

If you encounter issues with local controller commands, you can inspect the exact API payloads:

1. Open your UniFi controller web UI in Chrome
2. Open Developer Tools (F12) → Network tab
3. Perform the action in the web UI (e.g., create a WLAN)
4. Find the request in the Network tab
5. Check the **Payload** tab to see exact JSON structure
6. Compare with the `[DEBUG] Raw Payload` output from the CLI

**Note**: Local controller API endpoints differ from Cloud API:
- Cloud: `/v1/sites/{id}/devices`
- Local: `/proxy/network/api/s/{site}/stat/device`

### Implemented Local Controller Features

✅ **Working**:
- Site listing (`sites list`)
- Device listing and details (`devices list`, `devices get`)
- Client listing (`clients list`)
- WLAN CRUD operations (`wlans list`, `create`, `update`, `delete`)
- Device restart (`devices restart`)

🚧 **Stubbed** (not yet implemented):
- Site creation/update/delete
- Host/Console management
- Device adoption and firmware upgrades
- Client blocking/unblocking
- Alerts and events
- Network configuration

## License

MIT
