# UniFi Controller for Home Assistant

A custom integration for Home Assistant that connects to your local UniFi Network Controller and exposes devices, clients, and network statistics as sensors and device trackers.

## Features

- **Sensors**: Monitor devices online, clients connected, CPU/memory usage, WAN speeds, and alerts
- **Device Trackers**: Track all connected clients with home/away status
- **Binary Sensors**: Firmware update availability notifications
- **Services**: Restart devices, block/unblock clients, and force client reconnections
- **UI Configuration**: Easy setup through Home Assistant UI (no YAML required)

## Requirements

- Home Assistant 2023.1.0 or newer
- UniFi Network Controller (UDM, UDM-Pro, UDR, UDM-SE, Cloud Key, or standalone)
- Local administrator account credentials

## Installation

### HACS (Recommended)

1. Add this repository as a custom repository in HACS:
   - Go to HACS → Integrations → ⋮ (menu) → Custom repositories
   - Add URL: `https://github.com/dl-alexandre/Local-UniFi-CLI`
   - Category: Integration

2. Search for "UniFi Controller" in HACS and install

3. Restart Home Assistant

### Manual Installation

1. Copy the `custom_components/unifi_controller` directory to your Home Assistant's `custom_components` folder:
   ```
   config/custom_components/unifi_controller/
   ```

2. Restart Home Assistant

## Configuration

1. Go to **Settings** → **Devices & Services**

2. Click **Add Integration** and search for "UniFi Controller"

3. Enter your controller details:
   - **Host**: IP address or hostname (e.g., `192.168.1.1` or `unifi.local`)
   - **Port**: Usually `443` for UniFi OS, `8443` for standalone
   - **Username**: Local admin username
   - **Password**: Local admin password
   - **Site ID**: Usually `default` (leave as default for most setups)
   - **UniFi OS**: Check for UDM, UDM-Pro, UDR, UDM-SE; uncheck for standalone controllers
   - **Verify SSL**: Leave unchecked for self-signed certificates

4. Click **Submit** - the integration will test the connection and discover your sites

## Entities Created

### Sensors

| Entity | Description |
|--------|-------------|
| `sensor.unifi_devices_online` | Total UniFi devices online |
| `sensor.unifi_clients_connected` | Total clients connected |
| `sensor.unifi_alerts` | Number of unacknowledged alerts |
| `sensor.unifi_cpu_usage` | CPU usage percentage (if available) |
| `sensor.unifi_memory_usage` | Memory usage percentage (if available) |
| `sensor.unifi_wan_upload` | WAN upload speed in Mbps |
| `sensor.unifi_wan_download` | WAN download speed in Mbps |

### Device Trackers

Each connected client creates a `device_tracker` entity:
- Home/away status based on connection
- Attributes: IP address, hostname, connection type (wired/wireless), SSID, signal strength, data usage

### Binary Sensors

| Entity | Description |
|--------|-------------|
| `binary_sensor.unifi_*_firmware_update` | Firmware update available for each device |

## Services

### `unifi_controller.restart_device`

Restart a UniFi device (AP, switch, etc.) by MAC address.

```yaml
service: unifi_controller.restart_device
data:
  device_mac: "aa:bb:cc:dd:ee:ff"
```

### `unifi_controller.block_client`

Block or unblock a network client.

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

### `unifi_controller.reconnect_client`

Force a client to reconnect (kick them off temporarily).

```yaml
service: unifi_controller.reconnect_client
data:
  client_mac: "aa:bb:cc:dd:ee:ff"
```

## Troubleshooting

### Cannot Connect

1. Verify the host IP/hostname is correct
2. Ensure you're using a **local** admin account, not a Ubiquiti cloud account
3. For UniFi OS devices (UDM, etc.), make sure "UniFi OS" is checked
4. Try toggling "Verify SSL" if using self-signed certificates

### No Data Appearing

1. Check Home Assistant logs for errors
2. Verify the Site ID is correct (usually "default")
3. Ensure the user account has read permissions

### Authentication Issues

- The integration requires a **local account** on your UniFi Controller
- Cloud-only accounts (SSO) won't work - create a local admin account in UniFi settings

## Compatible Devices

Tested with:
- UniFi Dream Machine (UDM)
- UniFi Dream Machine Pro (UDM-Pro)
- UniFi Dream Router (UDR)
- UniFi Cloud Key Gen2 Plus
- Standalone UniFi Network Controller

Should work with all UniFi OS devices and standalone controllers running UniFi Network 7.x+

## Development

This integration is based on the API patterns from the [Local UniFi CLI](https://github.com/dl-alexandre/Local-UniFi-CLI) project.

### API Endpoints Used

The integration uses the local UniFi API endpoints:
- UniFi OS: `/proxy/network/api/s/{site}/...`
- Standalone: `/api/s/{site}/...`

## License

MIT License - See LICENSE file for details

## Support

For issues, feature requests, or questions:
- GitHub Issues: [github.com/dl-alexandre/Local-UniFi-CLI/issues](https://github.com/dl-alexandre/Local-UniFi-CLI/issues)

## Credits

- Built by [Devin Alexander](https://github.com/dl-alexandre)
- Based on the UniFi API documentation and the Local UniFi CLI project
