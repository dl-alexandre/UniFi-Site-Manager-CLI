# usm - UniFi Site Manager CLI

A command-line interface for the [UniFi Site Manager API](https://developer.ui.com/site-manager/v1.0.0/gettingstarted).

## Overview

`usm` provides a simple, fast interface for managing UniFi sites via the official Site Manager API at `api.ui.com`. This tool focuses on cloud-based site management - for local controller APIs (Network, Protect), see the separate `unifi` CLI.

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/dl-alexandre/usm/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/dl-alexandre/usm/releases/latest/download/usm-darwin-arm64 -o usm

# macOS (Intel)
curl -L https://github.com/dl-alexandre/usm/releases/latest/download/usm-darwin-amd64 -o usm

# Linux
# amd64 or arm64

# Windows (PowerShell)
# amd64

chmod +x usm
sudo mv usm /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/dl-alexandre/UniFi-Site-Manager-CLI.git
cd usm
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

## Commands

### `usm init`
Interactive configuration setup. Creates `~/.config/usm/config.yaml`.

```bash
usm init          # First-time setup
usm init --force # Overwrite existing config
```

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

### `usm whoami`
Show authenticated user account information.

```bash
usm whoami
usm whoami --output json
```

### `usm version`
Show version information and check for updates.

```bash
usm version              # Show current version
usm version --check      # Check GitHub for latest release
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

## License

MIT
