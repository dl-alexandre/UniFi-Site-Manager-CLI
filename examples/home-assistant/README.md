# Home Assistant Integration

Integrate UniFi Site Manager CLI with Home Assistant for network monitoring and control.

## Overview

This integration allows you to:
- Monitor site health and device status
- Track connected clients
- Receive alerts on network issues
- Control devices from Home Assistant
- Create automations based on network events

## Prerequisites

- Home Assistant 2023.1.0 or newer
- UniFi Site Manager CLI installed and configured
- API key from unifi.ui.com

## Installation

### Method 1: Command Line Sensor

Add to `configuration.yaml`:

```yaml
# configuration.yaml
sensor:
  - platform: command_line
    name: UniFi Site Health
    command: 'usm sites health YOUR_SITE_ID --output json 2>/dev/null || echo "{}"'
    value_template: '{{ value_json.status | default("unknown") }}'
    scan_interval: 300  # 5 minutes
    json_attributes:
      - devices
      - wan
      - wlan

  - platform: command_line
    name: UniFi Devices Online
    command: 'usm devices list YOUR_SITE_ID --status online --output json 2>/dev/null || echo "{}"'
    value_template: '{{ value_json.devices | default([]) | length }}'
    unit_of_measurement: 'devices'
    scan_interval: 300

  - platform: command_line
    name: UniFi Clients Connected
    command: 'usm clients list YOUR_SITE_ID --output json 2>/dev/null || echo "{}"'
    value_template: '{{ value_json.clients | default([]) | length }}'
    unit_of_measurement: 'clients'
    scan_interval: 60

  - platform: command_line
    name: UniFi Alerts
    command: 'usm alerts list --site-id YOUR_SITE_ID --output json 2>/dev/null || echo "{}"'
    value_template: '{{ value_json.alerts | default([]) | length }}'
    unit_of_measurement: 'alerts'
    scan_interval: 300
```

### Method 2: RESTful Sensor

```yaml
# This requires setting up a proxy/middleware
# that converts USM CLI output to REST API

rest:
  - resource: http://localhost:5000/api/health
    scan_interval: 300
    sensor:
      - name: UniFi Site Status
        value_template: '{{ value_json.status }}'
```

### Method 3: Shell Command Integration

```yaml
# configuration.yaml
shell_command:
  unifi_restart_device: 'usm devices restart {{ site_id }} {{ device_id }}'
  unifi_block_client: 'usm clients block {{ site_id }} "{{ client_mac }}"'
  unifi_unblock_client: 'usm clients unblock {{ site_id }} "{{ client_mac }}"'
```

## Entity Examples

### Binary Sensors

```yaml
binary_sensor:
  - platform: template
    sensors:
      unifi_site_healthy:
        friendly_name: "UniFi Site Healthy"
        value_template: '{{ states("sensor.unifi_site_health") == "healthy" }}'
        device_class: connectivity
      
      unifi_has_alerts:
        friendly_name: "UniFi Has Alerts"
        value_template: '{{ states("sensor.unifi_alerts") | int > 0 }}'
        device_class: problem
```

### Device Trackers

```yaml
# Create device trackers for important clients
# This requires a script that runs periodically
```

### Automations

**Alert on Device Offline**:
```yaml
automation:
  - alias: "UniFi Device Offline Alert"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_devices_online
        below: 10  # Adjust based on your setup
    condition:
      - condition: state
        entity_id: binary_sensor.unifi_site_healthy
        state: 'off'
    action:
      - service: notify.mobile_app_phone
        data:
          title: "Network Alert"
          message: "{{ states('sensor.unifi_devices_online') }} devices online. Check UniFi site."
      
      - service: persistent_notification.create
        data:
          title: "UniFi Alert"
          message: "Site health degraded"
          notification_id: unifi_alert
```

**Client Count Monitoring**:
```yaml
  - alias: "High Client Count Warning"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_clients_connected
        above: 100  # Adjust threshold
    action:
      - service: notify.slack
        data:
          message: "High client count: {{ states('sensor.unifi_clients_connected') }} connected"
```

**Automated Reboot**:
```yaml
  - alias: "Auto Reboot Offline Device"
    trigger:
      - platform: time_pattern
        minutes: "/30"  # Check every 30 minutes
    condition:
      - condition: template
        value_template: >
          {% set health = states.sensor.unifi_site_health.attributes.devices %}
          {{ health.offline | default(0) > 0 }}
    action:
      - service: shell_command.unifi_restart_device
        data:
          site_id: "YOUR_SITE_ID"
          device_id: "DEVICE_ID_HERE"
```

## Lovelace Dashboard

Create `ui-lovelace.yaml`:

```yaml
title: Network Dashboard
views:
  - title: UniFi Overview
    cards:
      - type: entities
        title: Site Health
        entities:
          - entity: sensor.unifi_site_health
            name: Status
          - entity: sensor.unifi_devices_online
            name: Devices Online
          - entity: sensor.unifi_clients_connected
            name: Clients
          - entity: sensor.unifi_alerts
            name: Active Alerts
      
      - type: gauge
        entity: sensor.unifi_devices_online
        name: Devices Online
        max: 20  # Adjust to your device count
        severity:
          green: 15
          yellow: 10
          red: 5
      
      - type: gauge
        entity: sensor.unifi_clients_connected
        name: Connected Clients
        max: 100  # Adjust to your capacity
      
      - type: conditional
        conditions:
          - entity: binary_sensor.unifi_has_alerts
            state: 'on'
        card:
          type: markdown
          content: |
            ## ⚠️ Network Alerts
            
            {{ states.sensor.unifi_alerts.state }} active alerts require attention.
            
            Check UniFi dashboard for details.
```

## Advanced Integration

### Custom Component (Python)

Create custom component for deeper integration:

**File**: `custom_components/unifi_usm/sensor.py`

```python
"""UniFi Site Manager integration for Home Assistant."""
import subprocess
import json
from homeassistant.components.sensor import SensorEntity
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.config_entries import ConfigEntry

DOMAIN = "unifi_usm"

async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up USM sensors."""
    site_id = entry.data["site_id"]
    api_key = entry.data["api_key"]
    
    sensors = [
        USMSiteHealthSensor(site_id, api_key),
        USMDeviceCountSensor(site_id, api_key),
        USMClientCountSensor(site_id, api_key),
    ]
    
    async_add_entities(sensors, update_before_add=True)


class USMBaseSensor(SensorEntity):
    """Base class for USM sensors."""
    
    def __init__(self, site_id: str, api_key: str):
        """Initialize the sensor."""
        self._site_id = site_id
        self._api_key = api_key
        self._attr_device_info = {
            "identifiers": {(DOMAIN, site_id)},
            "name": f"UniFi Site {site_id}",
            "manufacturer": "Ubiquiti",
        }


class USMSiteHealthSensor(USMBaseSensor):
    """Site health sensor."""
    
    def __init__(self, site_id: str, api_key: str):
        super().__init__(site_id, api_key)
        self._attr_name = "Site Health"
        self._attr_unique_id = f"{site_id}_health"
    
    def update(self):
        """Update sensor state."""
        result = subprocess.run(
            ["usm", "sites", "health", self._site_id, "--output", "json"],
            capture_output=True,
            text=True,
            env={**os.environ, "USM_API_KEY": self._api_key}
        )
        
        if result.returncode == 0:
            data = json.loads(result.stdout)
            self._attr_native_value = data.get("status", "unknown")
            self._attr_extra_state_attributes = {
                "devices_online": data.get("devices", {}).get("online", 0),
                "devices_offline": data.get("devices", {}).get("offline", 0),
                "wan_status": data.get("wan", {}).get("status", "unknown"),
            }
```

### MQTT Bridge

Publish USM data to MQTT for Home Assistant:

```bash
#!/bin/bash
# usm-mqtt-bridge.sh

SITE_ID="YOUR_SITE_ID"
MQTT_HOST="homeassistant.local"
MQTT_USER="mqtt-user"
MQTT_PASS="mqtt-pass"

while true; do
    # Get health
    HEALTH=$(usm sites health $SITE_ID --output json)
    mosquitto_pub -h $MQTT_HOST -u $MQTT_USER -P $MQTT_PASS \
        -t "unifi/health" -m "$HEALTH"
    
    # Get clients
    CLIENTS=$(usm clients list $SITE_ID --output json)
    mosquitto_pub -h $MQTT_HOST -u $MQTT_USER -P $MQTT_PASS \
        -t "unifi/clients" -m "$CLIENTS"
    
    sleep 60
done
```

Home Assistant MQTT configuration:
```yaml
mqtt:
  sensor:
    - name: "UniFi MQTT Health"
      state_topic: "unifi/health"
      value_template: "{{ value_json.status }}"
      json_attributes_topic: "unifi/health"
      json_attributes_template: "{{ value_json | tojson }}"
```

## Security Considerations

### API Key Storage

**Don't** store API keys in configuration files:

```yaml
# ❌ Bad
sensor:
  - platform: command_line
    command: 'USM_API_KEY=hardcoded-key usm sites list'
```

**Do** use Home Assistant secrets:

```yaml
# ❌ Bad
sensor:
  - platform: command_line
    command: 'usm sites list'
    # API key in environment or secrets.yaml
```

```yaml
# secrets.yaml
usm_api_key: your-secret-api-key

# configuration.yaml
sensor:
  - platform: command_line
    name: UniFi Health
    command: 'usm sites health SITE_ID'
    # USM_API_KEY from environment or pre-configured
```

### File Permissions

```bash
chmod 600 /config/secrets.yaml
```

## Troubleshooting

### Command Not Found

```bash
# Ensure usm is in PATH for Home Assistant
which usm
# If not found, specify full path:
command: '/usr/local/bin/usm sites list'
```

### Empty Results

```bash
# Test manually in Home Assistant container
docker exec -it homeassistant /bin/bash
usm sites list
```

### Performance Issues

```bash
# Reduce scan frequency
scan_interval: 600  # 10 minutes instead of default 30 seconds
```

## Additional Resources

- [Home Assistant Command Line Sensor](https://www.home-assistant.io/integrations/sensor.command_line/)
- [Home Assistant RESTful Sensor](https://www.home-assistant.io/integrations/sensor.rest/)
- [Shell Command Integration](https://www.home-assistant.io/integrations/shell_command/)
- [USM CLI Documentation](../../docs/)

## Example Automations

### Network Presence Detection

```yaml
automation:
  - alias: "Track Phone Presence"
    trigger:
      - platform: time_pattern
        minutes: "/2"  # Check every 2 minutes
    action:
      - service: python_script.unifi_presence
        data:
          site_id: "YOUR_SITE_ID"
          client_mac: "aa:bb:cc:dd:ee:ff"
          device_tracker_id: "device_tracker.phone"
```

### Bandwidth Monitoring

```yaml
  - alias: "Bandwidth Alert"
    trigger:
      - platform: time_pattern
        minutes: "0"
    action:
      - service: notify.slack
        data:
          message: >
            Daily Stats:
            - Clients: {{ states('sensor.unifi_clients_connected') }}
            - Devices: {{ states('sensor.unifi_devices_online') }}/20
```

### Guest Network Automation

```yaml
  - alias: "Enable Guest WiFi for Events"
    trigger:
      - platform: calendar
        event_id: "meeting_room"
    action:
      - service: shell_command.unifi_enable_guest
        data:
          site_id: "YOUR_SITE_ID"
          wlan_id: "GUEST_WLAN_ID"
```

---

For more integration examples, see:
- [Basic Usage](../basic/)
- [Automation Scripts](../automation/)
- [Monitoring Setup](../monitoring/)
