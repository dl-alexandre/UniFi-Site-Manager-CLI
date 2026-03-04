# Usage Guide

Comprehensive guide for using UniFi Controller Home Assistant integration.

## Table of Contents

- [Getting Started](#getting-started)
- [Understanding Entities](#understanding-entities)
- [Sensors](#sensors)
- [Device Trackers](#device-trackers)
- [Binary Sensors](#binary-sensors)
- [Services](#services)
- [Automations](#automations)
- [Dashboards](#dashboards)
- [Advanced Usage](#advanced-usage)

## Getting Started

### After Installation

Once configured, the integration will:
1. Connect to your UniFi controller
2. Discover all sites, devices, and clients
3. Create appropriate entities
4. Begin real-time updates via WebSocket

### Finding Your Entities

1. **Settings** → **Devices & Services**
2. Click on **"UniFi Controller"** integration
3. See all connected devices and entities

Or use the Entity Registry:
1. **Settings** → **Devices & Services**
4. Click **Entities** tab
5. Filter by "unifi"

## Understanding Entities

### Entity Naming Convention

All entities follow the pattern: `domain.unifi_description`

Examples:
- `sensor.unifi_devices_online`
- `sensor.unifi_clients_connected`
- `binary_sensor.unifi_udm_firmware_update`
- `device_tracker.unifi_john_iphone`

### Entity Attributes

Most entities have additional attributes (metadata). View by:
1. Click entity in Developer Tools
2. Or click entity in dashboard and check **Attributes** section

Example sensor attributes:
```yaml
friendly_name: UniFi Devices Online
unit_of_measurement: devices
icon: mdi:router-network
site_id: default
controller_host: 192.168.1.1
```

## Sensors

### Site Status Sensors

#### sensor.unifi_devices_online

Number of UniFi devices currently online.

- **Unit**: devices
- **Updates**: Real-time
- **Attributes**:
  - `total_devices`: Total managed devices
  - `offline_devices`: Number offline
  - `pending_devices`: Number pending adoption

**Usage**:
```yaml
automation:
  - alias: "Device Offline Alert"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_devices_online
        below: 5  # Adjust to your device count
    action:
      - service: notify.mobile_app_phone
        data:
          message: "⚠️ Only {{ states('sensor.unifi_devices_online') }} devices online"
```

#### sensor.unifi_clients_connected

Total number of connected clients (wired + wireless).

- **Unit**: clients
- **Updates**: Real-time

**Usage**:
```yaml
# In dashboard
type: gauge
entity: sensor.unifi_clients_connected
name: Connected Clients
max: 100
```

#### sensor.unifi_clients_wireless

Number of wireless (WiFi) clients.

#### sensor.unifi_clients_wired

Number of wired (Ethernet) clients.

#### sensor.unifi_guest_clients

Number of clients on guest networks.

### Performance Sensors

#### sensor.unifi_cpu_usage

Controller CPU usage percentage.

- **Unit**: %
- **Range**: 0-100

**Usage**:
```yaml
automation:
  - alias: "High CPU Alert"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_cpu_usage
        above: 90
        for: "00:05:00"
    action:
      - service: persistent_notification.create
        data:
          title: "Controller CPU High"
          message: "CPU usage at {{ states('sensor.unifi_cpu_usage') }}% for 5 minutes"
```

#### sensor.unifi_memory_usage

Controller memory usage percentage.

#### sensor.unifi_wan_download

Current WAN download bandwidth.

- **Unit**: Mbps
- **Updates**: Every 30 seconds

**Usage**:
```yaml
# In dashboard
type: history-graph
entities:
  - entity: sensor.unifi_wan_download
  - entity: sensor.unifi_wan_upload
hours_to_show: 24
```

#### sensor.unifi_wan_upload

Current WAN upload bandwidth.

#### sensor.unifi_wan_latency

WAN latency (ping time).

- **Unit**: ms

### Alert Sensors

#### sensor.unifi_alerts

Number of unacknowledged UniFi alerts.

- **Unit**: alerts
- **Updates**: Every 5 minutes

**Usage**:
```yaml
automation:
  - alias: "New UniFi Alert"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_alerts
        above: 0
    action:
      - service: persistent_notification.create
        data:
          title: "UniFi Alert"
          message: "{{ states('sensor.unifi_alerts') }} new alert(s) on your network"
```

## Device Trackers

Device trackers monitor specific network clients.

### How Device Trackers Work

1. Integration discovers clients from controller
2. Creates `device_tracker` entity for each
3. Tracks home/away status based on connection
4. Updates in real-time via WebSocket

### Entity Format

`device_tracker.unifi_<hostname>` or `device_tracker.unifi_<mac>`

Examples:
- `device_tracker.unifi_john_iphone`
- `device_tracker.unifi_aabbccdd1122`

### Device Tracker Attributes

```yaml
source_type: router
ip: 192.168.1.100
mac: AA:BB:CC:DD:EE:FF
hostname: Johns-iPhone
connection_type: wireless
ssid: Home-WiFi
signal: -45  # RSSI in dBm
ap_mac: 11:22:33:44:55:66
ap_name: Living Room AP
uptime: 3600  # seconds
rx_bytes: 104857600
 tx_bytes: 52428800
is_guest: false
blocked: false
first_seen: 2024-01-01T00:00:00
last_seen: 2024-01-15T12:30:00
```

### Important Device Trackers

Mark important devices to exclude from purge:

1. **Settings** → **People & Zones**
2. Click **Devices**
3. Find device tracker
4. Click **⋮** → **Make this device a tracker for person**

### Automation with Device Trackers

```yaml
# Track when someone arrives home
automation:
  - alias: "Welcome Home"
    trigger:
      - platform: state
        entity_id: device_tracker.unifi_john_phone
        from: 'not_home'
        to: 'home'
    action:
      - service: notify.mobile_app_john_phone
        data:
          message: "Welcome home! Your phone connected to WiFi."
```

### Guest Detection

```yaml
automation:
  - alias: "Guest WiFi Alert"
    trigger:
      - platform: state
        entity_id: sensor.unifi_guest_clients
    condition:
      - condition: template
        value_template: "{{ trigger.to_state.state | int > trigger.from_state.state | int }}"
    action:
      - service: notify.mobile_app_phone
        data:
          message: "New guest connected ({{ states('sensor.unifi_guest_clients') }} total)"
```

## Binary Sensors

Binary sensors are simple on/off states.

### sensor.unifi_site_healthy

Overall site health status.

- **On**: Site is healthy
- **Off**: Site has issues

### sensor.unifi_wan_connected

WAN internet connection status.

- **On**: Internet connected
- **Off**: Internet disconnected

### sensor.unifi_xxx_firmware_update

Firmware update availability for each device.

- **On**: Update available
- **Off**: Up to date

Example entities:
- `binary_sensor.unifi_udm_firmware_update`
- `binary_sensor.unifi_living_room_ap_firmware_update`

**Usage**:
```yaml
automation:
  - alias: "Firmware Update Available"
    trigger:
      - platform: state
        entity_id: binary_sensor.unifi_udm_firmware_update
        to: 'on'
    action:
      - service: persistent_notification.create
        data:
          title: "Firmware Update"
          message: "A firmware update is available for your UDM"
```

## Services

Services allow you to control your network from Home Assistant.

### Restart Device

Restart a UniFi device (AP, switch, gateway).

**Service**: `unifi_controller.restart_device`

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| `device_mac` | Yes | MAC address of device to restart |

**Example**:
```yaml
service: unifi_controller.restart_device
data:
  device_mac: "aa:bb:cc:dd:ee:ff"
```

**Automation Example**:
```yaml
automation:
  - alias: "Auto-restart offline AP"
    trigger:
      - platform: state
        entity_id: device_tracker.unifi_living_room_ap
        to: 'not_home'
        for: "00:10:00"
    action:
      - service: unifi_controller.restart_device
        data:
          device_mac: "11:22:33:44:55:66"
      - service: notify.mobile_app_phone
        data:
          message: "Restarted Living Room AP (was offline 10+ min)"
```

### Block Client

Block or unblock a network client.

**Service**: `unifi_controller.block_client`

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| `client_mac` | Yes | Client MAC address |
| `block` | Yes | `true` to block, `false` to unblock |

**Example - Block**:
```yaml
service: unifi_controller.block_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
  block: true
```

**Example - Unblock**:
```yaml
service: unifi_controller.block_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
  block: false
```

**Automation - Block Unknown Devices**:
```yaml
automation:
  - alias: "Block Unknown Device"
    trigger:
      - platform: event
        event_type: unifi_controller_new_client
    condition:
      - condition: template
        value_template: >
          {{ trigger.event.data.mac not in 
             state_attr('group.allowed_devices', 'entity_id') | default([]) }}
    action:
      - service: unifi_controller.block_client
        data:
          client_mac: "{{ trigger.event.data.mac }}"
          block: true
      - service: notify.mobile_app_phone
        data:
          message: "Blocked unknown device: {{ trigger.event.data.mac }}"
```

### Reconnect Client

Force a client to reconnect (kick and re-authenticate).

**Service**: `unifi_controller.reconnect_client`

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| `client_mac` | Yes | Client MAC address |

**Example**:
```yaml
service: unifi_controller.reconnect_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
```

### Set WLAN Enabled

Enable or disable a wireless network.

**Service**: `unifi_controller.set_wlan_enabled`

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| `wlan_id` | Yes | WLAN identifier |
| `enabled` | Yes | `true` or `false` |

**Example**:
```yaml
service: unifi_controller.set_wlan_enabled
data:
  wlan_id: "wl-abc123"
  enabled: false
```

## Automations

### Presence Detection

```yaml
automation:
  - alias: "Track Family Presence"
    trigger:
      - platform: state
        entity_id:
          - device_tracker.unifi_john_phone
          - device_tracker.unifi_jane_phone
        to: 'home'
    action:
      - service: input_boolean.turn_on
        target:
          entity_id: input_boolean.someone_home
```

### Network Health Monitoring

```yaml
  - alias: "Network Health Check"
    trigger:
      - platform: time_pattern
        minutes: "0"
    action:
      - choose:
          - conditions:
              - condition: state
                entity_id: binary_sensor.unifi_site_healthy
                state: 'off'
            sequence:
              - service: notify.slack
                data:
                  message: "🚨 Network unhealthy: {{ states('sensor.unifi_alerts') }} alerts"
```

### Guest Network Management

```yaml
  - alias: "Auto-enable Guest WiFi"
    trigger:
      - platform: time
        at: "09:00:00"
    condition:
      - condition: time
        weekday:
          - sat
          - sun
    action:
      - service: unifi_controller.set_wlan_enabled
        data:
          wlan_id: "wl-guest"
          enabled: true
      - service: notify.mobile_app_phone
        data:
          message: "Guest WiFi enabled for the weekend"
```

### Bandwidth Monitoring

```yaml
  - alias: "High Bandwidth Usage"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_wan_download
        above: 800  # Mbps
    action:
      - service: persistent_notification.create
        data:
          title: "High Bandwidth Usage"
          message: "Download at {{ states('sensor.unifi_wan_download') }} Mbps"
```

### Device Maintenance

```yaml
  - alias: "Weekly Device Restart"
    trigger:
      - platform: time
        at: "04:00:00"
    condition:
      - condition: time
        weekday:
          - sun
    action:
      - service: unifi_controller.restart_device
        data:
          device_mac: "11:22:33:44:55:66"  # Specific AP
      - delay: "00:05:00"
      - service: unifi_controller.restart_device
        data:
          device_mac: "aa:bb:cc:dd:ee:ff"  # Another AP
```

## Dashboards

### Network Overview Dashboard

```yaml
title: Network Status
views:
  - title: Overview
    cards:
      - type: vertical-stack
        cards:
          - type: markdown
            content: |
              ## Network Status
              {% if is_state('binary_sensor.unifi_site_healthy', 'on') %}
              ✅ **All Systems Operational**
              {% else %}
              ⚠️ **Issues Detected**
              {{ states('sensor.unifi_alerts') }} active alerts
              {% endif %}
          
      - type: entities
        title: Quick Stats
        entities:
          - sensor.unifi_devices_online
          - sensor.unifi_clients_connected
          - sensor.unifi_guest_clients
          - sensor.unifi_cpu_usage
          - sensor.unifi_memory_usage
      
      - type: gauge
        entity: sensor.unifi_wan_download
        name: Download Speed
        max: 1000
        severity:
          green: 0
          yellow: 500
          red: 800
      
      - type: gauge
        entity: sensor.unifi_wan_upload
        name: Upload Speed
        max: 1000
      
      - type: history-graph
        title: 24h Bandwidth Usage
        entities:
          - sensor.unifi_wan_download
          - sensor.unifi_wan_upload
        hours_to_show: 24
      
      - type: history-graph
        title: Client Activity
        entities:
          - sensor.unifi_clients_connected
          - sensor.unifi_devices_online
        hours_to_show: 24
```

### Device Management Card

```yaml
type: entities
title: Device Controls
entities:
  - type: button
    entity: binary_sensor.unifi_living_room_ap_firmware_update
    name: Living Room AP Update
    tap_action:
      action: call-service
      service: unifi_controller.restart_device
      service_data:
        device_mac: "11:22:33:44:55:66"
  
  - type: button
    name: Restart All APs
    tap_action:
      action: call-service
      service: unifi_controller.restart_device
      service_data:
        device_mac: "all"
```

### Client Tracker Card

```yaml
type: map
title: Connected Clients
default_zoom: 18
entities:
  - device_tracker.unifi_john_phone
  - device_tracker.unifi_jane_phone
  - device_tracker.unifi_laptop
hours_to_show: 1
theme: default
```

## Advanced Usage

### Template Sensors

Create calculated sensors:

```yaml
template:
  - sensor:
      - name: "Network Health Score"
        state: >
          {% set devices = states('sensor.unifi_devices_online') | int %}
          {% set total = states('sensor.unifi_devices_online') | int + states('sensor.unifi_devices_offline') | int %}
          {% set percent = (devices / total * 100) if total > 0 else 0 %}
          {{ percent | round(0) }}
        unit_of_measurement: "%"
      
      - name: "Total Network Usage"
        state: >
          {{ (states('sensor.unifi_wan_download') | float + 
              states('sensor.unifi_wan_upload') | float) | round(2) }}
        unit_of_measurement: "Mbps"
```

### Utility Meters

Track cumulative usage:

```yaml
utility_meter:
  daily_bandwidth:
    source: sensor.unifi_wan_download
    cycle: daily
  
  monthly_bandwidth:
    source: sensor.unifi_wan_download
    cycle: monthly
```

### InfluxDB Integration

Store historical data:

```yaml
influxdb:
  host: localhost
  port: 8086
  database: home_assistant
  entities:
    - sensor.unifi_devices_online
    - sensor.unifi_clients_connected
    - sensor.unifi_wan_download
    - sensor.unifi_wan_upload
```

### Prometheus Integration

```yaml
prometheus:
  namespace: ha
  filter:
    include_entities:
      - sensor.unifi_devices_online
      - sensor.unifi_clients_connected
```

## Tips and Best Practices

### 1. Entity Naming

Rename entities for easier identification:
1. **Settings** → **Devices & Services**
2. Click **Entities**
3. Filter "unifi"
4. Click entity
5. Click **⚙️** → Change **Name**

### 2. Disable Unused Entities

Reduce clutter by disabling unused device trackers:
1. Find entity in list
2. Click **⚙️**
3. Toggle **Enabled** off

### 3. Group Related Entities

```yaml
group:
  network_sensors:
    name: Network Sensors
    entities:
      - sensor.unifi_devices_online
      - sensor.unifi_clients_connected
      - sensor.unifi_cpu_usage
```

### 4. Use Area Cards

Assign devices to Home Assistant areas for better organization.

### 5. Monitor Changes

Enable change notifications:
```yaml
automation:
  - alias: "Entity Change Logger"
    mode: parallel
    trigger:
      - platform: state
        entity_id:
          - sensor.unifi_devices_online
          - sensor.unifi_clients_connected
    action:
      - service: logbook.log
        data:
          name: "{{ trigger.entity_id }}"
          message: "changed to {{ trigger.to_state.state }}"
```

## Troubleshooting

### Entities Not Updating

1. Check integration status:
   - **Settings** → **Devices & Services** → UniFi Controller
   - Look for errors
2. Restart integration:
   - Click **⋮** → **Reload**
3. Enable debug logging:
   ```yaml
   logger:
     logs:
       custom_components.unifi_controller: debug
   ```

### High CPU Usage

1. Reduce number of device trackers
2. Increase scan interval in configuration
3. Disable unused entities

### Missing Clients

1. Check client is connected to network
2. Verify Site ID is correct
3. Restart integration

## Next Steps

- Explore [API Documentation](API.md) for technical details
- Read [FAQ](FAQ.md) for common questions
- Check [Configuration Guide](CONFIGURATION.md) for advanced settings
- Join the community for tips and support
