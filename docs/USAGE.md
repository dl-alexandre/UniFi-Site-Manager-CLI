# Usage Guide

Comprehensive guide for using UniFi Site Manager CLI with real-world examples.

## Table of Contents

- [Getting Started](#getting-started)
- [Global Options](#global-options)
- [Authentication](#authentication)
- [Site Management](#site-management)
- [Host/Console Management](#hostconsole-management)
- [Device Management](#device-management)
- [Client Management](#client-management)
- [WLAN Management](#wlan-management)
- [Alert Management](#alert-management)
- [Event Management](#event-management)
- [Network Management](#network-management)
- [Output Formats](#output-formats)
- [Automation Examples](#automation-examples)
- [Error Handling](#error-handling)

## Getting Started

### First Run

```bash
# Initialize configuration (interactive)
usm init

# Or use environment variables
export USM_API_KEY="your-api-key"

# Test connection
usm whoami
```

### Quick Reference

| Task | Command |
|------|---------|
| List sites | `usm sites list` |
| Get site details | `usm sites get <site-id>` |
| List devices | `usm devices list <site-id>` |
| List clients | `usm clients list <site-id>` |
| Create WLAN | `usm wlans create <site-id> <name> <ssid>` |
| Check health | `usm sites health <site-id>` |
| View alerts | `usm alerts list` |

## Global Options

Options available for all commands:

```bash
usm [global-options] <command> [command-options]
```

### Output Options

```bash
# Change output format
usm sites list --output json
usm sites list --output table  # default

# Disable colors
usm sites list --color never

# Disable table headers (useful for scripts)
usm sites list --no-headers
```

### Cloud Mode Options

```bash
# Specify API key via flag
usm --api-key="your-key" sites list

# Custom base URL
usm --base-url="https://api.ui.com" sites list

# Increase timeout
usm --timeout 60 sites list
```

### Local Mode Options

```bash
# Enable local mode
usm --local --host=192.168.1.1 --username=admin --password=secret sites list

# With environment variables
export USM_LOCAL=true
export USM_HOST=192.168.1.1
export USM_USERNAME=admin
export USM_PASSWORD=secret
usm sites list
```

### Debug and Verbose

```bash
# Verbose output
usm -v sites list

# Debug mode (shows API requests/responses, credentials redacted)
usm --debug sites list
```

## Authentication

### Cloud API Mode

**Method 1: Environment Variable (Recommended)**

```bash
export USM_API_KEY="your-api-key"
usm whoami
```

**Method 2: Command Flag**

```bash
usm --api-key="your-api-key" whoami
```

**Method 3: Config File**

```bash
# Created by 'usm init'
# Config file does NOT store API keys for security
cat ~/.config/usm/config.yaml
```

### Local Controller Mode

**Environment Variables**

```bash
export USM_LOCAL=true
export USM_HOST="192.168.1.1"
export USM_USERNAME="admin"
export USM_PASSWORD="your-password"
```

**Command Flags**

```bash
usm --local --host=192.168.1.1 --username=admin --password=secret sites list
```

**Security Note**: Always use environment variables for passwords, never command flags in production.

### Getting API Keys

1. Log into [unifi.ui.com](https://unifi.ui.com)
2. Navigate to Settings → Control Plane → Integrations → API Keys
3. Click "Create API Key"
4. Copy and securely store the key

## Site Management

### List Sites

```bash
# Basic list
usm sites list

# With pagination
usm sites list --page-size 100
usm sites list --page-size 0  # Fetch all

# Search by name
usm sites list --search "office"

# JSON output for scripting
usm sites list --output json | jq '.sites[].name'
```

### Get Site Details

```bash
# Get single site
usm sites get 60abcdef1234567890abcdef

# JSON output
usm sites get 60abcdef1234567890abcdef --output json
```

### Create Site

```bash
# Basic creation
usm sites create "New Office"

# With description
usm sites create "New Office" --description "Main headquarters"

# Capture new site ID
SITE_ID=$(usm sites create "New Office" --output json | jq -r '.id')
echo "Created site: $SITE_ID"
```

### Update Site

```bash
# Update name
usm sites update 60abcdef1234567890abcdef --name "Updated Name"

# Update description
usm sites update 60abcdef1234567890abcdef --description "New description"

# Update both
usm sites update 60abcdef1234567890abcdef \
  --name "New Name" \
  --description "New description"
```

### Delete Site

```bash
# With confirmation prompt
usm sites delete 60abcdef1234567890abcdef

# Force delete (no confirmation)
usm sites delete 60abcdef1234567890abcdef --force

# Delete multiple sites
for site in site1 site2 site3; do
  usm sites delete $site --force
done
```

### Site Health

```bash
# Get health status
usm sites health 60abcdef1234567890abcdef

# Check specific components
usm sites health 60abcdef1234567890abcdef --output json | jq '.devices[]'
```

### Site Statistics

```bash
# Current stats
usm sites stats 60abcdef1234567890abcdef

# Daily stats
usm sites stats 60abcdef1234567890abcdef --period day

# Weekly stats
usm sites stats 60abcdef1234567890abcdef --period week

# Monthly stats
usm sites stats 60abcdef1234567890abcdef --period month

# JSON output for analysis
usm sites stats 60abcdef1234567890abcdef --period day --output json
```

## Host/Console Management

### List Hosts

```bash
# All hosts
usm hosts list

# With pagination
usm hosts list --page-size 100

# Search by name
usm hosts list --search "UDM"

# Get total count
usm hosts list --output json | jq '.hosts | length'
```

### Get Host Details

```bash
usm hosts get 60abcdef1234567890abcdef
```

### Host Health

```bash
usm hosts health 60abcdef1234567890abcdef
```

### Host Statistics

```bash
# Daily stats
usm hosts stats 60abcdef1234567890abcdef --period day

# Weekly stats
usm hosts stats 60abcdef1234567890abcdef --period week
```

### Restart Host

```bash
# With confirmation
usm hosts restart 60abcdef1234567890abcdef

# Without confirmation
usm hosts restart 60abcdef1234567890abcdef --force
```

## Device Management

### List Devices

```bash
# All devices for a site
usm devices list 60abcdef1234567890abcdef

# Pagination
usm devices list 60abcdef1234567890abcdef --page-size 100

# Filter by status
usm devices list 60abcdef1234567890abcdef --status online
usm devices list 60abcdef1234567890abcdef --status offline
usm devices list 60abcdef1234567890abcdef --status pending

# Filter by type
usm devices list 60abcdef1234567890abcdef --type ap        # Access Points
usm devices list 60abcdef1234567890abcdef --type switch    # Switches
usm devices list 60abcdef1234567890abcdef --type gateway  # Gateways

# Combine filters
usm devices list 60abcdef1234567890abcdef --status offline --type ap

# JSON for processing
usm devices list 60abcdef1234567890abcdef --output json | jq '.devices[].name'
```

### Get Device Details

```bash
usm devices get 60abcdef1234567890abcdef 60fedcba0987654321fedcba
```

### Restart Device

```bash
# Restart with confirmation
usm devices restart 60abcdef1234567890abcdef 60fedcba0987654321fedcba

# Restart without confirmation
usm devices restart 60abcdef1234567890abcdef 60fedcba0987654321fedcba --force

# Restart multiple devices
for device in device1 device2 device3; do
  usm devices restart 60abcdef1234567890abcdef $device --force
  sleep 5  # Wait between restarts
done
```

### Upgrade Firmware

```bash
# Upgrade single device
usm devices upgrade 60abcdef1234567890abcdef 60fedcba0987654321fedcba

# Check for upgrades and apply
usm devices list 60abcdef1234567890abcdef --output json | \
  jq -r '.devices[] | select(.upgrade_available == true) | .id' | \
  while read device_id; do
    usm devices upgrade 60abcdef1234567890abcdef $device_id
  done
```

### Adopt Device

```bash
# Adopt by MAC address
usm devices adopt 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"

# Adopt multiple
for mac in "aa:bb:cc:dd:ee:ff" "11:22:33:44:55:66"; do
  usm devices adopt 60abcdef1234567890abcdef "$mac"
done
```

## Client Management

### List Clients

```bash
# All clients
usm clients list 60abcdef1234567890abcdef

# Pagination
usm clients list 60abcdef1234567890abcdef --page-size 500

# Filter by type
usm clients list 60abcdef1234567890abcdef --wired-only
usm clients list 60abcdef1234567890abcdef --wireless-only

# Search by name or MAC
usm clients list 60abcdef1234567890abcdef --search "iPhone"
usm clients list 60abcdef1234567890abcdef --search "aa:bb:cc"

# JSON output
usm clients list 60abcdef1234567890abcdef --output json | jq '.clients[] | {name, ip, mac}'
```

### Get Client Statistics

```bash
usm clients stats 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

### Block Client

```bash
# Block by MAC
usm clients block 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"

# Block multiple
for mac in "aa:bb:cc:dd:ee:ff" "11:22:33:44:55:66"; do
  usm clients block 60abcdef1234567890abcdef "$mac"
done
```

### Unblock Client

```bash
usm clients unblock 60abcdef1234567890abcdef "aa:bb:cc:dd:ee:ff"
```

### Find Client by Name

```bash
# Find client's MAC by hostname
MAC=$(usm clients list 60abcdef1234567890abcdef --search "Johns-iPhone" --output json | \
  jq -r '.clients[0].mac')
echo "MAC: $MAC"
```

## WLAN Management

### List WLANs

```bash
# All WLANs for a site
usm wlans list 60abcdef1234567890abcdef

# JSON output
usm wlans list 60abcdef1234567890abcdef --output json | jq '.wlans[].ssid'
```

### Get WLAN Details

```bash
usm wlans get 60abcdef1234567890abcdef 60wlan0987654321abcdef
```

### Create WLAN

```bash
# Basic WLAN with password
usm wlans create 60abcdef1234567890abcdef "Office WiFi" "Office-5G" \
  --password "secure-password-123" \
  --security wpapsk

# Advanced WLAN
usm wlans create 60abcdef1234567890abcdef "Guest Network" "Guest-WiFi" \
  --password "guest123" \
  --security wpapsk \
  --vlan 20 \
  --band both \
  --wpa3

# Guest network (no password)
usm wlans create 60abcdef1234567890abcdef "Public WiFi" "Free-WiFi" \
  --security open \
  --band 2.4

# IoT network with restrictions
usm wlans create 60abcdef1234567890abcdef "IoT Devices" "IoT-Network" \
  --password "iot-secret" \
  --security wpapsk \
  --vlan 30 \
  --band 2.4
```

### Update WLAN

```bash
# Change password
usm wlans update 60abcdef1234567890abcdef 60wlan0987654321abcdef \
  --password "new-password"

# Disable WLAN
usm wlans update 60abcdef1234567890abcdef 60wlan0987654321abcdef \
  --enabled=false

# Enable WLAN
usm wlans update 60abcdef1234567890abcdef 60wlan0987654321abcdef \
  --enabled=true

# Update multiple settings
usm wlans update 60abcdef1234567890abcdef 60wlan0987654321abcdef \
  --password "updated-password" \
  --security wpapsk \
  --wpa3
```

### Delete WLAN

```bash
# Delete single WLAN
usm wlans delete 60abcdef1234567890abcdef 60wlan0987654321abcdef

# Delete multiple
for wlan in wlan1 wlan2 wlan3; do
  usm wlans delete 60abcdef1234567890abcdef $wlan
done
```

## Alert Management

### List Alerts

```bash
# All alerts
usm alerts list

# Site-specific alerts
usm alerts list --site-id 60abcdef1234567890abcdef

# Include archived
usm alerts list --archived

# JSON for processing
usm alerts list --output json | jq '.alerts[] | select(.severity == "high")'
```

### Acknowledge Alert

```bash
usm alerts ack 60abcdef1234567890abcdef 60alert0987654321abcdef
```

### Archive Alert

```bash
usm alerts archive 60abcdef1234567890abcdef 60alert0987654321abcdef
```

### Alert Automation

```bash
# Acknowledge all unacknowledged alerts
usm alerts list --output json | \
  jq -r '.alerts[] | select(.acknowledged == false) | "\(.site_id) \(.id)"' | \
  while read site_id alert_id; do
    usm alerts ack "$site_id" "$alert_id"
  done
```

## Event Management

### List Events

```bash
# All events
usm events list

# Site-specific events
usm events list --site-id 60abcdef1234567890abcdef

# JSON output
usm events list --output json | jq '.events[] | {time, type, message}'
```

## Network Management

### List Networks

```bash
# All networks for a site
usm networks list 60abcdef1234567890abcdef

# JSON output
usm networks list 60abcdef1234567890abcdef --output json | jq '.networks[] | {name, vlan}'
```

## Output Formats

### Table Format (Default)

```bash
usm sites list
# Output:
# ┌─────────────────────────┬─────────────────┬──────────┬──────────┐
# │ ID                      │ NAME            │ HOSTS    │ DEVICES  │
# ├─────────────────────────┼─────────────────┼──────────┼──────────┤
# │ 60abcdef1234567890abcde │ Main Office     │ 1        │ 12       │
# └─────────────────────────┴─────────────────┴──────────┴──────────┘
```

### JSON Format

```bash
usm sites list --output json
# Output:
# {
#   "sites": [
#     {
#       "id": "60abcdef1234567890abcdef",
#       "name": "Main Office",
#       "hosts": 1,
#       "devices": 12
#     }
#   ]
# }
```

### Processing JSON Output

```bash
# Extract site IDs
usm sites list --output json | jq -r '.sites[].id'

# Count devices
usm devices list SITE_ID --output json | jq '.devices | length'

# Find offline devices
usm devices list SITE_ID --output json | jq '.devices[] | select(.status == "offline")'

# Get client names and IPs
usm clients list SITE_ID --output json | jq '.clients[] | {name, ip}'
```

## Automation Examples

### Health Check Script

```bash
#!/bin/bash
# health-check.sh - Check site health and alert on issues

SITE_ID="60abcdef1234567890abcdef"
HEALTH=$(usm sites health "$SITE_ID" --output json)

# Check for offline devices
OFFLINE_COUNT=$(echo "$HEALTH" | jq '[.devices[] | select(.status == "offline")] | length')

if [ "$OFFLINE_COUNT" -gt 0 ]; then
  echo "⚠️  WARNING: $OFFLINE_COUNT devices are offline"
  echo "$HEALTH" | jq '.devices[] | select(.status == "offline") | {name, mac, status}'
  exit 1
else
  echo "✅ All devices are online"
  exit 0
fi
```

### Client Audit Script

```bash
#!/bin/bash
# client-audit.sh - List all clients with details

SITE_ID="60abcdef1234567890abcdef"
OUTPUT_FILE="clients-$(date +%Y%m%d).csv"

echo "Name,MAC,IP,Type,Connection,SSID,Signal" > "$OUTPUT_FILE"

usm clients list "$SITE_ID" --page-size 0 --output json | \
  jq -r '.clients[] | [.name, .mac, .ip, .type, .connection, .ssid // "N/A", .signal // "N/A"] | @csv' \
  >> "$OUTPUT_FILE"

echo "Client audit saved to: $OUTPUT_FILE"
```

### Backup Configuration

```bash
#!/bin/bash
# backup-config.sh - Backup site configuration

BACKUP_DIR="backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup sites
usm sites list --output json > "$BACKUP_DIR/sites.json"

# Backup each site's details
for site_id in $(usm sites list --output json | jq -r '.sites[].id'); do
  echo "Backing up site: $site_id"
  
  # Site details
  usm sites get "$site_id" --output json > "$BACKUP_DIR/site-$site-id.json"
  
  # Devices
  usm devices list "$site_id" --output json > "$BACKUP_DIR/devices-$site_id.json"
  
  # WLANs
  usm wlans list "$site_id" --output json > "$BACKUP_DIR/wlans-$site_id.json"
  
  # Networks
  usm networks list "$site_id" --output json > "$BACKUP_DIR/networks-$site_id.json"
done

echo "Backup complete: $BACKUP_DIR"
```

### Firmware Update Automation

```bash
#!/bin/bash
# firmware-update.sh - Update all devices with available upgrades

SITE_ID="60abcdef1234567890abcdef"

# Get devices with upgrades available
usm devices list "$SITE_ID" --output json | \
  jq -r '.devices[] | select(.upgrade_available == true) | .id' | \
  while read device_id; do
    echo "Upgrading device: $device_id"
    usm devices upgrade "$SITE_ID" "$device_id"
    sleep 10  # Wait between upgrades
done
```

### Daily Report Generation

```bash
#!/bin/bash
# daily-report.sh - Generate daily network report

SITE_ID="60abcdef1234567890abcdef"
REPORT_FILE="report-$(date +%Y-%m-%d).md"

cat > "$REPORT_FILE" << EOF
# Network Daily Report - $(date +%Y-%m-%d)

## Site Health
\`\`\`
$(usm sites health "$SITE_ID")
\`\`\`

## Device Status
\`\`\`
$(usm devices list "$SITE_ID")
\`\`\`

## Client Count
$(usm clients list "$SITE_ID" --output json | jq '.clients | length') clients connected

## Active Alerts
$(usm alerts list --site-id "$SITE_ID" --output json | jq '.alerts | length') unacknowledged alerts
EOF

echo "Report generated: $REPORT_FILE"
```

## Error Handling

### Exit Codes

| Code | Meaning | Handling |
|------|---------|----------|
| 0 | Success | Continue execution |
| 1 | General error | Check error message |
| 2 | Authentication failure | Verify API key |
| 3 | Permission denied | Check account permissions |
| 4 | Validation error | Check input parameters |
| 5 | Rate limited | Wait and retry |
| 6 | Network error | Check connectivity |

### Script Error Handling

```bash
#!/bin/bash

set -e  # Exit on error

SITE_ID="60abcdef1234567890abcdef"

# Check if command succeeded
if usm sites get "$SITE_ID" > /dev/null 2>&1; then
  echo "✅ Site exists"
else
  echo "❌ Site not found or error occurred"
  exit 1
fi

# Handle specific errors
case $? in
  0) echo "Success" ;;
  2) echo "Authentication failed - check API key" ;;
  3) echo "Permission denied" ;;
  5) echo "Rate limited - waiting..."; sleep 60 ;;
  *) echo "Unknown error" ;;
esac
```

### Rate Limiting

```bash
#!/bin/bash
# Handle rate limiting with exponential backoff

MAX_RETRIES=5
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if usm sites list > /dev/null 2>&1; then
    echo "Success!"
    break
  else
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 5 ]; then
      RETRY_COUNT=$((RETRY_COUNT + 1))
      WAIT_TIME=$((2 ** RETRY_COUNT))  # Exponential backoff
      echo "Rate limited. Waiting ${WAIT_TIME}s..."
      sleep $WAIT_TIME
    else
      echo "Error: $EXIT_CODE"
      exit 1
    fi
  fi
done
```

## Best Practices

1. **Use environment variables** for sensitive data (API keys, passwords)
2. **Use --output json** for automation and piping to jq
3. **Handle errors** with proper exit code checking
4. **Implement rate limiting** in scripts with sleep delays
5. **Use --no-headers** for parsing table output
6. **Enable debug mode** (`--debug`) when troubleshooting
7. **Cache site IDs** in scripts instead of fetching repeatedly
8. **Use pagination** for large datasets (`--page-size`)

## Next Steps

- See [API Documentation](API.md) for endpoint details
- Read [Configuration Guide](CONFIGURATION.md) for advanced settings
- Check [FAQ](FAQ.md) for common questions
