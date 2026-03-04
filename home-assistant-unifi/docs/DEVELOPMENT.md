# Development Guide

Guide for contributing to the UniFi Controller Home Assistant integration.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Setup](#development-setup)
- [Architecture](#architecture)
- [Adding Features](#adding-features)
- [Testing](#testing)
- [Debugging](#debugging)
- [Pull Requests](#pull-requests)

## Prerequisites

- **Python**: 3.11 or newer
- **Home Assistant**: Development environment
- **Git**: For version control
- **Code Editor**: VS Code recommended

### Knowledge Required

- Python 3.x
- Home Assistant architecture
- Async/await patterns
- Basic networking concepts
- UniFi API (helpful but not required)

## Getting Started

### Fork Repository

```bash
# Fork on GitHub, then clone
git clone https://github.com/YOUR_USERNAME/Local-UniFi-CLI.git
cd Local-UniFi-CLI/home-assistant-unifi
```

### Create Branch

```bash
git checkout -b feature/my-new-feature
# or
git checkout -b fix/issue-description
```

## Project Structure

```
custom_components/unifi_controller/
├── __init__.py           # Integration setup
├── config_flow.py        # Configuration UI
├── const.py             # Constants
├── api.py               # API client
├── coordinator.py       # Data update coordinator
├── sensor.py            # Sensor platform
├── device_tracker.py    # Device tracker platform
├── binary_sensor.py     # Binary sensor platform
├── services.py          # Service definitions
└── manifest.json        # Integration manifest
```

### File Descriptions

| File | Purpose |
|------|---------|
| `__init__.py` | Integration entry point, sets up platforms |
| `config_flow.py` | UI configuration flow |
| `const.py` | Constants (domains, config keys) |
| `api.py` | HTTP/WebSocket client for UniFi API |
| `coordinator.py` | Data update coordinator for polling |
| `sensor.py` | Sensor entity implementation |
| `device_tracker.py` | Device tracker implementation |
| `binary_sensor.py` | Binary sensor implementation |
| `services.py` | Service definitions and handlers |
| `manifest.json` | Integration metadata |

## Development Setup

### Option 1: Home Assistant Development Container

1. Install [VS Code](https://code.visualstudio.com/)
2. Install [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
3. Clone repository
4. Open in VS Code
5. Click "Reopen in Container"

### Option 2: Local Development

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# Install Home Assistant
git clone https://github.com/home-assistant/core.git
cd core
pip install -e .

# Copy integration
mkdir -p config/custom_components
cp -r /path/to/home-assistant-unifi/custom_components/unifi_controller config/custom_components/

# Run Home Assistant
cd config
hass -c .
```

### Install Dependencies

```bash
pip install -r requirements_dev.txt
```

## Architecture

### Data Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   UniFi     │────▶│   API       │────▶│Coordinator  │
│ Controller  │     │   Client    │     │             │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                                │
                       ┌────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Sensor    │ │  Device     │ │   Binary    │
│   Entities  │ │  Trackers   │ │   Sensor   │
└─────────────┘ └─────────────┘ └─────────────┘
```

### Key Classes

#### UniFiControllerAPI

```python
class UniFiControllerAPI:
    """API client for UniFi Controller."""
    
    async def login(self) -> bool:
        """Authenticate with controller."""
        
    async def get_devices(self) -> list[Device]:
        """Get all devices."""
        
    async def get_clients(self) -> list[Client]:
        """Get all clients."""
        
    async def restart_device(self, mac: str) -> None:
        """Restart a device."""
```

#### UniFiDataUpdateCoordinator

```python
class UniFiDataUpdateCoordinator(DataUpdateCoordinator):
    """Coordinator to fetch data."""
    
    async def _async_update_data(self):
        """Fetch data from API."""
        devices = await self.api.get_devices()
        clients = await self.api.get_clients()
        return {"devices": devices, "clients": clients}
```

#### Sensor Entity

```python
class UniFiDeviceSensor(CoordinatorEntity, SensorEntity):
    """Device count sensor."""
    
    @property
    def native_value(self):
        """Return the state."""
        data = self.coordinator.data
        return len([d for d in data["devices"] if d["state"] == 1])
```

## Adding Features

### Adding a New Sensor

1. **Define the sensor class** in `sensor.py`:

```python
class UniFiNewSensor(CoordinatorEntity, SensorEntity):
    """Description of new sensor."""
    
    def __init__(self, coordinator, site_id):
        super().__init__(coordinator)
        self._site_id = site_id
        self._attr_unique_id = f"{site_id}_new_sensor"
        self._attr_name = "UniFi New Sensor"
    
    @property
    def native_value(self):
        """Return sensor value."""
        data = self.coordinator.data
        # Calculate value from data
        return value
    
    @property
    def extra_state_attributes(self):
        """Return additional attributes."""
        return {
            "site_id": self._site_id,
            "custom_attr": "value",
        }
```

2. **Register in async_setup_entry**:

```python
async def async_setup_entry(hass, entry, async_add_entities):
    coordinator = hass.data[DOMAIN][entry.entry_id]
    
    sensors = [
        UniFiDeviceSensor(coordinator, site_id),
        UniFiNewSensor(coordinator, site_id),  # Add new sensor
    ]
    
    async_add_entities(sensors)
```

### Adding a New Service

1. **Define service handler** in `services.py`:

```python
async def async_setup_services(hass: HomeAssistant, entry: ConfigEntry) -> None:
    """Set up services."""
    
    async def handle_new_service(call: ServiceCall) -> None:
        """Handle the service call."""
        controller = hass.data[DOMAIN][entry.entry_id]
        param = call.data.get("param")
        
        await controller.api.new_api_method(param)
    
    hass.services.async_register(
        DOMAIN, "new_service", handle_new_service,
        schema=vol.Schema({
            vol.Required("param"): cv.string,
        })
    )
```

2. **Call in `__init__.py`**:

```python
from .services import async_setup_services

async def async_setup_entry(hass, entry):
    # ... existing setup ...
    await async_setup_services(hass, entry)
```

### Adding WebSocket Event Handler

1. **Extend event handler** in `api.py`:

```python
async def _process_websocket_message(self, message: dict) -> None:
    """Process WebSocket message."""
    meta = message.get("meta", {})
    msg_type = meta.get("message", "")
    
    if msg_type == "new:event:type":
        await self._handle_new_event(message.get("data", []))

async def _handle_new_event(self, data: list) -> None:
    """Handle new event type."""
    for event in data:
        # Process event
        self.hass.bus.fire("unifi_new_event", event)
```

## Testing

### Running Tests

```bash
# Run all tests
pytest tests/

# Run specific test
pytest tests/test_sensor.py

# Run with coverage
pytest --cov=custom_components/unifi_controller tests/

# Verbose output
pytest -v tests/
```

### Writing Tests

```python
# tests/test_sensor.py
from homeassistant.core import HomeAssistant
from custom_components.unifi_controller.sensor import UniFiDeviceSensor

async def test_device_sensor(hass: HomeAssistant):
    """Test device sensor."""
    # Setup test data
    coordinator = Mock()
    coordinator.data = {
        "devices": [
            {"_id": "1", "state": 1},
            {"_id": "2", "state": 0},
        ]
    }
    
    sensor = UniFiDeviceSensor(coordinator, "default")
    
    # Assert expected value
    assert sensor.native_value == 1  # 1 device online
```

### Mock API for Testing

```python
# conftest.py
import pytest
from unittest.mock import AsyncMock, Mock

@pytest.fixture
def mock_api():
    """Create mock API client."""
    api = Mock()
    api.get_devices = AsyncMock(return_value=[])
    api.get_clients = AsyncMock(return_value=[])
    api.login = AsyncMock(return_value=True)
    return api
```

## Debugging

### Enable Debug Logging

```yaml
# configuration.yaml
logger:
  logs:
    custom_components.unifi_controller: debug
```

### Debug with VS Code

1. Set breakpoints in code
2. Press F5 to start debugging
3. Use Debug Console to inspect variables

### Common Debug Output

```python
_LOGGER.debug("Fetching devices from controller")
_LOGGER.debug("Received %s devices", len(devices))
_LOGGER.warning("Device offline: %s", device_mac)
_LOGGER.error("API request failed: %s", error)
```

### Testing API Calls

```python
# Test script
import asyncio
from custom_components.unifi_controller.api import UniFiControllerAPI

async def test():
    api = UniFiControllerAPI("192.168.1.1", 443, "admin", "pass")
    await api.login()
    devices = await api.get_devices()
    print(f"Found {len(devices)} devices")

asyncio.run(test())
```

## Pull Requests

### Before Submitting

1. **Test your changes**:
   - Run all tests
   - Test in real Home Assistant
   - Test different controller types

2. **Check code style**:
   ```bash
   flake8 custom_components/unifi_controller/
   pylint custom_components/unifi_controller/
   black custom_components/unifi_controller/
   ```

3. **Update documentation**:
   - Update README.md if needed
   - Add to CHANGELOG.md
   - Update USAGE.md for new features

4. **Write good commit messages**:
   ```
   feat: Add device restart service
   fix: Handle WebSocket disconnection
   docs: Update installation instructions
   test: Add coverage for sensor entities
   ```

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tested on UDM
- [ ] Tested on Cloud Key
- [ ] All tests pass

## Checklist
- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. PR submitted
2. Automated tests run
3. Maintainer review
4. Feedback addressed
5. Approved and merged

## Resources

### Documentation

- [Home Assistant Developer Docs](https://developers.home-assistant.io/)
- [Architecture Decision Records](https://github.com/home-assistant/architecture)
- [UniFi API Documentation](https://developer.ui.com/unifi-api)

### Tools

- [Home Assistant Dev Container](https://developers.home-assistant.io/docs/development_environment)
- [pytest-homeassistant-custom-component](https://github.com/MatthewFlamm/pytest-homeassistant-custom-component)
- [HACS Action](https://github.com/hacs/action)

### Community

- [Home Assistant Discord](https://discord.gg/home-assistant)
- [Home Assistant Community Forum](https://community.home-assistant.io/)
- GitHub Discussions

## Questions?

- Open an issue for bugs
- Start a discussion for questions
- Check existing documentation

Thank you for contributing!
