# UniFi Controller for Home Assistant

[![HACS](https://img.shields.io/badge/HACS-Custom-orange.svg)](https://hacs.xyz/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Home Assistant](https://img.shields.io/badge/home%20assistant-2023.1%2B-blue.svg)](https://www.home-assistant.io/)

> Custom Home Assistant integration for monitoring and controlling UniFi Network Controllers locally.

## Features

- 📊 **Sensors**: Monitor devices online, clients connected, CPU/memory usage, WAN speeds, and alerts
- 👤 **Device Trackers**: Track connected clients with home/away status
- 🔔 **Binary Sensors**: Firmware update availability notifications
- 🎛️ **Services**: Restart devices, block/unblock clients, force reconnections
- 🖥️ **UI Configuration**: Easy setup through Home Assistant UI (no YAML required)
- 🔒 **Secure**: Direct local connection, no cloud required
- ⚡ **Real-time**: Live updates via WebSocket connection

## Demo

```yaml
# Dashboard showing network status
title: Network Dashboard
views:
  - cards:
      - type: entities
        entities:
          - sensor.unifi_devices_online
          - sensor.unifi_clients_connected
          - sensor.unifi_cpu_usage
          - sensor.unifi_wan_download
```

## Quick Start

### 1. Install via HACS

1. Add custom repository: `https://github.com/dl-alexandre/Local-UniFi-CLI`
2. Install "UniFi Controller" integration
3. Restart Home Assistant

### 2. Configure

1. Go to **Settings** → **Devices & Services**
2. Click **Add Integration** → Search "UniFi Controller"
3. Enter controller details:
   - **Host**: `192.168.1.1` (your UDM/Cloud Key IP)
   - **Port**: `443` (UniFi OS) or `8443` (standalone)
   - **Username**: Local admin username
   - **Password**: Local admin password
   - **Site ID**: Usually `default`
   - **UniFi OS**: Check for UDM/UDR/UDM-Pro
   - **Verify SSL**: Uncheck for self-signed certificates

### 3. Done!

Entities are automatically created for all your devices and clients.

## Installation

### HACS (Recommended)

```
HACS → Integrations → ⋮ → Custom repositories
→ Add: https://github.com/dl-alexandre/Local-UniFi-CLI
→ Category: Integration
→ Install
```

### Manual Installation

```bash
# Copy to custom_components
mkdir -p config/custom_components
cp -r custom_components/unifi_controller config/custom_components/

# Restart Home Assistant
```

## Configuration

### UI Configuration

All configuration is done through the UI:

| Option | Description | Default |
|--------|-------------|---------|
| Host | Controller IP/hostname | Required |
| Port | Controller port | 443/8443 |
| Username | Local admin account | Required |
| Password | Account password | Required |
| Site ID | Site identifier | default |
| UniFi OS | Enable for UDM/UDR/Cloud Key | false |
| Verify SSL | Check certificate | false |

### YAML Configuration (Legacy)

```yaml
# configuration.yaml
unifi_controller:
  host: 192.168.1.1
  port: 443
  username: admin
  password: !secret unifi_password
  site_id: default
  unifi_os: true
  verify_ssl: false
```

## Entities

### Sensors

| Entity | Description | Unit |
|--------|-------------|------|
| `sensor.unifi_devices_online` | Total devices online | devices |
| `sensor.unifi_devices_offline` | Devices offline | devices |
| `sensor.unifi_clients_connected` | Total clients connected | clients |
| `sensor.unifi_clients_wireless` | Wireless clients | clients |
| `sensor.unifi_clients_wired` | Wired clients | clients |
| `sensor.unifi_guest_clients` | Guest clients | clients |
| `sensor.unifi_cpu_usage` | Controller CPU usage | % |
| `sensor.unifi_memory_usage` | Controller memory usage | % |
| `sensor.unifi_alerts` | Active alerts | alerts |
| `sensor.unifi_wan_download` | WAN download speed | Mbps |
| `sensor.unifi_wan_upload` | WAN upload speed | Mbps |
| `sensor.unifi_wan_latency` | WAN latency | ms |

### Device Trackers

Each connected client creates a `device_tracker` entity:

- Home/away status based on connection state
- Attributes:
  - IP address
  - Hostname
  - MAC address
  - Connection type (wired/wireless)
  - SSID (for wireless clients)
  - Signal strength (RSSI)
  - Data usage (rx/tx)
  - First seen / last seen timestamps

### Binary Sensors

| Entity | Description |
|--------|-------------|
| `binary_sensor.unifi_xxx_firmware_update` | Firmware update available for each device |
| `binary_sensor.unifi_site_healthy` | Overall site health status |
| `binary_sensor.unifi_wan_connected` | WAN connection status |

## Services

### Restart Device

Restart a UniFi device (AP, switch, gateway).

```yaml
service: unifi_controller.restart_device
data:
  device_mac: "aa:bb:cc:dd:ee:ff"
```

### Block/Unblock Client

Block or unblock network clients.

```yaml
# Block a client
service: unifi_controller.block_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
  block: true

# Unblock a client
service: unifi_controller.block_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
  block: false
```

### Reconnect Client

Force a client to reconnect (kick and re-authenticate).

```yaml
service: unifi_controller.reconnect_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
```

### Enable/Disable WLAN

Turn wireless networks on/off.

```yaml
service: unifi_controller.set_wlan_enabled
data:
  wlan_id: "wl-xxxxxx"
  enabled: false
```

## Automation Examples

### Guest WiFi Welcome

```yaml
automation:
  - alias: "Guest WiFi Welcome"
    trigger:
      - platform: state
        entity_id: sensor.unifi_guest_clients
    condition:
      - condition: numeric_state
        entity_id: sensor.unifi_guest_clients
        above: 0
    action:
      - service: notify.mobile_app_phone
        data:
          title: "Guest Connected"
          message: "A guest has connected to your WiFi network"
```

### Device Offline Alert

```yaml
  - alias: "Critical Device Offline"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_devices_online
        below: 5  # Adjust to your setup
    action:
      - service: persistent_notification.create
        data:
          title: "Network Alert"
          message: "{{ states('sensor.unifi_devices_offline') }} devices are offline!"
      
      - service: notify.slack
        data:
          message: "🚨 Network devices offline: {{ states('sensor.unifi_devices_offline') }}"
```

### High Bandwidth Usage

```yaml
  - alias: "High WAN Usage"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_wan_download
        above: 500  # Mbps
    action:
      - service: notify.mobile_app_phone
        data:
          message: "High bandwidth usage detected: {{ states('sensor.unifi_wan_download') }} Mbps"
```

### Automatic Firmware Updates

```yaml
  - alias: "Notify on Firmware Update"
    trigger:
      - platform: state
        entity_id: binary_sensor.unifi_udm_firmwate_update
        to: 'on'
    action:
      - service: persistent_notification.create
        data:
          title: "Firmware Update Available"
          message: "A firmware update is available for your UniFi devices"
```

### Block Unknown Devices

```yaml
  - alias: "Block Unknown MAC"
    trigger:
      - platform: event
        event_type: unifi_controller_new_client
    condition:
      - condition: template
        value_template: >
          {{ trigger.event.data.mac not in 
             state_attr('group.allowed_devices', 'entity_id') }}
    action:
      - service: unifi_controller.block_client
        data:
          client_mac: "{{ trigger.event.data.mac }}"
          block: true
      
      - service: notify.mobile_app_phone
        data:
          title: "Unknown Device Blocked"
          message: "Device {{ trigger.event.data.mac }} was automatically blocked"
```

## Lovelace Dashboard

### Network Overview Card

```yaml
type: vertical-stack
cards:
  - type: entities
    title: Network Status
    entities:
      - entity: sensor.unifi_devices_online
        name: Devices Online
      - entity: sensor.unifi_clients_connected
        name: Connected Clients
      - entity: sensor.unifi_alerts
        name: Active Alerts
  
  - type: glance
    entities:
      - entity: sensor.unifi_cpu_usage
        name: CPU
      - entity: sensor.unifi_memory_usage
        name: Memory
      - entity: sensor.unifi_wan_latency
        name: Latency
  
  - type: history-graph
    entities:
      - entity: sensor.unifi_clients_connected
      - entity: sensor.unifi_devices_online
    hours_to_show: 24
```

### Device Status Card

```yaml
type: conditional
conditions:
  - entity: sensor.unifi_devices_offline
    state: "0"
card:
  type: markdown
  content: |
    ## ✅ All Systems Operational
    
    All {{ states('sensor.unifi_devices_online') }} devices are online.
```

### Client Map

```yaml
type: map
entities:
  - entity: device_tracker.john_phone
  - entity: device_tracker.jane_phone
  - entity: device_tracker.laptop
hours_to_show: 24
default_zoom: 18
theme: default
```

## Troubleshooting

### Cannot Connect

**Symptoms**: "Failed to connect" during setup

**Solutions**:
1. Verify host IP/hostname is correct
2. Confirm using **local admin account**, not Ubiquiti cloud account
3. For UniFi OS (UDM/UDR): Check "UniFi OS" option
4. Try toggling "Verify SSL" for self-signed certificates
5. Check firewall rules allow connection on port 443/8443

### No Data Appearing

**Symptoms**: Entities created but show "unavailable"

**Solutions**:
1. Check Home Assistant logs for errors
2. Verify Site ID is correct (usually "default")
3. Confirm user account has read permissions
4. Restart integration: Settings → Integrations → UniFi Controller → Reload

### Authentication Issues

**Symptoms**: "Invalid credentials" or "Permission denied"

**Solutions**:
1. Create a **local account** on UniFi controller (not SSO)
2. Verify account has Administrator role
3. Check password for special characters (may need escaping)

### High CPU/Memory Usage

**Symptoms**: Home Assistant slow after adding integration

**Solutions**:
1. Reduce update frequency in configuration
2. Limit number of device trackers
3. Disable unused entities
4. Check for duplicate integrations

### Entity Names

Entity naming convention:
- Sensors: `sensor.unifi_*`
- Device trackers: `device_tracker.unifi_*`
- Binary sensors: `binary_sensor.unifi_*`

## Advanced Configuration

### Options Flow

After initial setup, you can reconfigure:
1. Settings → Devices & Services
2. Find "UniFi Controller"
3. Click **Configure**
4. Adjust:
   - Update interval
   - Entity types to create
   - Device tracker settings

### Disable Specific Entities

```yaml
# configuration.yaml (if using YAML mode)
homeassistant:
  customize:
    sensor.unifi_some_sensor:
      entity_registry_visible_default: false
```

### Manual Polling

```yaml
# Trigger update manually
automation:
  - alias: "Manual UniFi Update"
    trigger:
      - platform: time_pattern
        minutes: "/5"
    action:
      - service: homeassistant.update_entity
        target:
          entity_id:
            - sensor.unifi_devices_online
            - sensor.unifi_clients_connected
```

## API Reference

### Endpoints Used

**UniFi OS**:
- Authentication: `POST /api/auth/login`
- Sites: `GET /proxy/network/api/s/{site}/self`
- Devices: `GET /proxy/network/api/s/{site}/stat/device`
- Clients: `GET /proxy/network/api/s/{site}/stat/sta`
- Health: `GET /proxy/network/api/s/{site}/stat/health`
- Events: `WS wss://{host}/proxy/network/wss/s/{site}/events`

**Standalone Controller**:
- Similar endpoints without `/proxy/network` prefix
- Port 8443 instead of 443

### Event Types

The integration listens for real-time events:
- Device connected/disconnected
- Client connected/disconnected
- Firmware updates available
- Alerts and warnings
- Configuration changes

## Compatible Devices

Tested with:
- ✅ UniFi Dream Machine (UDM)
- ✅ UniFi Dream Machine Pro (UDM-Pro)
- ✅ UniFi Dream Router (UDR)
- ✅ UniFi Dream Machine SE (UDM-SE)
- ✅ UniFi Cloud Key Gen2 Plus
- ✅ Standalone UniFi Network Controller 7.x+

Should work with all UniFi OS devices and standalone controllers.

## Troubleshooting Guide

### Debug Logging

Enable debug logging:
```yaml
# configuration.yaml
logger:
  logs:
    custom_components.unifi_controller: debug
```

### Diagnostic Information

Settings → Devices & Services → UniFi Controller → ⋮ → Download diagnostics

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "Cannot connect" | Network/Auth issue | Check host, credentials, UniFi OS setting |
| "Invalid auth" | Wrong credentials | Use local account, verify password |
| "Timeout" | Controller slow | Increase timeout, check controller load |
| "SSL error" | Certificate issue | Disable SSL verification |
| "Site not found" | Wrong site ID | Use "default" or check controller |

## Contributing

Contributions welcome! See [CONTRIBUTING.md](../CONTRIBUTING.md)

### Development Setup

```bash
git clone https://github.com/dl-alexandre/Local-UniFi-CLI.git
cd Local-UniFi-CLI/home-assistant-unifi

# Copy to dev container
# Or use Home Assistant dev environment
```

### Testing

```bash
# Run tests
pytest tests/

# Run with coverage
pytest --cov=custom_components/unifi_controller tests/
```

## Support

- GitHub Issues: [github.com/dl-alexandre/Local-UniFi-CLI/issues](https://github.com/dl-alexandre/Local-UniFi-CLI/issues)
- Home Assistant Community Forum
- UniFi Community Forums

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built by [Devin Alexandre](https://github.com/dl-alexandre)
- Based on UniFi API documentation
- Inspired by the official UniFi integration

---

**Disclaimer**: This is not an official Ubiquiti product. UniFi is a trademark of Ubiquiti Inc.
