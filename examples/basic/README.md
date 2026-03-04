# Basic Usage Examples

Simple examples demonstrating common CLI operations.

## Setup

```bash
# Set your API key
export USM_API_KEY="your-api-key-here"

# Or initialize interactively
usm init
```

## Site Operations

### List Sites

```bash
# Basic list
usm sites list

# Get all sites (no pagination)
usm sites list --page-size 0

# Search by name
usm sites list --search "office"

# JSON output
usm sites list --output json
```

### Get Site Details

```bash
SITE_ID="60abcdef1234567890abcdef"

# Get site info
usm sites get $SITE_ID

# Get health
usm sites health $SITE_ID

# Get statistics
usm sites stats $SITE_ID --period day
```

### Create and Manage Sites

```bash
# Create new site
NEW_SITE=$(usm sites create "New Branch Office" --description "Regional office" --output json)
SITE_ID=$(echo $NEW_SITE | jq -r '.id')
echo "Created site: $SITE_ID"

# Update site
usm sites update $SITE_ID --name "Updated Branch Office"

# Delete site (with confirmation)
usm sites delete $SITE_ID
```

## Device Operations

### List Devices

```bash
SITE_ID="60abcdef1234567890abcdef"

# All devices
usm devices list $SITE_ID

# Filter by status
usm devices list $SITE_ID --status online
usm devices list $SITE_ID --status offline

# Filter by type
usm devices list $SITE_ID --type ap
usm devices list $SITE_ID --type switch
usm devices list $SITE_ID --type gateway

# Get device count
usm devices list $SITE_ID --output json | jq '.devices | length'
```

### Device Actions

```bash
SITE_ID="60abcdef1234567890abcdef"
DEVICE_ID="60fedcba0987654321fedcba"

# Get device details
usm devices get $SITE_ID $DEVICE_ID

# Restart device
usm devices restart $SITE_ID $DEVICE_ID --force

# Upgrade firmware
usm devices upgrade $SITE_ID $DEVICE_ID
```

## Client Operations

### List Clients

```bash
SITE_ID="60abcdef1234567890abcdef"

# All clients
usm clients list $SITE_ID

# Wired only
usm clients list $SITE_ID --wired-only

# Wireless only
usm clients list $SITE_ID --wireless-only

# Search by name
usm clients list $SITE_ID --search "iPhone"

# Get count
usm clients list $SITE_ID --output json | jq '.clients | length'
```

### Block/Unblock Clients

```bash
SITE_ID="60abcdef1234567890abcdef"

# Block client
usm clients block $SITE_ID "aa:bb:cc:dd:ee:ff"

# Unblock client
usm clients unblock $SITE_ID "aa:bb:cc:dd:ee:ff"

# Get client stats
usm clients stats $SITE_ID "aa:bb:cc:dd:ee:ff"
```

## WLAN Operations

### List WLANs

```bash
SITE_ID="60abcdef1234567890abcdef"

usm wlans list $SITE_ID
usm wlans list $SITE_ID --output json
```

### Create WLAN

```bash
SITE_ID="60abcdef1234567890abcdef"

# Basic WLAN
usm wlans create $SITE_ID "Office WiFi" "Office-5G" \
  --password "securepassword" \
  --security wpapsk

# Guest network with VLAN
usm wlans create $SITE_ID "Guest WiFi" "Guest-Network" \
  --password "guestpass" \
  --security wpapsk \
  --vlan 20 \
  --guest

# Hidden network
usm wlans create $SITE_ID "Admin Network" "Admin-WiFi" \
  --password "adminsecret" \
  --security wpapsk \
  --hide-ssid
```

### Update WLAN

```bash
SITE_ID="60abcdef1234567890abcdef"
WLAN_ID="60wlan0987654321abcdef"

# Change password
usm wlans update $SITE_ID $WLAN_ID --password "newpassword"

# Disable WLAN
usm wlans update $SITE_ID $WLAN_ID --enabled=false

# Enable WLAN
usm wlans update $SITE_ID $WLAN_ID --enabled=true
```

### Delete WLAN

```bash
SITE_ID="60abcdef1234567890abcdef"
WLAN_ID="60wlan0987654321abcdef"

usm wlans delete $SITE_ID $WLAN_ID
```

## Alert Operations

### List Alerts

```bash
# All alerts
usm alerts list

# Site-specific
usm alerts list --site-id 60abcdef1234567890abcdef

# Include archived
usm alerts list --archived
```

### Manage Alerts

```bash
SITE_ID="60abcdef1234567890abcdef"
ALERT_ID="60alert0987654321abcdef"

# Acknowledge
usm alerts ack $SITE_ID $ALERT_ID

# Archive
usm alerts archive $SITE_ID $ALERT_ID
```

## Network Operations

### List Networks

```bash
SITE_ID="60abcdef1234567890abcdef"

usm networks list $SITE_ID
usm networks list $SITE_ID --output json
```

## Event Operations

### List Events

```bash
# All events
usm events list

# Site-specific
usm events list --site-id 60abcdef1234567890abcdef
```

## Host Operations

### List Hosts

```bash
usm hosts list
usm hosts list --search "UDM"
```

### Host Actions

```bash
HOST_ID="60abcdef1234567890abcdef"

# Get details
usm hosts get $HOST_ID

# Get health
usm hosts health $HOST_ID

# Get stats
usm hosts stats $HOST_ID --period day

# Restart
usm hosts restart $HOST_ID --force
```

## Local Controller Examples

### Connect to UDM

```bash
export USM_LOCAL=true
export USM_HOST="192.168.1.1"
export USM_USERNAME="admin"
export USM_PASSWORD="yourpassword"
export USM_INSECURE=true

# List sites
usm sites list

# List devices
usm devices list default

# List clients
usm clients list default
```

## Output Formatting

### Table Format

```bash
# Default (colorful table)
usm sites list

# No colors (for piping)
usm sites list --color never

# No headers (for parsing)
usm sites list --no-headers
```

### JSON Format

```bash
# Pretty JSON
usm sites list --output json

# Extract specific fields
usm sites list --output json | jq '.sites[].name'

# Complex queries
usm devices list SITE_ID --output json | \
  jq '.devices[] | select(.status == "offline") | {name, mac}'
```

## Help and Version

```bash
# Show version
usm version

# Check for updates
usm check-update

# Show help
usm --help

# Command-specific help
usm sites --help
usm sites list --help
```

## Running Examples

```bash
# Make sure usm is installed
which usm

# Set your API key
export USM_API_KEY="your-api-key"

# Run any example
usm sites list
```

---

For more advanced examples, see:
- [Automation Scripts](../automation/)
- [Monitoring Setup](../monitoring/)
- [CI/CD Integration](../ci-cd/)
