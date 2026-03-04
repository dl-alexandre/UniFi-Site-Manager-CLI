# Monitoring Setup

Integrate UniFi Site Manager CLI with monitoring systems.

## Prometheus Integration

### Prometheus Configuration

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'usm-exporter'
    static_configs:
      - targets: ['localhost:9100']
    scrape_interval: 30s
```

### Custom Exporter Script

**File**: `usm-prometheus-exporter.sh`

```bash
#!/bin/bash
# usm-prometheus-exporter.sh - Prometheus metrics exporter

set -euo pipefail

SITE_ID="${USM_SITE_ID:-}"
PORT="${EXPORTER_PORT:-9100}"

if [[ -z "$SITE_ID" ]]; then
    echo "Error: USM_SITE_ID not set"
    exit 1
fi

# Generate metrics
generate_metrics() {
    # Site health
    HEALTH=$(usm sites health "$SITE_ID" --output json 2>/dev/null || echo '{}')
    
    # Device metrics
    DEVICES=$(usm devices list "$SITE_ID" --output json 2>/dev/null || echo '{}')
    
    # Client metrics
    CLIENTS=$(usm clients list "$SITE_ID" --output json 2>/dev/null || echo '{}')
    
    # Output Prometheus format
    echo "# HELP usm_site_health Site health status (1=healthy, 0=unhealthy)"
    echo "# TYPE usm_site_health gauge"
    STATUS=$(echo "$HEALTH" | jq -r '.status // "unknown"')
    HEALTH_VALUE=$([[ "$STATUS" == "healthy" ]] && echo "1" || echo "0")
    echo "usm_site_health{site=\"$SITE_ID\"} $HEALTH_VALUE"
    
    echo "# HELP usm_devices_total Total number of devices"
    echo "# TYPE usm_devices_total gauge"
    TOTAL_DEVICES=$(echo "$DEVICES" | jq '.devices | length')
    echo "usm_devices_total{site=\"$SITE_ID\"} $TOTAL_DEVICES"
    
    echo "# HELP usm_devices_online Number of online devices"
    echo "# TYPE usm_devices_online gauge"
    ONLINE_DEVICES=$(echo "$DEVICES" | jq '[.devices[] | select(.status == "online")] | length')
    echo "usm_devices_online{site=\"$SITE_ID\"} $ONLINE_DEVICES"
    
    echo "# HELP usm_devices_offline Number of offline devices"
    echo "# TYPE usm_devices_offline gauge"
    OFFLINE_DEVICES=$(echo "$DEVICES" | jq '[.devices[] | select(.status == "offline")] | length')
    echo "usm_devices_offline{site=\"$SITE_ID\"} $OFFLINE_DEVICES"
    
    echo "# HELP usm_clients_total Total number of connected clients"
    echo "# TYPE usm_clients_total gauge"
    TOTAL_CLIENTS=$(echo "$CLIENTS" | jq '.clients | length')
    echo "usm_clients_total{site=\"$SITE_ID\"} $TOTAL_CLIENTS"
    
    echo "# HELP usm_clients_wireless Number of wireless clients"
    echo "# TYPE usm_clients_wireless gauge"
    WIRELESS_CLIENTS=$(echo "$CLIENTS" | jq '[.clients[] | select(.type == "wireless")] | length')
    echo "usm_clients_wireless{site=\"$SITE_ID\"} $WIRELESS_CLIENTS"
    
    echo "# HELP usm_clients_wired Number of wired clients"
    echo "# TYPE usm_clients_wired gauge"
    WIRED_CLIENTS=$(echo "$CLIENTS" | jq '[.clients[] | select(.type == "wired")] | length')
    echo "usm_clients_wired{site=\"$SITE_ID\"} $WIRED_CLIENTS"
    
    echo "# HELP usm_alerts_total Total number of unacknowledged alerts"
    echo "# TYPE usm_alerts_total gauge"
    ALERTS=$(usm alerts list --site-id "$SITE_ID" --output json 2>/dev/null || echo '{}')
    ALERTS_COUNT=$(echo "$ALERTS" | jq '.alerts | length')
    echo "usm_alerts_total{site=\"$SITE_ID\"} $ALERTS_COUNT"
}

# HTTP server
while true; do
    { echo -e "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n"; generate_metrics; } | \
        nc -l -p "$PORT" -q 1
done
```

### Systemd Service

```ini
# /etc/systemd/system/usm-exporter.service
[Unit]
Description=UniFi Site Manager Prometheus Exporter
After=network.target

[Service]
Type=simple
Environment="USM_API_KEY=your-api-key"
Environment="USM_SITE_ID=your-site-id"
Environment="EXPORTER_PORT=9100"
ExecStart=/usr/local/bin/usm-prometheus-exporter.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Install
sudo cp usm-prometheus-exporter.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/usm-prometheus-exporter.sh
sudo cp usm-exporter.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable usm-exporter
sudo systemctl start usm-exporter
```

## Grafana Dashboard

### Dashboard JSON

Create `grafana-dashboard.json`:

```json
{
  "dashboard": {
    "title": "UniFi Site Manager",
    "panels": [
      {
        "title": "Site Health",
        "type": "stat",
        "targets": [
          {
            "expr": "usm_site_health",
            "legendFormat": "Health"
          }
        ]
      },
      {
        "title": "Devices",
        "type": "stat",
        "targets": [
          {
            "expr": "usm_devices_online",
            "legendFormat": "Online"
          },
          {
            "expr": "usm_devices_offline",
            "legendFormat": "Offline"
          }
        ]
      },
      {
        "title": "Clients",
        "type": "timeseries",
        "targets": [
          {
            "expr": "usm_clients_total",
            "legendFormat": "Total"
          },
          {
            "expr": "usm_clients_wireless",
            "legendFormat": "Wireless"
          },
          {
            "expr": "usm_clients_wired",
            "legendFormat": "Wired"
          }
        ]
      },
      {
        "title": "Alerts",
        "type": "stat",
        "targets": [
          {
            "expr": "usm_alerts_total",
            "legendFormat": "Active Alerts"
          }
        ]
      }
    ]
  }
}
```

Import to Grafana:
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d @grafana-dashboard.json \
  http://admin:admin@localhost:3000/api/dashboards/db
```

## Nagios/Icinga Integration

### Check Script

**File**: `check_usm.sh`

```bash
#!/bin/bash
# check_usm.sh - Nagios/Icinga check plugin

SITE_ID="$1"
CHECK_TYPE="${2:-health}"

if [[ -z "$SITE_ID" ]]; then
    echo "UNKNOWN: Site ID required"
    exit 3
fi

case "$CHECK_TYPE" in
    health)
        HEALTH=$(usm sites health "$SITE_ID" --output json 2>/dev/null)
        if [[ $? -ne 0 ]]; then
            echo "CRITICAL: Failed to get health data"
            exit 2
        fi
        
        STATUS=$(echo "$HEALTH" | jq -r '.status // "unknown"')
        OFFLINE=$(echo "$HEALTH" | jq '.devices.offline // 0')
        
        if [[ "$STATUS" == "healthy" && "$OFFLINE" -eq 0 ]]; then
            echo "OK: Site healthy, all devices online"
            exit 0
        elif [[ "$OFFLINE" -gt 0 ]]; then
            echo "WARNING: $OFFLINE devices offline"
            exit 1
        else
            echo "CRITICAL: Site status is $STATUS"
            exit 2
        fi
        ;;
    
    devices)
        DEVICES=$(usm devices list "$SITE_ID" --status offline --output json 2>/dev/null)
        COUNT=$(echo "$DEVICES" | jq '.devices | length')
        
        if [[ "$COUNT" -eq 0 ]]; then
            echo "OK: All devices online"
            exit 0
        elif [[ "$COUNT" -lt 3 ]]; then
            echo "WARNING: $COUNT devices offline"
            exit 1
        else
            echo "CRITICAL: $COUNT devices offline"
            exit 2
        fi
        ;;
    
    clients)
        CLIENTS=$(usm clients list "$SITE_ID" --output json 2>/dev/null)
        COUNT=$(echo "$CLIENTS" | jq '.clients | length')
        
        echo "OK: $COUNT clients connected"
        exit 0
        ;;
    
    *)
        echo "UNKNOWN: Unknown check type: $CHECK_TYPE"
        exit 3
        ;;
esac
```

### Nagios Configuration

```cfg
# commands.cfg
define command {
    command_name    check_usm_health
    command_line    /usr/lib/nagios/plugins/check_usm.sh $ARG1$ health
}

define command {
    command_name    check_usm_devices
    command_line    /usr/lib/nagios/plugins/check_usm.sh $ARG1$ devices
}
```

```cfg
# services.cfg
define service {
    use                     generic-service
    host_name               unifi-site
    service_description     UniFi Site Health
    check_command           check_usm_health!60abcdef1234567890abcdef
}
```

## Zabbix Integration

### User Parameter

```bash
# /etc/zabbix/zabbix_agentd.d/usm.conf
UserParameter=usm.health[*],usm sites health $1 --output json | jq -r '.status'
UserParameter=usm.devices.online[*],usm devices list $1 --output json | jq '[.devices[] | select(.status == "online")] | length'
UserParameter=usm.devices.offline[*],usm devices list $1 --output json | jq '[.devices[] | select(.status == "offline")] | length'
UserParameter=usm.clients.total[*],usm clients list $1 --output json | jq '.clients | length'
UserParameter=usm.alerts[*],usm alerts list --site-id $1 --output json | jq '.alerts | length'
```

### Zabbix Template

Import template items:
- `usm.health[{SITE_ID}]` - Site health status
- `usm.devices.online[{SITE_ID}]` - Online device count
- `usm.devices.offline[{SITE_ID}]` - Offline device count
- `usm.clients.total[{SITE_ID}]` - Total client count
- `usm.alerts[{SITE_ID}]` - Active alert count

## Datadog Integration

### Custom Check

**File**: `usm_check.py`

```python
# usm_check.py - Datadog custom check
from checks import AgentCheck
import subprocess
import json

class USMCheck(AgentCheck):
    def check(self, instance):
        site_id = instance.get('site_id')
        
        # Get health
        try:
            result = subprocess.run(
                ['usm', 'sites', 'health', site_id, '--output', 'json'],
                capture_output=True, text=True, timeout=30
            )
            health = json.loads(result.stdout)
            
            # Send metrics
            status_val = 1 if health.get('status') == 'healthy' else 0
            self.gauge('usm.site.health', status_val, tags=[f'site:{site_id}'])
            
            devices = health.get('devices', {})
            self.gauge('usm.devices.total', devices.get('total', 0), tags=[f'site:{site_id}'])
            self.gauge('usm.devices.online', devices.get('online', 0), tags=[f'site:{site_id}'])
            self.gauge('usm.devices.offline', devices.get('offline', 0), tags=[f'site:{site_id}'])
            
        except Exception as e:
            self.service_check('usm.health', AgentCheck.CRITICAL, message=str(e))
```

## New Relic Integration

### Custom Event Script

```bash
#!/bin/bash
# newrelic-usm-events.sh

SITE_ID="$1"
API_KEY="$NEW_RELIC_API_KEY"

# Collect metrics
HEALTH=$(usm sites health "$SITE_ID" --output json)
DEVICES=$(usm devices list "$SITE_ID" --output json)
CLIENTS=$(usm clients list "$SITE_ID" --output json)

# Create event payload
cat <<EOF | curl -X POST \
  -H "Content-Type: application/json" \
  -H "Api-Key: $API_KEY" \
  https://insights-collector.newrelic.com/v1/accounts/$ACCOUNT_ID/events \
  -d @-
{
  "eventType": "UniFiSiteHealth",
  "siteId": "$SITE_ID",
  "status": $(echo "$HEALTH" | jq -r '.status'),
  "devicesOnline": $(echo "$HEALTH" | jq '.devices.online'),
  "devicesOffline": $(echo "$HEALTH" | jq '.devices.offline'),
  "totalClients": $(echo "$CLIENTS" | jq '.clients | length'),
  "timestamp": $(date +%s)
}
EOF
```

## PagerDuty Integration

### Alert Routing

```bash
#!/bin/bash
# pagerduty-alert.sh

INCIDENT_KEY="$1"
MESSAGE="$2"
SEVERITY="$3"  # critical, error, warning, info

PAGERDUTY_KEY="your-integration-key"

curl -X POST \
  -H "Content-Type: application/json" \
  -d "{
    \"routing_key\": \"$PAGERDUTY_KEY\",
    \"event_action\": \"trigger\",
    \"dedup_key\": \"$INCIDENT_KEY\",
    \"payload\": {
      \"summary\": \"$MESSAGE\",
      \"severity\": \"$SEVERITY\",
      \"source\": \"usm-monitor\"
    }
  }" \
  https://events.pagerduty.com/v2/enqueue
```

## Slack Integration

### Webhook Notifications

```bash
#!/bin/bash
# slack-notify.sh

WEBHOOK_URL="$SLACK_WEBHOOK_URL"
SITE_ID="$1"
STATUS="$2"
MESSAGE="$3"

COLOR=$([[ "$STATUS" == "ok" ]] && echo "good" || echo "danger")

curl -X POST -H 'Content-type: application/json' \
  --data "{
    \"attachments\": [{
      \"color\": \"$COLOR\",
      \"title\": \"UniFi Site Manager Alert\",
      \"fields\": [
        {\"title\": \"Site\", \"value\": \"$SITE_ID\", \"short\": true},
        {\"title\": \"Status\", \"value\": \"$STATUS\", \"short\": true}
      ],
      \"text\": \"$MESSAGE\"
    }]
  }" \
  "$WEBHOOK_URL"
```

## Log Aggregation

### Splunk Integration

```bash
# Send to Splunk HTTP Event Collector
usm sites list --output json | curl -X POST \
  -H "Authorization: Splunk your-token" \
  -d "{\"sourcetype\": \"usm\", \"event\": $(cat)}" \
  https://splunk.example.com:8088/services/collector/event
```

### ELK Stack

```bash
# Filebeat configuration
filebeat.inputs:
- type: log
  paths:
    - /var/log/usm/*.log
  fields:
    service: usm
  fields_under_root: true

# Logstash filter
filter {
  if [service] == "usm" {
    json {
      source => "message"
    }
  }
}
```

## Deployment

### Docker Compose (Full Stack)

```yaml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"
  
  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana-dashboard.json:/var/lib/grafana/dashboards/usm.json
    ports:
      - "3000:3000"
  
  usm-exporter:
    build: .
    environment:
      - USM_API_KEY=${USM_API_KEY}
      - USM_SITE_ID=${USM_SITE_ID}
      - EXPORTER_PORT=9100
    ports:
      - "9100:9100"

volumes:
  prometheus-data:
  grafana-data:
```

---

For more monitoring examples, see:
- [Basic Usage](../basic/)
- [Automation Scripts](../automation/)
- [CI/CD Integration](../ci-cd/)
