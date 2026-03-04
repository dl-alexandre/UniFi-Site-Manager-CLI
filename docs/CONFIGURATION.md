# Configuration Reference

Complete guide to configuring UniFi Site Manager CLI.

## Table of Contents

- [Configuration Overview](#configuration-overview)
- [Configuration File](#configuration-file)
- [Environment Variables](#environment-variables)
- [Command-Line Flags](#command-line-flags)
- [Cloud Mode Configuration](#cloud-mode-configuration)
- [Local Mode Configuration](#local-mode-configuration)
- [Output Configuration](#output-configuration)
- [Security Best Practices](#security-best-practices)
- [Configuration Examples](#configuration-examples)
- [Troubleshooting](#troubleshooting)

## Configuration Overview

The CLI supports three configuration methods (in order of precedence):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (lowest priority)

## Configuration File

### Location

| Platform | Path |
|----------|------|
| macOS | `~/.config/usm/config.yaml` |
| Linux | `~/.config/usm/config.yaml` or `$XDG_CONFIG_HOME/usm/config.yaml` |
| Windows | `%APPDATA%\usm\config.yaml` |

### File Format

```yaml
# API Configuration
api:
  base_url: https://api.ui.com
  timeout: 30

# Output Configuration
output:
  format: table      # table, json
  color: auto        # auto, always, never
  no_headers: false

# Logging
logging:
  level: info        # debug, info, warn, error
  file: ""           # Path to log file (empty = stdout)
```

### Creating Configuration

```bash
# Interactive setup (recommended)
usm init

# Manual creation
mkdir -p ~/.config/usm
cat > ~/.config/usm/config.yaml << 'EOF'
api:
  base_url: https://api.ui.com
  timeout: 30

output:
  format: table
  color: auto
  no_headers: false
EOF
```

## Environment Variables

### Cloud Mode Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `USM_API_KEY` | Site Manager API key | - | Yes (Cloud) |
| `USM_BASE_URL` | API base URL | `https://api.ui.com` | No |
| `USM_TIMEOUT` | Request timeout (seconds) | `30` | No |
| `USM_FORMAT` | Default output format | `table` | No |
| `USM_COLOR` | Color output mode | `auto` | No |
| `USM_NO_HEADERS` | Disable table headers | `false` | No |
| `USM_CONFIG` | Custom config file path | - | No |
| `USM_VERBOSE` | Enable verbose output | `false` | No |
| `USM_DEBUG` | Enable debug mode | `false` | No |

### Local Mode Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `USM_LOCAL` | Enable local controller mode | Yes (Local) |
| `USM_HOST` | Controller IP/hostname | Yes (Local) |
| `USM_USERNAME` | Controller username | Yes (Local) |
| `USM_PASSWORD` | Controller password | Yes (Local) |
| `USM_INSECURE` | Skip TLS verification | No |
| `USM_PORT` | Controller port | No |

### Example Environment Setup

```bash
# Cloud mode
export USM_API_KEY="your-api-key-here"
export USM_TIMEOUT=60
export USM_FORMAT="json"

# Local mode
export USM_LOCAL=true
export USM_HOST="192.168.1.1"
export USM_USERNAME="admin"
export USM_PASSWORD="your-password"
export USM_INSECURE=true
```

### Persistent Environment Variables

**macOS/Linux (bash/zsh)**:
```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.bash_profile
export USM_API_KEY="your-api-key"
```

**macOS (launchd)**:
```bash
# Create ~/Library/LaunchAgents/usm.env.plist
launchctl setenv USM_API_KEY "your-api-key"
```

**Windows (PowerShell)**:
```powershell
# User scope
[Environment]::SetEnvironmentVariable("USM_API_KEY", "your-api-key", "User")

# Machine scope (requires admin)
[Environment]::SetEnvironmentVariable("USM_API_KEY", "your-api-key", "Machine")
```

**Windows (GUI)**:
1. System Properties → Advanced → Environment Variables
2. New → Variable name: `USM_API_KEY` → Variable value: `your-api-key`
3. OK → Restart terminal

## Command-Line Flags

### Global Flags

```
--api-key string        API key for authentication ($USM_API_KEY)
--base-url string       API base URL ($USM_BASE_URL)
--color string          Color mode: auto, always, never ($USM_COLOR)
--config string         Config file path ($USM_CONFIG)
--debug                 Enable debug output ($USM_DEBUG)
--format string         Output format: table, json ($USM_FORMAT)
--help                  Show help
--host string           Local controller host ($USM_HOST)
--insecure-skip-verify  Skip TLS certificate verification
--local                 Enable local controller mode ($USM_LOCAL)
--no-headers            Disable table headers ($USM_NO_HEADERS)
--password string       Local controller password ($USM_PASSWORD)
--timeout int           Request timeout in seconds ($USM_TIMEOUT)
--username string       Local controller username ($USM_USERNAME)
--verbose, -v           Enable verbose output ($USM_VERBOSE)
--version               Show version
```

### Command-Specific Flags

#### Sites Commands

```
usm sites list:
  --page-size int    Number of results per page
  --search string    Filter by name/description

usm sites create:
  --description string   Site description

usm sites update:
  --description string   New description
  --name string          New name

usm sites delete:
  --force    Skip confirmation prompt

usm sites stats:
  --period string    Statistics period: day, week, month
```

#### Devices Commands

```
usm devices list:
  --page-size int    Number of results per page
  --status string    Filter by status: online, offline, pending
  --type string      Filter by type: ap, switch, gateway

usm devices restart, upgrade:
  --force    Skip confirmation prompt
```

#### Clients Commands

```
usm clients list:
  --page-size int      Number of results per page
  --search string      Filter by name/MAC/IP
  --wired-only         Show only wired clients
  --wireless-only      Show only wireless clients
```

#### WLANs Commands

```
usm wlans create:
  --band string        Radio band: both, 2.4, 5
  --enabled            Enable the WLAN (default true)
  --guest              Mark as guest network
  --hide-ssid          Hide SSID from broadcast
  --password string    WPA password
  --security string    Security type: open, wpapsk, wpa2, wpa3
  --vlan int           VLAN ID
  --wpa3               Enable WPA3

usm wlans update:
  --band string        Radio band
  --enabled bool       Enable/disable WLAN
  --password string    New password
  --security string    Security type
  --vlan int           VLAN ID
```

## Cloud Mode Configuration

### Minimal Setup

```bash
# Using environment variable
export USM_API_KEY="your-api-key"
usm whoami
```

### With Custom Settings

```yaml
# ~/.config/usm/config.yaml
api:
  base_url: https://api.ui.com
  timeout: 45  # Longer timeout for large sites

output:
  format: json
  color: never  # Disable colors for scripting
```

### Enterprise Setup

```bash
# ~/.usmrc (sourced by scripts)
export USM_API_KEY="${USM_API_KEY:?API key not set}"
export USM_BASE_URL="${USM_BASE_URL:-https://api.ui.com}"
export USM_TIMEOUT="${USM_TIMEOUT:-30}"
export USM_FORMAT="${USM_FORMAT:-table}"

# Verify configuration
usm whoami
```

## Local Mode Configuration

### UniFi OS (UDM/UDR/Cloud Key)

```bash
# Environment setup
export USM_LOCAL=true
export USM_HOST="192.168.1.1"
export USM_USERNAME="admin"
export USM_PASSWORD="your-password"
export USM_INSECURE=true  # Self-signed certs

# Test connection
usm whoami
```

### Standalone Controller

```bash
export USM_LOCAL=true
export USM_HOST="192.168.1.100"
export USM_PORT=8443
export USM_USERNAME="admin"
export USM_PASSWORD="your-password"
export USM_INSECURE=true

usm sites list
```

### Using Command Flags

```bash
# All-in-one command
usm \
  --local \
  --host=192.168.1.1 \
  --username=admin \
  --password=secret \
  --insecure-skip-verify \
  sites list
```

### Configuration File for Local Mode

```yaml
# ~/.config/usm/config-local.yaml
api:
  base_url: https://192.168.1.1
  timeout: 30

local:
  mode: true
  insecure: true
```

Usage:
```bash
usm --config ~/.config/usm/config-local.yaml sites list
```

## Output Configuration

### Format Options

```yaml
# config.yaml
output:
  format: table    # Human-readable tables
  # format: json   # Machine-readable JSON
```

### Color Modes

```yaml
output:
  color: auto      # Enable if terminal supports it
  # color: always  # Always enable colors
  # color: never   # Disable colors (for piping)
```

### Table Headers

```yaml
output:
  no_headers: false  # Show headers (default)
  # no_headers: true # Hide headers (for parsing)
```

### Example: Script-Friendly Output

```bash
# Disable colors and headers for easy parsing
usm sites list --color never --no-headers

# Or via environment
USM_COLOR=never USM_NO_HEADERS=true usm sites list

# JSON for structured data
usm sites list --output json | jq -r '.sites[].id'
```

## Security Best Practices

### 1. Never Commit API Keys

```bash
# .gitignore
config.yaml
.env
*.key
```

### 2. Use Environment Variables

```bash
# Good
export USM_API_KEY="your-key"
usm sites list

# Bad - never do this
usm --api-key="your-key" sites list  # Visible in process list!
```

### 3. Restrict File Permissions

```bash
chmod 600 ~/.config/usm/config.yaml
```

### 4. Rotate API Keys Regularly

Set a reminder to rotate keys every 90 days:
```bash
# Add to calendar: "Rotate USM API Key"
# 1. Generate new key at unifi.ui.com
# 2. Update environment variable
# 3. Revoke old key
```

### 5. Use Dedicated API Keys

Create separate keys for:
- Development
- Production
- CI/CD automation
- Monitoring scripts

### 6. Secure Local Controller Passwords

```bash
# Use password manager or secrets service
export USM_PASSWORD="$(security find-generic-password -s usm-password -w)"
```

## Configuration Examples

### Example 1: Personal Use

```yaml
# ~/.config/usm/config.yaml
api:
  timeout: 30

output:
  format: table
  color: auto
```

Environment:
```bash
# ~/.bashrc
export USM_API_KEY="ui_api_xxxxxxxxxxxx"
```

### Example 2: Production Environment

```yaml
# /etc/usm/config.yaml
api:
  base_url: https://api.ui.com
  timeout: 60

output:
  format: json
  color: never
  no_headers: true
```

Environment (systemd service):
```ini
# /etc/systemd/system/usm-monitor.service
[Service]
Environment="USM_API_KEY=ui_api_xxxxxxxxxxxx"
Environment="USM_CONFIG=/etc/usm/config.yaml"
```

### Example 3: CI/CD Pipeline

```yaml
# .github/workflows/monitor.yml
env:
  USM_API_KEY: ${{ secrets.USM_API_KEY }}
  USM_FORMAT: json
  USM_NO_HEADERS: true

jobs:
  monitor:
    steps:
      - run: usm sites health $SITE_ID
```

### Example 4: Multiple Sites

```bash
#!/bin/bash
# multi-site.sh

# Site A (Cloud API)
export USM_API_KEY="key-for-site-a"
usm sites list

# Site B (Local UDM)
export USM_LOCAL=true
export USM_HOST="192.168.2.1"
export USM_PASSWORD="password-b"
usm sites list
```

### Example 5: Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  usm:
    image: ghcr.io/dl-alexandre/usm:latest
    environment:
      - USM_API_KEY=${USM_API_KEY}
      - USM_FORMAT=json
    volumes:
      - ./config:/root/.config/usm:ro
    command: sites list
    networks:
      - usm-network
```

## Troubleshooting

### Configuration Not Loading

```bash
# Check config file location
usm --help | grep -i config

# Verify file exists and is readable
ls -la ~/.config/usm/config.yaml

# Test with explicit path
usm --config ~/.config/usm/config.yaml sites list
```

### Environment Variables Not Working

```bash
# Verify variable is set
echo $USM_API_KEY

# Check for typos
env | grep USM

# Try with explicit export
export USM_API_KEY="your-key"
usm whoami
```

### Permission Denied

```bash
# Fix config file permissions
chmod 600 ~/.config/usm/config.yaml

# Check directory permissions
ls -la ~/.config/
```

### API Key Not Found

```bash
# Error: "API key is required"

# Solution 1: Set environment variable
export USM_API_KEY="your-key"

# Solution 2: Use command flag
usm --api-key="your-key" sites list

# Solution 3: Run init
usm init
```

## Configuration Reference Table

| Setting | Config File | Environment | Flag | Default |
|---------|------------|-------------|------|---------|
| API Key | N/A (security) | `USM_API_KEY` | `--api-key` | Required |
| Base URL | `api.base_url` | `USM_BASE_URL` | `--base-url` | `https://api.ui.com` |
| Timeout | `api.timeout` | `USM_TIMEOUT` | `--timeout` | `30` |
| Format | `output.format` | `USM_FORMAT` | `--format` | `table` |
| Color | `output.color` | `USM_COLOR` | `--color` | `auto` |
| No Headers | `output.no_headers` | `USM_NO_HEADERS` | `--no-headers` | `false` |
| Verbose | N/A | `USM_VERBOSE` | `--verbose, -v` | `false` |
| Debug | N/A | `USM_DEBUG` | `--debug` | `false` |
| Local Mode | N/A | `USM_LOCAL` | `--local` | `false` |
| Host | N/A | `USM_HOST` | `--host` | - |
| Username | N/A | `USM_USERNAME` | `--username` | - |
| Password | N/A | `USM_PASSWORD` | `--password` | - |
| Insecure | N/A | `USM_INSECURE` | `--insecure-skip-verify` | `false` |
| Config File | N/A | `USM_CONFIG` | `--config` | Auto-detected |

## Next Steps

- See [Installation Guide](INSTALL.md) for setup instructions
- Read [Usage Guide](USAGE.md) for command examples
- Check [API Documentation](API.md) for endpoint details
