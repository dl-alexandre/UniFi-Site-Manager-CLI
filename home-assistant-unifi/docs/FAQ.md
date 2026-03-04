# Frequently Asked Questions

Common questions and answers about UniFi Controller Home Assistant integration.

## Table of Contents

- [General Questions](#general-questions)
- [Installation](#installation)
- [Configuration](#configuration)
- [Entities](#entities)
- [Device Trackers](#device-trackers)
- [Troubleshooting](#troubleshooting)
- [Automation](#automation)
- [Advanced](#advanced)

## General Questions

### What is this integration?

A custom Home Assistant integration that connects directly to your local UniFi Network Controller (UDM, Cloud Key, or standalone) to monitor and control your network.

### What's the difference from the official UniFi integration?

| Feature | This Integration | Official HA Integration |
|---------|-----------------|------------------------|
| Connection | Direct local API | UniFi OS Cloud |
| Real-time updates | WebSocket (instant) | Polling (30s delay) |
| Site Manager API | No | Yes |
| Local Controller | Yes | Limited |

### Do I need a UniFi account?

No! This integration connects directly to your local controller using a **local admin account**. No cloud account or UniFi account required.

### What controllers are supported?

- ✅ UniFi Dream Machine (UDM)
- ✅ UniFi Dream Machine Pro (UDM-Pro)
- ✅ UniFi Dream Router (UDR)
- ✅ UniFi Dream Machine SE (UDM-SE)
- ✅ Cloud Key Gen2+
- ✅ Standalone Controller 7.x+

### Is this free?

Yes! Open source (MIT license) and free to use.

## Installation

### Where do I find the integration?

**Via HACS**: Search for "UniFi Controller" in HACS Integrations after adding the custom repository.

**Manual**: Files are in `custom_components/unifi_controller/`.

### How do I update the integration?

**HACS**:
1. Go to HACS → Integrations
2. Find "UniFi Controller"
3. Click Update button
4. Restart Home Assistant

**Manual**:
1. Download latest files
2. Replace files in `custom_components/unifi_controller/`
3. Restart Home Assistant

### Can I use this with the official integration?

Yes, but it's not recommended as you'll have duplicate entities. Choose one:
- Use this for **local controllers** (UDM, etc.)
- Use official for **cloud sites**

### Do I need to restart after installation?

Yes! Always restart Home Assistant after:
- Installing the integration
- Updating the integration
- Changing configuration

## Configuration

### What credentials should I use?

Use a **local admin account** on your UniFi controller:
- Username: Any local user with Admin role
- Password: That user's password
- NOT your Ubiquiti SSO/cloud account

### How do I create a local account?

1. Log into your UniFi controller web UI
2. Settings → Admins → + Invite Admin
3. Enter:
   - Name: "Home Assistant"
   - Username: `ha_service` (or any name)
   - Password: Strong password
4. Role: Administrator
5. Save

### What's the Site ID?

The site identifier, usually `default`.

To find it:
1. Log into controller
2. Look at URL: `https://192.168.1.1/network/default/dashboard`
3. Site ID is between `/network/` and `/dashboard` → `default`

### Should I check "UniFi OS"?

**Yes** if using:
- UDM/UDR/UDM-Pro/UDM-SE
- Cloud Key Gen2+

**No** if using:
- Standalone Controller software
- Older Cloud Key

### Should I verify SSL?

**No** for most setups. UniFi uses self-signed certificates by default.

Check "Verify SSL" only if:
- You've installed a valid certificate
- You're using a reverse proxy with valid SSL

### Can I add multiple sites?

Yes! Add multiple instances:
1. Settings → Devices & Services
2. + Add Integration → UniFi Controller
3. Configure different Site ID or controller IP

## Entities

### What entities are created?

**Sensors**:
- `sensor.unifi_devices_online`
- `sensor.unifi_clients_connected`
- `sensor.unifi_cpu_usage`
- `sensor.unifi_wan_download`
- And more...

**Device Trackers**:
- One per connected client
- `device_tracker.unifi_<hostname>`

**Binary Sensors**:
- `binary_sensor.unifi_site_healthy`
- `binary_sensor.unifi_xxx_firmware_update`

### Why don't I see any entities?

Common causes:
1. Integration still loading (wait 30 seconds)
2. Wrong Site ID
3. No devices/clients in that site
4. Configuration error (check logs)

### How do I rename entities?

1. Settings → Devices & Services → Entities
2. Find entity
3. Click entity name
4. Click ⚙️ (gear icon)
5. Change "Name" field
6. Save

### Can I disable some entities?

Yes:
1. Settings → Devices & Services → Entities
2. Find entity
3. Click ⚙️
4. Toggle "Enabled" off
5. Save

### Why are entity IDs so long?

Entity IDs follow the pattern: `domain.unifi_description`

Example: `sensor.unifi_devices_online`

You can rename the **friendly name** (shown in UI) without changing the entity ID.

## Device Trackers

### What are device trackers?

Entities that track if a device is "home" (connected to network) or "not_home" (disconnected).

### Why are there so many device trackers?

The integration creates a tracker for **every client** connected to your network.

To reduce:
1. Go to Entities
2. Filter "unifi"
3. Disable unwanted trackers

### How do I track only specific devices?

Option 1: Disable unwanted in UI
Option 2: Use `device_tracker.see` service in automation
Option 3: Configuration option to filter by MAC

### Do device trackers drain battery?

No, device trackers are **receivers**, not transmitters. They just monitor WiFi connection status from the controller.

### Why is my phone showing "not_home" when connected?

Possible causes:
1. Private/MAC address feature on phone (changes MAC)
2. iOS MAC randomization
3. Device using different SSID

Solutions:
- Disable private WiFi address on phone
- Track by hostname instead of MAC
- Add multiple MAC addresses

## Troubleshooting

### "Failed to connect" error

**Check**:
1. Host IP is correct
2. Port is correct (443 for UniFi OS, 8443 for standalone)
3. Username/password are correct
4. Using local account (not SSO)
5. "UniFi OS" checkbox matches your controller
6. Controller firmware is up to date

### "Invalid credentials"

**Check**:
1. Username exists in UniFi controller
2. Password is correct
3. Account has Admin role
4. Not using Ubiquiti cloud account

### Entities show "unavailable"

**Solutions**:
1. Check integration status
2. Reload integration:
   - Settings → Integration → ⋮ → Reload
3. Check Home Assistant logs
4. Verify controller is online
5. Restart Home Assistant

### High CPU/memory usage

**Solutions**:
1. Disable unused device trackers
2. Increase scan interval:
   - Settings → Integration → Configure
   - Change "Scan Interval" to 60 or higher
3. Restart Home Assistant

### No real-time updates

**Check**:
1. Integration connected to WebSocket
2. Check logs for WebSocket errors
3. Firewall not blocking WebSocket
4. Try reloading integration

### "Unknown error occurred"

Enable debug logging:
```yaml
# configuration.yaml
logger:
  logs:
    custom_components.unifi_controller: debug
```

Then check logs:
```
Settings → System → Logs
```

## Automation

### How do I create an automation?

**UI Method**:
1. Settings → Automations & Scenes
2. Create Automation
3. Use UI to build automation

**YAML Method**:
```yaml
automation:
  - alias: "My UniFi Automation"
    trigger: ...
    action: ...
```

### How do I get notified when a device goes offline?

```yaml
automation:
  - alias: "Device Offline Alert"
    trigger:
      - platform: numeric_state
        entity_id: sensor.unifi_devices_online
        below: 5  # Your normal device count
    action:
      - service: notify.mobile_app_phone
        data:
          message: "⚠️ Only {{ states('sensor.unifi_devices_online') }} devices online!"
```

### How do I track new clients?

```yaml
automation:
  - alias: "New Client Alert"
    trigger:
      - platform: state
        entity_id: sensor.unifi_clients_connected
    condition:
      - condition: template
        value_template: "{{ trigger.to_state.state | int > trigger.from_state.state | int }}"
    action:
      - service: notify.mobile_app_phone
        data:
          message: "New device connected. Total: {{ states('sensor.unifi_clients_connected') }}"
```

### How do I block a specific client?

```yaml
automation:
  - alias: "Block Kids at Bedtime"
    trigger:
      - platform: time
        at: "21:00:00"
    action:
      - service: unifi_controller.block_client
        data:
          client_mac: "aa:bb:cc:dd:ee:ff"
          block: true
```

### How do I restart a device?

```yaml
automation:
  - alias: "Weekly AP Restart"
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
          device_mac: "11:22:33:44:55:66"
```

## Advanced

### Can I use this with Node-RED?

Yes! All entities and services work with Node-RED:
- Use entity nodes for sensors/trackers
- Use call-service nodes for actions

### Can I export data to InfluxDB?

Yes, via Home Assistant's InfluxDB integration:
```yaml
influxdb:
  host: localhost
  entities:
    - sensor.unifi_devices_online
    - sensor.unifi_clients_connected
```

### Can I access raw API data?

Use the REST integration or custom scripts:
```yaml
rest:
  - resource: http://your-ha:8123/api/states/sensor.unifi_devices_online
    sensor:
      - name: Raw UniFi Data
        value_template: "{{ value_json.state }}"
```

### How do I change update frequency?

1. Settings → Devices & Services
2. Find UniFi Controller
3. Click Configure
4. Change "Scan Interval"
5. Submit

### Can I filter what gets tracked?

Yes, options include:
- Enable/disable device trackers
- Track wired/wireless/guests separately
- Filter by MAC address

Configure via Options flow.

### Does this work with VLANs?

Yes! The integration works with any UniFi network configuration including VLANs, but doesn't directly manage VLAN settings.

### Can I manage multiple controllers?

Yes! Add multiple integration instances:
- One per controller
- One per site
- Mix of controller types

### Is there a mobile app?

No separate app. Use Home Assistant companion app to view and control your UniFi network.

### Can I backup/restore configurations?

Integration settings are part of Home Assistant's core configuration and backed up with HA backups.

### How do I completely remove the integration?

1. Settings → Devices & Services
2. Find UniFi Controller card
3. Click ⋮ → Delete
4. Restart Home Assistant
5. (Optional) Remove files from `custom_components/`

### Where can I get help?

- [GitHub Issues](https://github.com/dl-alexandre/Local-UniFi-CLI/issues)
- [Home Assistant Community Forum](https://community.home-assistant.io/)
- [UniFi Community Forums](https://community.ui.com/)

### How can I contribute?

- Report bugs on GitHub
- Suggest features
- Submit pull requests
- Help with documentation

### Is this official?

No. This is a community integration, not affiliated with Ubiquiti or Home Assistant.

---

**Tip**: Use Developer Tools → States to explore all available entities and their attributes.
