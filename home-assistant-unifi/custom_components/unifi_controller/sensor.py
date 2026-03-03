"""Support for UniFi Controller sensors."""

import logging
from typing import Any

from homeassistant.components.sensor import (
    SensorDeviceClass,
    SensorEntity,
    SensorStateClass,
)
from homeassistant.config_entries import ConfigEntry
from homeassistant.const import (
    UnitOfDataRate,
    UnitOfInformation,
    PERCENTAGE,
)
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from . import UniFiDataUpdateCoordinator
from .const import (
    DOMAIN,
    SENSOR_ALERTS_COUNT,
    SENSOR_CLIENTS_CONNECTED,
    SENSOR_CPU_USAGE,
    SENSOR_DEVICES_ONLINE,
    SENSOR_GATEWAY_DOWN_SPEED,
    SENSOR_GATEWAY_UP_SPEED,
    SENSOR_MEMORY_USAGE,
)

_LOGGER = logging.getLogger(__name__)


async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up UniFi Controller sensors."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    
    sensors = [
        UniFiDevicesOnlineSensor(coordinator, entry),
        UniFiClientsConnectedSensor(coordinator, entry),
        UniFiAlertsCountSensor(coordinator, entry),
    ]
    
    # Add optional sensors if data available
    sensors.append(UniFiCPUUsageSensor(coordinator, entry))
    sensors.append(UniFiMemoryUsageSensor(coordinator, entry))
    sensors.append(UniFiGatewayUpSpeedSensor(coordinator, entry))
    sensors.append(UniFiGatewayDownSpeedSensor(coordinator, entry))
    
    async_add_entities(sensors)


class UniFiSensor(CoordinatorEntity, SensorEntity):
    """Base class for UniFi sensors."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
        sensor_type: str,
        name: str,
        icon: str,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(coordinator)
        self._sensor_type = sensor_type
        self._attr_name = f"UniFi {name}"
        self._attr_unique_id = f"{entry.entry_id}_{sensor_type}"
        self._attr_icon = icon
        self._entry = entry

    @property
    def device_info(self):
        """Return device info."""
        return {
            "identifiers": {(DOMAIN, self._entry.entry_id)},
            "name": f"UniFi Controller ({self._entry.data['host']})",
            "manufacturer": "Ubiquiti",
            "model": "UniFi Network Controller",
        }

    @property
    def available(self):
        """Return True if entity is available."""
        return self.coordinator.last_update_success


class UniFiDevicesOnlineSensor(UniFiSensor):
    """Sensor for total devices online."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_DEVICES_ONLINE,
            "Devices Online",
            "mdi:network-outline",
        )
        self._attr_state_class = SensorStateClass.MEASUREMENT

    @property
    def native_value(self):
        """Return the state of the sensor."""
        devices = self.coordinator.data.get("devices", [])
        # Count devices with state 1 (online) or status ONLINE
        online_count = 0
        for device in devices:
            state = device.get("state")
            status = device.get("status")
            if state == 1 or status == "ONLINE":
                online_count += 1
        return online_count

    @property
    def extra_state_attributes(self):
        """Return extra attributes."""
        devices = self.coordinator.data.get("devices", [])
        total = len(devices)
        return {
            "total_devices": total,
            "online_devices": self.native_value,
            "offline_devices": total - self.native_value,
        }


class UniFiClientsConnectedSensor(UniFiSensor):
    """Sensor for total clients connected."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_CLIENTS_CONNECTED,
            "Clients Connected",
            "mdi:devices",
        )
        self._attr_state_class = SensorStateClass.MEASUREMENT

    @property
    def native_value(self):
        """Return the state of the sensor."""
        clients = self.coordinator.data.get("clients", [])
        return len(clients)

    @property
    def extra_state_attributes(self):
        """Return extra attributes."""
        clients = self.coordinator.data.get("clients", [])
        wired = sum(1 for c in clients if c.get("is_wired") or c.get("isWired"))
        wireless = len(clients) - wired
        
        return {
            "total_clients": len(clients),
            "wired_clients": wired,
            "wireless_clients": wireless,
        }


class UniFiAlertsCountSensor(UniFiSensor):
    """Sensor for alerts count."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_ALERTS_COUNT,
            "Alerts",
            "mdi:alert",
        )
        self._attr_state_class = SensorStateClass.MEASUREMENT

    @property
    def native_value(self):
        """Return the state of the sensor."""
        alerts = self.coordinator.data.get("alerts", [])
        # Count unacknowledged alerts
        unacknowledged = sum(
            1 for a in alerts 
            if not a.get("acknowledged") and not a.get("archived")
        )
        return unacknowledged

    @property
    def extra_state_attributes(self):
        """Return extra attributes."""
        alerts = self.coordinator.data.get("alerts", [])
        total = len(alerts)
        unacknowledged = self.native_value
        
        # Get recent alert messages
        recent_alerts = [
            a.get("msg", a.get("message", "Unknown"))
            for a in alerts[:5]
        ]
        
        return {
            "total_alerts": total,
            "unacknowledged": unacknowledged,
            "recent_alerts": recent_alerts,
        }


class UniFiCPUUsageSensor(UniFiSensor):
    """Sensor for CPU usage."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_CPU_USAGE,
            "CPU Usage",
            "mdi:cpu-64-bit",
        )
        self._attr_native_unit_of_measurement = PERCENTAGE
        self._attr_state_class = SensorStateClass.MEASUREMENT

    @property
    def native_value(self):
        """Return the state of the sensor."""
        # Try to get CPU from gateway device
        devices = self.coordinator.data.get("devices", [])
        for device in devices:
            if device.get("type") == "udm" or "gateway" in device.get("type", "").lower():
                cpu = device.get("cpu")
                if cpu is not None:
                    return round(cpu, 1)
        return None


class UniFiMemoryUsageSensor(UniFiSensor):
    """Sensor for memory usage."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_MEMORY_USAGE,
            "Memory Usage",
            "mdi:memory",
        )
        self._attr_native_unit_of_measurement = PERCENTAGE
        self._attr_state_class = SensorStateClass.MEASUREMENT

    @property
    def native_value(self):
        """Return the state of the sensor."""
        # Try to get memory from gateway device
        devices = self.coordinator.data.get("devices", [])
        for device in devices:
            if device.get("type") == "udm" or "gateway" in device.get("type", "").lower():
                memory = device.get("memory")
                if memory is not None:
                    return round(memory, 1)
        return None


class UniFiGatewayUpSpeedSensor(UniFiSensor):
    """Sensor for gateway upload speed."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_GATEWAY_UP_SPEED,
            "WAN Upload",
            "mdi:upload",
        )
        self._attr_native_unit_of_measurement = UnitOfDataRate.MEGABITS_PER_SECOND
        self._attr_state_class = SensorStateClass.MEASUREMENT
        self._attr_device_class = SensorDeviceClass.DATA_RATE

    @property
    def native_value(self):
        """Return the state of the sensor."""
        health = self.coordinator.data.get("health", [])
        for item in health:
            if item.get("subsystem") == "wan":
                tx_rate = item.get("txRate")
                if tx_rate:
                    # Convert to Mbps
                    return round(tx_rate / 1000000, 2)
        return None


class UniFiGatewayDownSpeedSensor(UniFiSensor):
    """Sensor for gateway download speed."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(
            coordinator,
            entry,
            SENSOR_GATEWAY_DOWN_SPEED,
            "WAN Download",
            "mdi:download",
        )
        self._attr_native_unit_of_measurement = UnitOfDataRate.MEGABITS_PER_SECOND
        self._attr_state_class = SensorStateClass.MEASUREMENT
        self._attr_device_class = SensorDeviceClass.DATA_RATE

    @property
    def native_value(self):
        """Return the state of the sensor."""
        health = self.coordinator.data.get("health", [])
        for item in health:
            if item.get("subsystem") == "wan":
                rx_rate = item.get("rxRate")
                if rx_rate:
                    # Convert to Mbps
                    return round(rx_rate / 1000000, 2)
        return None
