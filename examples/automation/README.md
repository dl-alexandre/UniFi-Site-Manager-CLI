# Automation Scripts

Production-ready scripts for automating UniFi network management.

## Scripts Overview

| Script | Purpose | Schedule |
|--------|---------|----------|
| `health-check.sh` | Monitor site health | Every 5 minutes |
| `backup-config.sh` | Backup configurations | Daily |
| `firmware-update.sh` | Update device firmware | Weekly |
| `client-audit.sh` | Audit connected clients | Monthly |
| `alert-monitor.sh` | Monitor and alert on issues | Real-time |

## Prerequisites

```bash
# Install dependencies
# macOS
brew install jq curl

# Ubuntu/Debian
sudo apt-get install jq curl

# RHEL/CentOS
sudo yum install jq curl

# Verify
jq --version
curl --version
```

## 1. Health Check Script

Monitors site health and reports issues.

**File**: `health-check.sh`

```bash
#!/bin/bash
# health-check.sh - Monitor site health

set -euo pipefail

# Configuration
SITE_ID="${USM_SITE_ID:-}"
SLACK_WEBHOOK="${SLACK_WEBHOOK_URL:-}"
EMAIL="${ALERT_EMAIL:-}"
LOG_FILE="/var/log/usm-health-check.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if required vars are set
if [[ -z "$SITE_ID" ]]; then
    log "${RED}Error: USM_SITE_ID not set${NC}"
    exit 1
fi

# Function to send Slack notification
send_slack() {
    local message="$1"
    if [[ -n "$SLACK_WEBHOOK" ]]; then
        curl -s -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"$message\"}" \
            "$SLACK_WEBHOOK" > /dev/null
    fi
}

# Function to send email
send_email() {
    local subject="$1"
    local body="$2"
    if [[ -n "$EMAIL" ]]; then
        echo "$body" | mail -s "$subject" "$EMAIL"
    fi
}

log "Starting health check for site: $SITE_ID"

# Get site health
if ! HEALTH=$(usm sites health "$SITE_ID" --output json 2>&1); then
    log "${RED}Error: Failed to get health data - $HEALTH${NC}"
    send_slack "❌ USM Health Check Failed for site $SITE_ID"
    exit 1
fi

# Check overall status
OVERALL_STATUS=$(echo "$HEALTH" | jq -r '.status // "unknown"')
log "Overall status: $OVERALL_STATUS"

# Initialize issue counter
ISSUES=0

# Check devices
OFFLINE_DEVICES=$(echo "$HEALTH" | jq '.devices.offline // 0')
if [[ "$OFFLINE_DEVICES" -gt 0 ]]; then
    log "${YELLOW}Warning: $OFFLINE_DEVICES devices offline${NC}"
    ISSUES=$((ISSUES + 1))
    
    # Get offline device details
    OFFLINE_LIST=$(usm devices list "$SITE_ID" --status offline --output json | \
        jq -r '.devices[] | "- \(.name) (\(.mac))"')
    
    send_slack "⚠️ $OFFLINE_DEVICES devices offline in site $SITE_ID"
    send_email "USM Alert: Devices Offline" "$OFFLINE_LIST"
fi

# Check WAN
WAN_STATUS=$(echo "$HEALTH" | jq -r '.wan.status // "unknown"')
if [[ "$WAN_STATUS" != "up" ]]; then
    log "${RED}Warning: WAN is $WAN_STATUS${NC}"
    ISSUES=$((ISSUES + 1))
    send_slack "🚨 WAN is $WAN_STATUS in site $SITE_ID"
fi

# Check WLAN
WLAN_STATUS=$(echo "$HEALTH" | jq -r '.wlan.status // "unknown"')
if [[ "$WLAN_STATUS" != "healthy" ]]; then
    log "${YELLOW}Warning: WLAN status is $WLAN_STATUS${NC}"
    ISSUES=$((ISSUES + 1))
fi

# Summary
if [[ "$ISSUES" -eq 0 ]]; then
    log "${GREEN}✅ Health check passed - no issues found${NC}"
else
    log "${YELLOW}⚠️ Health check completed with $ISSUES issues${NC}"
fi

exit "$ISSUES"
```

**Usage**:
```bash
chmod +x health-check.sh

# Run manually
./health-check.sh

# Or with cron (every 5 minutes)
crontab -e
# Add: */5 * * * * /path/to/health-check.sh
```

## 2. Backup Configuration Script

Backups all site configurations.

**File**: `backup-config.sh`

```bash
#!/bin/bash
# backup-config.sh - Backup UniFi configurations

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_PATH="$BACKUP_DIR/$DATE"

# Create backup directory
mkdir -p "$BACKUP_PATH"

echo "Starting backup: $BACKUP_PATH"

# Backup all sites
usm sites list --output json | jq -r '.sites[].id' | while read -r site_id; do
    echo "Backing up site: $site_id"
    
    # Create site directory
    mkdir -p "$BACKUP_PATH/$site_id"
    
    # Site details
    usm sites get "$site_id" --output json > "$BACKUP_PATH/$site_id/site.json"
    
    # Site health
    usm sites health "$site_id" --output json > "$BACKUP_PATH/$site_id/health.json" || true
    
    # Devices
    usm devices list "$site_id" --output json > "$BACKUP_PATH/$site_id/devices.json"
    
    # WLANs
    usm wlans list "$site_id" --output json > "$BACKUP_PATH/$site_id/wlans.json"
    
    # Networks
    usm networks list "$site_id" --output json > "$BACKUP_PATH/$site_id/networks.json" || true
    
    # Clients
    usm clients list "$site_id" --output json > "$BACKUP_PATH/$site_id/clients.json"
done

# Create archive
tar -czf "$BACKUP_PATH.tar.gz" -C "$BACKUP_DIR" "$DATE"
rm -rf "$BACKUP_PATH"

# Clean old backups
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: $BACKUP_PATH.tar.gz"
```

## 3. Firmware Update Script

Automates firmware updates with safety checks.

**File**: `firmware-update.sh`

```bash
#!/bin/bash
# firmware-update.sh - Update device firmware

set -euo pipefail

# Configuration
SITE_ID="${USM_SITE_ID:-}"
DRY_RUN="${DRY_RUN:-true}"
AUTO_CONFIRM="${AUTO_CONFIRM:-false}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [[ -z "$SITE_ID" ]]; then
    echo -e "${RED}Error: USM_SITE_ID not set${NC}"
    exit 1
fi

echo "Checking for firmware updates..."
echo "Site ID: $SITE_ID"
echo "Mode: $([[ "$DRY_RUN" == "true" ]] && echo "DRY RUN" || echo "LIVE")"

# Get devices with updates available
DEVICES=$(usm devices list "$SITE_ID" --output json | \
    jq -r '.devices[] | select(.upgrade_available == true) | [.id, .name, .version] | @tsv')

if [[ -z "$DEVICES" ]]; then
    echo -e "${GREEN}No firmware updates available${NC}"
    exit 0
fi

echo -e "\n${YELLOW}Devices with available updates:${NC}"
echo "$DEVICES" | while IFS=$'\t' read -r id name version; do
    echo "  - $name (current: $version)"
done

# Count devices
COUNT=$(echo "$DEVICES" | wc -l)
echo -e "\nTotal devices to update: $COUNT"

# Confirm
if [[ "$DRY_RUN" != "true" && "$AUTO_CONFIRM" != "true" ]]; then
    read -p "Proceed with updates? (y/N): " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        echo "Cancelled"
        exit 0
    fi
fi

# Update devices
if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "\n${YELLOW}DRY RUN - Would update:${NC}"
    echo "$DEVICES" | while IFS=$'\t' read -r id name version; do
        echo "  - $name"
    done
else
    echo -e "\n${YELLOW}Updating devices...${NC}"
    echo "$DEVICES" | while IFS=$'\t' read -r id name version; do
        echo "Updating $name..."
        if usm devices upgrade "$SITE_ID" "$id"; then
            echo -e "  ${GREEN}✓ Updated successfully${NC}"
        else
            echo -e "  ${RED}✗ Update failed${NC}"
        fi
        sleep 30  # Wait between updates
done
fi

echo -e "\n${GREEN}Firmware update process complete${NC}"
```

## 4. Client Audit Script

Audits and reports on network clients.

**File**: `client-audit.sh`

```bash
#!/bin/bash
# client-audit.sh - Audit connected clients

set -euo pipefail

# Configuration
SITE_ID="${USM_SITE_ID:-}"
OUTPUT_DIR="${OUTPUT_DIR:-./audits}"
DATE=$(date +%Y-%m-%d)
OUTPUT_FILE="$OUTPUT_DIR/client-audit-$DATE.csv"

mkdir -p "$OUTPUT_DIR"

if [[ -z "$SITE_ID" ]]; then
    echo "Error: USM_SITE_ID not set"
    exit 1
fi

echo "Running client audit for site: $SITE_ID"

# CSV Header
echo "timestamp,name,hostname,mac,ip,type,ssid,ap_name,signal,uptime_seconds" > "$OUTPUT_FILE"

# Get all clients
usm clients list "$SITE_ID" --page-size 0 --output json | jq -r '
    .clients[] | [
        .name // "Unknown",
        .hostname // "Unknown",
        .mac,
        .ip // "Unknown",
        .type // "unknown",
        .ssid // "N/A",
        .ap_name // "N/A",
        .signal // "N/A",
        .uptime // 0
    ] | @csv
' | while IFS= read -r line; do
    echo "\"$DATE\",$line" >> "$OUTPUT_FILE"
done

# Generate summary
echo -e "\n=== Client Audit Summary ==="
echo "Date: $DATE"
echo "Total clients: $(usm clients list "$SITE_ID" --output json | jq '.clients | length')"
echo "Wireless clients: $(usm clients list "$SITE_ID" --output json | jq '[.clients[] | select(.type == "wireless")] | length')"
echo "Wired clients: $(usm clients list "$SITE_ID" --output json | jq '[.clients[] | select(.type == "wired")] | length')"
echo ""
echo "Report saved: $OUTPUT_FILE"
```

## 5. Alert Monitor Script

Monitors alerts and notifies on critical issues.

**File**: `alert-monitor.sh`

```bash
#!/bin/bash
# alert-monitor.sh - Monitor and alert on UniFi alerts

set -euo pipefail

# Configuration
CHECK_INTERVAL="${CHECK_INTERVAL:-60}"
SLACK_WEBHOOK="${SLACK_WEBHOOK_URL:-}"
SEVERITIES="${ALERT_SEVERITIES:-high critical}"
STATE_FILE="/tmp/usm-alert-monitor.state"

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to send notification
notify() {
    local alert="$1"
    local severity=$(echo "$alert" | jq -r '.severity')
    local message=$(echo "$alert" | jq -r '.message')
    local site_id=$(echo "$alert" | jq -r '.site_id')
    
    local color="warning"
    [[ "$severity" == "critical" ]] && color="danger"
    [[ "$severity" == "high" ]] && color="warning"
    
    if [[ -n "$SLACK_WEBHOOK" ]]; then
        curl -s -X POST -H 'Content-type: application/json' \
            --data "{
                \"attachments\": [{
                    \"color\": \"$color\",
                    \"title\": \"UniFi Alert: $severity\",
                    \"text\": \"$message\",
                    \"fields\": [
                        {\"title\": \"Site\", \"value\": \"$site_id\", \"short\": true},
                        {\"title\": \"Severity\", \"value\": \"$severity\", \"short\": true}
                    ]
                }]
            }" \
            "$SLACK_WEBHOOK" > /dev/null
    fi
    
    echo -e "${RED}[$severity]${NC} $message (site: $site_id)"
}

# Initialize state file
if [[ ! -f "$STATE_FILE" ]]; then
    echo "[]" > "$STATE_FILE"
fi

echo "Starting alert monitor..."
echo "Check interval: ${CHECK_INTERVAL}s"
echo "Monitoring severities: $SEVERITIES"

while true; do
    # Get current alerts
    CURRENT_ALERTS=$(usm alerts list --output json 2>/dev/null || echo "[]")
    
    # Filter by severity
    for severity in $SEVERITIES; do
        echo "$CURRENT_ALERTS" | jq -r --arg sev "$severity" \
            '.alerts[] | select(.severity == $sev and .acknowledged == false)' | \
        while IFS= read -r alert; do
            [[ -z "$alert" ]] && continue
            
            alert_id=$(echo "$alert" | jq -r '.id')
            
            # Check if we've already notified
            if ! grep -q "$alert_id" "$STATE_FILE" 2>/dev/null; then
                notify "$alert"
                echo "$alert_id" >> "$STATE_FILE"
            fi
        done
    done
    
    sleep "$CHECK_INTERVAL"
done
```

## Installation

```bash
# 1. Create automation directory
mkdir -p /opt/usm-automation
cd /opt/usm-automation

# 2. Copy scripts
cp /path/to/scripts/*.sh .

# 3. Set permissions
chmod +x *.sh

# 4. Create configuration file
cat > /opt/usm-automation/config.env << 'EOF'
USM_API_KEY=your-api-key
USM_SITE_ID=your-site-id
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx
ALERT_EMAIL=admin@example.com
BACKUP_DIR=/backup/usm
EOF

chmod 600 config.env

# 5. Add to crontab
sudo crontab -e
```

## Crontab Examples

```bash
# Run every 5 minutes
*/5 * * * * cd /opt/usm-automation && source config.env && ./health-check.sh

# Daily backup at 2 AM
0 2 * * * cd /opt/usm-automation && source config.env && ./backup-config.sh

# Weekly firmware check (Sunday 3 AM)
0 3 * * 0 cd /opt/usm-automation && source config.env && ./firmware-update.sh

# Monthly client audit (1st of month at 4 AM)
0 4 1 * * cd /opt/usm-automation && source config.env && ./client-audit.sh

# Alert monitor (run as daemon)
@reboot cd /opt/usm-automation && source config.env && nohup ./alert-monitor.sh > /var/log/usm-alerts.log 2>&1 &
```

## Docker Usage

```dockerfile
# Dockerfile
FROM alpine:latest
RUN apk add --no-cache bash jq curl
COPY *.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/*.sh
ENTRYPOINT ["health-check.sh"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  health-check:
    build: .
    environment:
      - USM_API_KEY=${USM_API_KEY}
      - USM_SITE_ID=${USM_SITE_ID}
    env_file:
      - .env
    volumes:
      - ./logs:/var/log
```

---

For more examples, see:
- [Basic Usage](../basic/)
- [Monitoring Setup](../monitoring/)
- [CI/CD Integration](../ci-cd/)
