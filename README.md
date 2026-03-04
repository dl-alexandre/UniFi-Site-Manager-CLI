# UniFi Site Manager CLI

[![CI](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/actions/workflows/ci.yml/badge.svg)](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/UniFi-Site-Manager-CLI)](https://goreportcard.com/report/github.com/dl-alexandre/UniFi-Site-Manager-CLI)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/dl-alexandre/UniFi-Site-Manager-CLI)](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases)
[![Homebrew](https://img.shields.io/badge/homebrew-tap-blue)](https://github.com/dl-alexandre/homebrew-tap)

> A full-featured command-line interface for managing UniFi networks via the Site Manager Cloud API and local UniFi OS controllers.

## Features

- ✨ **Dual Mode Support**: Cloud API (Site Manager) and Local Controller (UDM/UDR/Cloud Key)
- 🚀 **Complete Site Management**: Create, update, delete sites with full API coverage
- 🔒 **Secure Authentication**: API key support with environment variable safety
- 📊 **Rich Output Formats**: Tables, JSON, and CSV for easy scripting
- 🏠 **Device Management**: List, restart, upgrade firmware, adopt devices
- 👥 **Client Management**: View, block/unblock connected clients
- 📡 **WLAN Management**: Create, configure, and optimize wireless networks
- 🔔 **Alert Monitoring**: Acknowledge and archive system alerts
- 🛠️ **UniFi OS Support**: Direct connection to UDM, UDM-Pro, UDR, Cloud Key

## Demo

```bash
$ usm whoami
┌─────────────┬─────────────────────────────────────────┐
│ EMAIL       │ admin@example.com                       │
│ NAME        │ Network Admin                           │
│ ROLE        │ Owner                                   │
│ SITES       │ 3                                       │
│ HOSTS       │ 2                                       │
└─────────────┴─────────────────────────────────────────┘

$ usm sites list
┌─────────────────────────┬─────────────────┬──────────┬──────────┐
│ ID                      │ NAME            │ HOSTS    │ DEVICES  │
├─────────────────────────┼─────────────────┼──────────┼──────────┤
│ 60abcdef1234567890abcde │ Main Office     │ 1        │ 12       │
│ 60abcdef1234567890abcdf │ Home Network    │ 1        │ 8        │
│ 60abcdef1234567890abcd0 │ Branch Office   │ 1        │ 6        │
└─────────────────────────┴─────────────────┴──────────┴──────────┘
```

## Quick Start

### 1. Installation

```bash
# macOS (Homebrew)
brew install dl-alexandre/tap/usm

# Linux (curl)
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
chmod +x usm
sudo mv usm /usr/local/bin/

# Or download from GitHub Releases
# https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases
```

### 2. Configure

```bash
# Interactive setup (recommended)
usm init

# Or use environment variables
export USM_API_KEY="your-api-key-from-unifi.ui.com"
```

### 3. Verify & Use

```bash
# Verify connection
usm whoami

# List sites
usm sites list

# List devices in a site
usm devices list <site-id>

# Get site health
usm sites health <site-id>
```

## Installation

### macOS

```bash
# Using Homebrew
brew tap dl-alexandre/tap
brew install usm

# Or download binary
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-darwin-arm64 -o usm
chmod +x usm
sudo mv usm /usr/local/bin/
```

### Linux

```bash
# Download latest release
wget https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64.tar.gz
tar -xzf usm-linux-amd64.tar.gz
sudo mv usm /usr/local/bin/

# Or using snap (coming soon)
snap install usm
```

### Windows

```powershell
# Using scoop
scoop bucket add unifi https://github.com/dl-alexandre/scoop-bucket
scoop install usm

# Or download from releases and add to PATH
# https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases
```

### Docker

```bash
# Pull the image
docker pull ghcr.io/dl-alexandre/usm:latest

# Run with environment variables
docker run --rm -e USM_API_KEY="your-key" ghcr.io/dl-alexandre/usm sites list

# With volume for config
docker run --rm -v ~/.config/usm:/root/.config/usm ghcr.io/dl-alexandre/usm sites list
```

### Build from Source

```bash
# Requirements: Go 1.24+
git clone https://github.com/dl-alexandre/UniFi-Site-Manager-CLI.git
cd UniFi-Site-Manager-CLI
make build

# Or use go install
go install github.com/dl-alexandre/UniFi-Site-Manager-CLI/cmd/usm@latest
```

## Usage Examples

### Basic Operations

```bash
# Initialize configuration
usm init

# Show current user
usm whoami

# List all sites
usm sites list

# Get site details
usm sites get 60abcdef1234567890abcdef

# Check site health
usm sites health 60abcdef1234567890abcdef
```

### Advanced Usage

```bash
# List devices with filtering
usm devices list 60abcdef1234567890abcdef --status online --type ap

# Create a new WLAN
usm wlans create 60abcdef1234567890abcdef "Guest WiFi" "Guest-Network" \
  --password "guest123" --security wpapsk --vlan 20

# Block a problematic client
usm clients block 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"

# Export data as JSON
usm sites list --output json > sites.json
```

### Local Controller Mode (UniFi OS)

```bash
# Connect to UDM/UDM-Pro directly
export USM_HOST="192.168.1.1"
export USM_USERNAME="admin"
export USM_PASSWORD="yourpassword"
export USM_LOCAL=true

# List devices from local controller
usm devices list

# Or use flags
usm --local --host=192.168.1.1 --username=admin --password=secret devices list
```

### Automation Scripts

```bash
# Health check script
#!/bin/bash
SITE_ID="60abcdef1234567890abcdef"
HEALTH=$(usm sites health $SITE_ID --output json)
DEVICE_ISSUES=$(echo $HEALTH | jq '.devices | map(select(.status != "online")) | length')

if [ "$DEVICE_ISSUES" -gt 0 ]; then
  echo "Warning: $DEVICE_ISSUES devices are not online"
  exit 1
fi
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `USM_API_KEY` | Site Manager API key | Cloud mode |
| `USM_BASE_URL` | API base URL (default: https://api.ui.com) | No |
| `USM_TIMEOUT` | Request timeout in seconds (default: 30) | No |
| `USM_FORMAT` | Default output: `table`, `json` | No |
| `USM_COLOR` | Color mode: `auto`, `always`, `never` | No |
| `USM_NO_HEADERS` | Disable table headers: `true`, `false` | No |
| `USM_LOCAL` | Enable local controller mode: `true` | Local mode |
| `USM_HOST` | Local controller IP/hostname | Local mode |
| `USM_USERNAME` | Local controller username | Local mode |
| `USM_PASSWORD` | Local controller password | Local mode |

### Configuration File

Location: `~/.config/usm/config.yaml`

```yaml
api:
  base_url: https://api.ui.com
  timeout: 30

output:
  format: table      # table, json
  color: auto        # auto, always, never
  no_headers: false
```

### Getting an API Key

1. Log into [unifi.ui.com](https://unifi.ui.com)
2. Go to **Settings** → **Control Plane** → **Integrations** → **API Keys**
3. Click **Create API Key**
4. Copy the key (it won't be shown again)

## API Coverage

| Feature | Cloud API | Local API | Endpoint |
|---------|-----------|-----------|----------|
| **Sites** ||||
| List sites | ✅ | ✅ | GET /v1/sites |
| Get site | ✅ | ✅ | GET /v1/sites/{id} |
| Create site | ✅ | ❌ | POST /v1/sites |
| Update site | ✅ | ❌ | PUT /v1/sites/{id} |
| Delete site | ✅ | ❌ | DELETE /v1/sites/{id} |
| Site health | ✅ | ✅ | GET /v1/sites/{id}/health |
| Site stats | ✅ | ✅ | GET /v1/sites/{id}/stats |
| **Hosts** ||||
| List hosts | ✅ | ❌ | GET /v1/hosts |
| Get host | ✅ | ❌ | GET /v1/hosts/{id} |
| Restart host | ✅ | ❌ | POST /v1/hosts/{id}/restart |
| Host health | ✅ | ❌ | GET /v1/hosts/{id}/health |
| Host stats | ✅ | ❌ | GET /v1/hosts/{id}/stats |
| **Devices** ||||
| List devices | ✅ | ✅ | GET /v1/sites/{id}/devices |
| Get device | ✅ | ✅ | GET /v1/sites/{id}/devices/{id} |
| Restart device | ✅ | ✅ | POST /v1/sites/{id}/devices/{id}/restart |
| Upgrade firmware | ✅ | ❌ | POST /v1/sites/{id}/devices/{id}/upgrade |
| Adopt device | ✅ | ❌ | POST /v1/sites/{id}/devices/adopt |
| **Clients** ||||
| List clients | ✅ | ✅ | GET /v1/sites/{id}/clients |
| Get client | ✅ | ✅ | GET /v1/sites/{id}/clients/{mac} |
| Block client | ✅ | ✅ | POST /v1/sites/{id}/clients/{mac}/block |
| Unblock client | ✅ | ✅ | POST /v1/sites/{id}/clients/{mac}/unblock |
| Client stats | ✅ | ✅ | GET /v1/sites/{id}/clients/{mac}/stats |
| **WLANs** ||||
| List WLANs | ✅ | ✅ | GET /v1/sites/{id}/wlans |
| Get WLAN | ✅ | ✅ | GET /v1/sites/{id}/wlans/{id} |
| Create WLAN | ✅ | ✅ | POST /v1/sites/{id}/wlans |
| Update WLAN | ✅ | ✅ | PUT /v1/sites/{id}/wlans/{id} |
| Delete WLAN | ✅ | ✅ | DELETE /v1/sites/{id}/wlans/{id} |
| **Alerts** ||||
| List alerts | ✅ | ❌ | GET /v1/alerts |
| Acknowledge | ✅ | ❌ | POST /v1/sites/{id}/alerts/{id}/ack |
| Archive | ✅ | ❌ | POST /v1/sites/{id}/alerts/{id}/archive |
| **Events** ||||
| List events | ✅ | ❌ | GET /v1/events |
| **Networks** ||||
| List networks | ✅ | ✅ | GET /v1/sites/{id}/networks |

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   sites     │  │  devices    │  │    clients, wlans      │  │
│  │   hosts     │  │  alerts     │  │    networks, events    │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
└─────────┼────────────────┼─────────────────────┼────────────────┘
          │                │                     │
          ▼                ▼                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Client Interface                       │
│              (SiteManager interface - unified API)              │
└─────────────────────────────────────────────────────────────────┘
          │                                    │
          ▼                                    ▼
┌──────────────────────┐          ┌──────────────────────────────┐
│   Cloud API Client   │          │   Local API Client           │
│   (api.ui.com)       │          │   (UDM/UDR/Cloud Key)        │
│                      │          │                              │
│ • API Key Auth       │          │ • Username/Password Auth     │
│ • Rate limiting      │          │ • CSRF token handling        │
│ • Retry logic        │          │ • Self-signed cert support   │
│ • Error mapping      │          │ • Proxy path handling        │
└──────────────────────┘          └──────────────────────────────┘
```

## Troubleshooting

### Authentication Errors

**Problem**: `Error: API key is required`

**Solution**:
```bash
# Verify API key is set
usm init

# Or set environment variable
export USM_API_KEY="your-api-key"
echo $USM_API_KEY  # Should show your key
```

### Connection Timeouts

**Problem**: Network errors when connecting to local controller

**Solution**:
```bash
# Increase timeout
usm --timeout 60 sites list

# For UniFi OS with self-signed certs
usm --local --host=192.168.1.1 --insecure-skip-verify devices list
```

### Rate Limiting (429)

**Problem**: `Error: rate limited`

**Solution**: The CLI automatically retries with exponential backoff. For heavy automation, add delays between requests.

### Empty Results in Local Mode

**Problem**: Commands return empty results on UDM/UDR

**Solution**:
1. Enable debug mode: `usm --local --debug sites list`
2. Verify UniFi OS version is up to date
3. Check credentials have admin access
4. See [Beta Testing](#beta-testing) section

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

## Documentation

- [Installation Guide](docs/INSTALL.md) - Detailed installation instructions
- [Usage Documentation](docs/USAGE.md) - Comprehensive command reference
- [API Documentation](docs/API.md) - API endpoint details
- [Configuration Reference](docs/CONFIGURATION.md) - All configuration options
- [FAQ](docs/FAQ.md) - Frequently asked questions
- [Development Guide](docs/DEVELOPMENT.md) - Contributing and development setup
- [Changelog](docs/CHANGELOG.md) - Version history
- [Security](docs/SECURITY.md) - Security best practices

## Examples

See the [examples/](examples/) directory for:
- [Basic usage examples](examples/basic/)
- [Automation scripts](examples/automation/)
- [Monitoring setup](examples/monitoring/)
- [CI/CD integration](examples/ci-cd/)
- [Home Assistant integration](examples/home-assistant/)

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

- [Report bugs](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/issues)
- [Request features](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/issues)
- [Submit pull requests](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/pulls)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Kong](https://github.com/alecthomas/kong) for CLI parsing
- HTTP client powered by [Resty](https://github.com/go-resty/resty)
- Configuration management with [Viper](https://github.com/spf13/viper)

## Support

- 📧 GitHub Issues: [github.com/dl-alexandre/UniFi-Site-Manager-CLI/issues](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/discussions)

---

**Note**: This project is not affiliated with or endorsed by Ubiquiti Inc. UniFi is a trademark of Ubiquiti Inc.
