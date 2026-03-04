# Frequently Asked Questions

Common questions and answers about UniFi Site Manager CLI.

## Table of Contents

- [General Questions](#general-questions)
- [Installation](#installation)
- [Authentication](#authentication)
- [Cloud API](#cloud-api)
- [Local Controller](#local-controller)
- [Commands and Usage](#commands-and-usage)
- [Troubleshooting](#troubleshooting)
- [Automation and Scripting](#automation-and-scripting)
- [Performance and Limits](#performance-and-limits)

## General Questions

### What is the difference between Cloud API and Local Controller mode?

**Cloud API (default)**:
- Connects to Ubiquiti's Site Manager at `api.ui.com`
- Manages multiple sites across different hosts
- Requires API key from unifi.ui.com
- Full feature set

**Local Controller**:
- Connects directly to your UDM/UDR/Cloud Key
- Manages single site only
- Uses username/password authentication
- Some features unavailable (see [API Coverage](../README.md#api-coverage))

### Is the CLI officially supported by Ubiquiti?

No, this is a community-maintained tool. It's built using the official UniFi Site Manager API but is not affiliated with or endorsed by Ubiquiti Inc.

### What are the system requirements?

- **OS**: macOS 10.15+, Linux, Windows 10+
- **Go**: 1.24+ (if building from source)
- **Network**: Internet for Cloud API mode
- **Disk**: 20 MB
- **RAM**: 50 MB

### Where can I get help?

- GitHub Issues: [github.com/dl-alexandre/UniFi-Site-Manager-CLI/issues](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/issues)
- GitHub Discussions: [github.com/dl-alexandre/UniFi-Site-Manager-CLI/discussions](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/discussions)

## Installation

### How do I install on macOS?

```bash
# Homebrew (recommended)
brew tap dl-alexandre/tap
brew install usm

# Or download directly
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-darwin-arm64 -o usm
chmod +x usm
sudo mv usm /usr/local/bin/
```

### How do I install on Linux?

```bash
# Download binary
wget https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -O usm
chmod +x usm
sudo mv usm /usr/local/bin/
```

### How do I update to the latest version?

**Homebrew**:
```bash
brew upgrade usm
```

**Manual**:
Download the latest release and replace the binary.

### Can I install via Go?

```bash
go install github.com/dl-alexandre/UniFi-Site-Manager-CLI/cmd/usm@latest
```

## Authentication

### Where do I get an API key?

1. Log into [unifi.ui.com](https://unifi.ui.com)
2. Go to **Settings** → **Control Plane** → **Integrations** → **API Keys**
3. Click **Create API Key**
4. Copy the key (shown only once!)

### Can I use my Ubiquiti SSO account?

No, Cloud API requires an API key. Local Controller mode requires a **local admin account** on your UniFi controller (not an SSO/cloud account).

### Is it safe to store the API key in environment variables?

Yes, environment variables are safer than command-line flags because:
- They're not visible in process lists (`ps aux`)
- They're not stored in shell history
- They can be restricted to specific processes

However, for maximum security:
- Use a secrets manager (1Password, Vault, etc.)
- Rotate keys regularly
- Use different keys for different environments

### How do I handle credentials securely in scripts?

```bash
#!/bin/bash
# Use a password manager or secrets service
export USM_API_KEY="$(secret-tool lookup service usm)"
export USM_PASSWORD="$(security find-generic-password -s usm-password -w)"

usm sites list
```

### Can I use a .env file?

Yes, but be careful:
```bash
# .env file
USM_API_KEY=your-key

# Load it
export $(cat .env | xargs)
usm sites list

# Add .env to .gitignore!
```

## Cloud API

### What features are available in Cloud API mode?

All features:
- Site management (CRUD)
- Host management
- Device management
- Client management
- WLAN management
- Alerts and events
- Networks

### What are the rate limits?

- 100 requests per minute per API key
- 20 requests per second burst
- The CLI handles rate limiting automatically with exponential backoff

### Can I manage multiple sites?

Yes, the Cloud API supports managing all sites associated with your account:
```bash
usm sites list  # Shows all your sites
```

### How do I switch between sites?

```bash
# List sites to get IDs
usm sites list

# Use site ID in commands
usm devices list 60abcdef1234567890abcdef
usm clients list 60abcdef1234567890abcdef
```

## Local Controller

### Which devices support Local Controller mode?

- UniFi Dream Machine (UDM)
- UniFi Dream Machine Pro (UDM-Pro)
- UniFi Dream Router (UDR)
- UniFi Dream Machine SE (UDM-SE)
- UniFi Cloud Key Gen2+
- Standalone UniFi Network Controller

### What features work with Local Controller?

**Working**:
- Sites, Devices, Clients, WLANs (list/get)
- Device restart
- Client block/unblock

**Not Available**:
- Site CRUD operations
- Host management
- Firmware upgrades
- Device adoption
- Alerts/Events
- Network management

### Why can't I connect to my UDM?

Common issues:
1. **Wrong credentials**: Use local account, not SSO
2. **TLS certificate**: Use `--insecure-skip-verify` for self-signed certs
3. **UniFi OS mode**: Ensure correct mode for UDM vs standalone
4. **Network connectivity**: Verify firewall rules

### How do I connect to a UDM Pro?

```bash
export USM_LOCAL=true
export USM_HOST="192.168.1.1"
export USM_USERNAME="admin"
export USM_PASSWORD="your-password"
export USM_INSECURE=true  # Self-signed cert

usm sites list
```

### Can I use the CLI with a Cloud Key?

Yes:
```bash
export USM_LOCAL=true
export USM_HOST="192.168.1.10"
export USM_PORT=443
export USM_USERNAME="admin"
export USM_PASSWORD="password"
export USM_INSECURE=true

usm sites list
```

## Commands and Usage

### How do I list all my sites?

```bash
usm sites list

# With pagination
usm sites list --page-size 100

# Get all (no pagination)
usm sites list --page-size 0
```

### How do I get device details?

```bash
# List devices first
usm devices list 60abcdef1234567890abcdef

# Get specific device
usm devices get 60abcdef1234567890abcdef 60fedcba0987654321fedcba

# JSON output for scripting
usm devices get 60abcdef1234567890abcdef 60fedcba0987654321fedcba --output json
```

### How do I create a new WLAN?

```bash
usm wlans create 60abcdef1234567890abcdef "Guest WiFi" "Guest-Network" \
  --password "guest123" \
  --security wpapsk \
  --vlan 20
```

### How do I block a client?

```bash
usm clients block 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

### How do I restart a device?

```bash
# With confirmation
usm devices restart 60abcdef1234567890abcdef 60fedcba0987654321fedcba

# Without confirmation
usm devices restart 60abcdef1234567890abcdef 60fedcba0987654321fedcba --force
```

### How do I export data?

```bash
# Export sites as JSON
usm sites list --output json > sites.json

# Export devices as CSV (via jq)
usm devices list SITE_ID --output json | \
  jq -r '.devices[] | [.id, .name, .mac, .status] | @csv' > devices.csv
```

## Troubleshooting

### "API key is required" error

```bash
# Verify API key is set
echo $USM_API_KEY

# Set it if missing
export USM_API_KEY="your-key"

# Or use flag
usm --api-key="your-key" sites list
```

### "Permission denied" error

Your API key or account doesn't have permission for the operation:
- Verify you're the site owner or have appropriate role
- Check if trying to modify a read-only resource
- For local mode, ensure account has admin privileges

### "Rate limited" error

The API is throttling your requests:
- The CLI automatically retries with backoff
- Add delays in scripts: `sleep 1` between requests
- Consider reducing request frequency

### Empty results on local controller

This is common in beta:
1. Enable debug mode: `usm --local --debug sites list`
2. Check UniFi OS version is up to date
3. Verify credentials and permissions
4. Report issue with debug output

### Connection timeout

```bash
# Increase timeout
usm --timeout 60 sites list

# Check network connectivity
ping api.ui.com  # Cloud mode
ping 192.168.1.1  # Local mode
```

### "Command not found" after installation

```bash
# Check if in PATH
which usm

# If not found, add to PATH
export PATH=$PATH:/usr/local/bin

# Or specify full path
/usr/local/bin/usm sites list
```

## Automation and Scripting

### How do I use the CLI in a cron job?

```bash
# crontab entry
0 * * * * /usr/local/bin/usm --api-key="$(cat /etc/usm/api-key)" sites health SITE_ID >> /var/log/usm.log 2>&1
```

### How do I parse JSON output?

```bash
# Extract site IDs
usm sites list --output json | jq -r '.sites[].id'

# Count devices
usm devices list SITE_ID --output json | jq '.devices | length'

# Find offline devices
usm devices list SITE_ID --output json | jq '.devices[] | select(.status == "offline")'
```

### Can I use the CLI with Ansible?

```yaml
# Ansible task
- name: Get UniFi sites
  shell: usm sites list --output json
  environment:
    USM_API_KEY: "{{ usm_api_key }}"
  register: usm_sites

- name: Parse sites
  set_fact:
    sites: "{{ usm_sites.stdout | from_json }}"
```

### How do I handle errors in scripts?

```bash
#!/bin/bash
set -e  # Exit on error

if ! usm sites get "$SITE_ID" > /dev/null 2>&1; then
  echo "Failed to get site"
  exit 1
fi

# Check specific exit codes
usm sites list
exit_code=$?

case $exit_code in
  0) echo "Success" ;;
  2) echo "Auth failed" ;;
  5) echo "Rate limited" ;;
  *) echo "Error: $exit_code" ;;
esac
```

### Can I monitor multiple sites?

```bash
#!/bin/bash
# multi-site-monitor.sh

SITES=("site1" "site2" "site3")

for site_id in "${SITES[@]}"; do
  echo "Checking site: $site_id"
  usm sites health "$site_id" --output json
  sleep 2
done
```

## Performance and Limits

### How many sites can I manage?

There's no hard limit, but:
- Large numbers may require pagination
- Consider using `--page-size 0` to fetch all
- Rate limiting applies per API key

### What's the maximum page size?

- Default: 50 items per page
- Maximum: 1000 items per page
- Use `--page-size 0` for all items (may be slow for large datasets)

### How fast are the commands?

Typical response times:
- Cloud API: 200-800ms
- Local Controller: 50-200ms
- With retries/rate limiting: add exponential backoff

### Can I run multiple commands in parallel?

Yes, but be mindful of:
- Rate limits (100 req/min for Cloud API)
- UniFi controller performance
- Concurrent connection limits

```bash
# Run in parallel (with care)
usm sites health site1 &
usm sites health site2 &
usm sites health site3 &
wait
```

## Still Have Questions?

- Check the [Usage Guide](USAGE.md)
- Review [API Documentation](API.md)
- Open a GitHub issue
- Start a GitHub Discussion

---

**Tip**: Use `usm --help` and `usm <command> --help` for command-specific help.
