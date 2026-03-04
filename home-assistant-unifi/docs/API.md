# API Documentation

Technical documentation for UniFi Controller Home Assistant integration.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Data Flow](#data-flow)
- [Entity Types](#entity-types)
- [API Endpoints](#api-endpoints)
- [WebSocket Events](#websocket-events)
- [Services](#services)
- [Events](#events)
- [Development](#development)

## Architecture Overview

### Components

```
┌─────────────────────────────────────────────────────────┐
│                  Home Assistant Core                    │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │
│  │   Sensors    │  │    Device    │  │   Binary    │  │
│  │              │  │  Trackers    │  │   Sensors   │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘  │
│         │                 │                  │         │
│         └─────────────────┼──────────────────┘         │
│                          │                             │
│  ┌───────────────────────┴───────────────────────────┐  │
│  │         UniFi Controller Integration               │  │
│  │                                                    │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │  │
│  │  │  API Client  │  │  Coordinator │  │  Config  │ │  │
│  │  │              │  │              │  │   Flow   │ │  │
│  │  └──────┬───────┘  └──────────────┘  └──────────┘ │  │
│  │         │                                          │  │
│  └─────────┼──────────────────────────────────────────┘  │
└────────────┼──────────────────────────────────────────────┘
             │
             │ HTTP(s) / WebSocket
             ▼
┌─────────────────────────────────────────────────────────┐
│               UniFi Controller                          │
│                                                          │
│  • UDM/UDR/UDM-Pro/UDM-SE                              │
│  • Cloud Key Gen2+                                     │
│  • Standalone Controller                               │
└─────────────────────────────────────────────────────────┘
```

### Key Components

1. **API Client**: Handles HTTP/WebSocket communication with controller
2. **Data Update Coordinator**: Manages data fetching and caching
3. **Entity Platform**: Creates Home Assistant entities
4. **Config Flow**: UI-based configuration

## Data Flow

### Initialization Flow

```
1. User configures integration
2. Config flow validates credentials
3. API client connects to controller
4. Discovers sites, devices, clients
5. Creates appropriate entities
6. Starts WebSocket for real-time updates
```

### Update Flow

```
1. Coordinator triggers update (every 30s default)
2. API client fetches data from controller
3. Data parsed and entities updated
4. WebSocket receives real-time events
5. Entities updated immediately on events
```

### Entity Lifecycle

```
Entity Creation:
├─ Platform setup called
├─ API client fetches initial data
├─ Entities created from data
└─ Entities added to HA registry

Entity Updates:
├─ Coordinator updates (polling)
├─ WebSocket events (real-time)
├─ Entity state updated
└─ HA UI reflects changes
```

## Entity Types

### Sensor Entities

**Class**: `SensorEntity`

**Created For**:
- Device counts (online, offline, total)
- Client counts (total, wired, wireless, guest)
- Performance metrics (CPU, memory, bandwidth)
- Alert counts

**Update Method**: DataUpdateCoordinator polling + WebSocket events

**Attributes** (example - device sensor):
```python
{
    "site_id": "default",
    "controller_host": "192.168.1.1",
    "total_devices": 12,
    "offline_devices": 1,
    "pending_devices": 0,
}
```

### Device Tracker Entities

**Class**: `ScannerEntity`

**Created For**:
- Each connected client
- Automatic discovery of new clients

**Update Method**: WebSocket events (real-time)

**Attributes**:
```python
{
    "ip": "192.168.1.100",
    "mac": "AA:BB:CC:DD:EE:FF",
    "hostname": "Johns-iPhone",
    "connection_type": "wireless",  # or "wired"
    "ssid": "Home-WiFi",
    "signal": -45,  # RSSI
    "ap_mac": "11:22:33:44:55:66",
    "ap_name": "Living Room AP",
    "uptime": 3600,
    "rx_bytes": 104857600,
    "tx_bytes": 52428800,
    "is_guest": False,
    "blocked": False,
    "first_seen": "2024-01-01T00:00:00",
    "last_seen": "2024-01-15T12:30:00",
}
```

### Binary Sensor Entities

**Class**: `BinarySensorEntity`

**Created For**:
- Site health status
- WAN connectivity
- Firmware update availability (per device)

**Update Method**: DataUpdateCoordinator + WebSocket events

## API Endpoints

### UniFi OS Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth/login` | POST | Authenticate and get session |
| `/proxy/network/api/s/{site}/self` | GET | Site information |
| `/proxy/network/api/s/{site}/stat/device` | GET | Device list |
| `/proxy/network/api/s/{site}/stat/sta` | GET | Client list |
| `/proxy/network/api/s/{site}/stat/health` | GET | Health status |
| `/proxy/network/api/s/{site}/rest/wlanconf` | GET | WLAN configurations |
| `/proxy/network/api/s/{site}/rest/networkconf` | GET | Network configurations |
| `/proxy/network/api/s/{site}/stat/alarm` | GET | Active alarms |
| `/proxy/network/api/s/{site}/stat/event` | GET | Recent events |
| `/proxy/network/wss/s/{site}/events` | WS | WebSocket events |

### Standalone Controller Endpoints

Same endpoints without `/proxy/network` prefix, using port 8443.

### Authentication

**Login Request**:
```json
POST /api/auth/login
{
    "username": "admin",
    "password": "secret"
}
```

**Response**:
```json
{
    "unique_id": "abc123",
    "csrf_token": "xyz789",
    "firstname": "Admin",
    "lastname": "User",
    "full_name": "Admin User",
    "email": "admin@example.com"
}
```

**Subsequent Requests**:
```
Headers:
    Cookie: unifises=session_token
    X-CSRF-Token: xyz789
```

### Data Structures

#### Device Object

```json
{
    "_id": "60abcdef1234567890abcdef",
    "mac": "aa:bb:cc:dd:ee:ff",
    "ip": "192.168.1.10",
    "name": "Living Room AP",
    "model": "U6-Pro",
    "type": "uap",
    "version": "6.2.0",
    "state": 1,
    "adopted": true,
    "site_id": "60fedcba0987654321fedcba",
    "upgrade_available": false,
    "device_type": "uap",
    "port_table": [...],
    "uplink": {...}
}
```

#### Client Object

```json
{
    "_id": "60abcdef1234567890abcdef",
    "mac": "aa:bb:cc:dd:ee:ff",
    "ip": "192.168.1.100",
    "hostname": "Johns-iPhone",
    "name": "John's iPhone",
    "is_guest": false,
    "is_wired": false,
    "ssid": "Home-WiFi",
    "ap_mac": "11:22:33:44:55:66",
    "channel": 36,
    "rssi": -45,
    "signal": -45,
    "noise": -90,
    "tx_rate": 866700,
    "rx_rate": 780000,
    "uptime": 3600,
    "tx_bytes": 52428800,
    "rx_bytes": 104857600,
    "first_seen": 1700000000,
    "last_seen": 1700003600,
    "blocked": false
}
```

#### Health Object

```json
{
    "subsystem": "wlan",
    "status": "ok",
    "num_adopted": 4,
    "num_ap": 4,
    "num_disabled": 0,
    "num_disconnected": 0,
    "num_pending": 0,
    "num_gw": 1,
    "num_sw": 2,
    "num_adopted_gw": 1,
    "num_adopted_sw": 2,
    "gw_mac": "aa:bb:cc:dd:ee:ff",
    "gw_ip": "192.168.1.1",
    "gw_version": "1.12.0"
}
```

## WebSocket Events

### Event Types

| Event | Description |
|-------|-------------|
| `device:sync` | Device state changed |
| `client:sync` | Client state changed |
| `sta:sync` | Station (client) update |
| `alert:add` | New alert created |
| `alert:sync` | Alert updated |
| `wlanconf:add` | New WLAN created |
| `wlanconf:sync` | WLAN updated |
| `event` | Generic event |
| `message` | Controller message |

### Event Format

```json
{
    "meta": {
        "rc": "ok",
        "message": "device:sync"
    },
    "data": [
        {
            "_id": "60abcdef1234567890abcdef",
            "state": 1,
            "adopted": true
        }
    ]
}
```

### Handling Events

```python
async def _process_websocket_message(self, message: dict) -> None:
    """Process WebSocket message."""
    meta = message.get("meta", {})
    msg_type = meta.get("message", "")
    
    if msg_type == "device:sync":
        await self._update_devices(message.get("data", []))
    elif msg_type == "sta:sync":
        await self._update_clients(message.get("data", []))
    elif msg_type == "alert:add":
        await self._handle_new_alert(message.get("data"))
```

## Services

### Service Definitions

```python
async def async_setup_services(hass: HomeAssistant, controller) -> None:
    """Set up integration services."""
    
    async def restart_device(call: ServiceCall) -> None:
        """Restart a device."""
        device_mac = call.data.get("device_mac")
        await controller.api.restart_device(device_mac)
    
    hass.services.async_register(
        DOMAIN, "restart_device", restart_device,
        schema=vol.Schema({
            vol.Required("device_mac"): cv.string,
        })
    )
```

### Available Services

| Service | Schema | Description |
|---------|--------|-------------|
| `restart_device` | `{device_mac: string}` | Restart device |
| `block_client` | `{client_mac: string, block: boolean}` | Block/unblock client |
| `reconnect_client` | `{client_mac: string}` | Force client reconnect |
| `set_wlan_enabled` | `{wlan_id: string, enabled: boolean}` | Enable/disable WLAN |

## Events

### Integration Events

The integration fires custom events for automations:

| Event | Data | Description |
|-------|------|-------------|
| `unifi_controller_new_client` | `{mac, hostname, ip}` | New client connected |
| `unifi_controller_client_left` | `{mac, hostname}` | Client disconnected |
| `unifi_controller_device_offline` | `{mac, name}` | Device went offline |
| `unifi_controller_device_online` | `{mac, name}` | Device came online |
| `unifi_controller_new_alert` | `{id, message, severity}` | New alert |

### Firing Events

```python
self.hass.bus.fire(
    "unifi_controller_new_client",
    {
        "mac": client.mac,
        "hostname": client.hostname,
        "ip": client.ip,
    }
)
```

### Listening to Events

```yaml
automation:
  - alias: "New Client Alert"
    trigger:
      - platform: event
        event_type: unifi_controller_new_client
    action:
      - service: notify.mobile_app_phone
        data:
          message: "New device: {{ trigger.event.data.hostname }}"
```

## Development

### Adding New Entity Types

1. Create platform file (e.g., `switch.py`):

```python
async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up UniFi switch entities."""
    controller = hass.data[DOMAIN][entry.entry_id]
    
    switches = []
    for device in controller.api.devices.values():
        switches.append(UniFiSwitch(device, controller))
    
    async_add_entities(switches)
```

2. Implement entity class:

```python
class UniFiSwitch(SwitchEntity):
    """UniFi Device Switch."""
    
    def __init__(self, device, controller):
        self._device = device
        self._controller = controller
    
    @property
    def is_on(self) -> bool:
        return self._device.state == 1
    
    async def async_turn_on(self, **kwargs):
        await self._controller.api.enable_device(self._device.mac)
    
    async def async_turn_off(self, **kwargs):
        await self._controller.api.disable_device(self._device.mac)
```

### API Client Methods

```python
class UniFiControllerAPI:
    """API client for UniFi Controller."""
    
    async def get_devices(self) -> list[Device]:
        """Get all devices."""
        response = await self._request("get", f"/api/s/{self.site_id}/stat/device")
        return [Device(d) for d in response.get("data", [])]
    
    async def restart_device(self, mac: str) -> None:
        """Restart a device."""
        await self._request(
            "post",
            f"/api/s/{self.site_id}/cmd/devmgr",
            json={"cmd": "restart", "mac": mac.lower()}
        )
    
    async def block_client(self, mac: str, block: bool) -> None:
        """Block or unblock a client."""
        await self._request(
            "post",
            f"/api/s/{self.site_id}/cmd/stamgr",
            json={"cmd": "block-sta" if block else "unblock-sta", "mac": mac.lower()}
        )
```

### Testing

```python
# test_api.py
@pytest.fixture
def mock_api():
    """Create mock API client."""
    api = Mock(spec=UniFiControllerAPI)
    api.get_devices = AsyncMock(return_value=[mock_device])
    return api

async def test_device_sensor(mock_api):
    """Test device sensor."""
    sensor = UniFiDeviceSensor(mock_api)
    await sensor.async_update()
    
    assert sensor.state == 5  # 5 devices online
    assert sensor.extra_state_attributes["offline_devices"] == 1
```

## Configuration Schema

```python
CONFIG_SCHEMA = vol.Schema(
    {
        DOMAIN: vol.Schema(
            {
                vol.Required(CONF_HOST): cv.string,
                vol.Optional(CONF_PORT, default=443): cv.port,
                vol.Required(CONF_USERNAME): cv.string,
                vol.Required(CONF_PASSWORD): cv.string,
                vol.Optional(CONF_SITE_ID, default="default"): cv.string,
                vol.Optional(CONF_UNIFI_OS, default=False): cv.boolean,
                vol.Optional(CONF_VERIFY_SSL, default=False): cv.boolean,
                vol.Optional(CONF_SCAN_INTERVAL, default=30): cv.positive_int,
            }
        )
    },
    extra=vol.ALLOW_EXTRA,
)
```

## Rate Limiting

The integration respects API rate limits:

- Minimum scan interval: 10 seconds
- Default scan interval: 30 seconds
- WebSocket events: Real-time (no limit)
- Burst protection: Max 10 requests per second

## Error Handling

### API Errors

```python
try:
    data = await self._request("get", "/api/s/default/stat/device")
except Unauthorized:
    _LOGGER.error("Authentication failed")
    raise ConfigEntryAuthFailed
except TimeoutError:
    _LOGGER.warning("Request timeout")
except ClientError as err:
    _LOGGER.error("API error: %s", err)
```

### Reconnection Logic

```python
async def _reconnect(self) -> None:
    """Reconnect to controller."""
    while not self._shutdown:
        try:
            await self.login()
            await self._start_websocket()
            return
        except (TimeoutError, ClientError) as err:
            _LOGGER.warning("Reconnection failed: %s", err)
            await asyncio.sleep(RETRY_INTERVAL)
```

## References

- [Home Assistant Integration Architecture](https://developers.home-assistant.io/docs/creating_integration_index)
- [UniFi API Documentation](https://developer.ui.com/unifi-api)
- [UniFi OS Architecture](https://help.ui.com/hc/en-us/articles/360012282453)
