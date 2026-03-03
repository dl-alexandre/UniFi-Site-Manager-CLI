"""Support for UniFi Controller binary sensors."""

import logging
from typing import Any

from homeassistant.components.binary_sensor import (
    BinarySensorDeviceClass,
    BinarySensorEntity,
)
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from . import UniFiDataUpdateCoordinator
from .const import (
    DOMAIN,
    BINARY_SENSOR_FIRMWARE_UPDATE,
)

_LOGGER = logging.getLogger(__name__)


async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up UniFi Controller binary sensors."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    
    # Create binary sensors for each device that has firmware update available
    entities = []
    devices = coordinator.data.get("devices", [])
    
    for device in devices:
        firmware_status = device.get("firmwareStatus") or device.get("upgradeable")
        if firmware_status:
            entities.append(
                UniFiFirmwareUpdateBinarySensor(coordinator, entry, device)
            )
    
    async_add_entities(entities)


class UniFiFirmwareUpdateBinarySensor(CoordinatorEntity, BinarySensorEntity):
    """Binary sensor for firmware update availability."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
        device: dict[str, Any],
    ) -> None:
        """Initialize the binary sensor."""
        super().__init__(coordinator)
        self._entry = entry
        self._device_id = device.get("_id") or device.get("id")
        self._device_mac = device.get("mac")
        self._device_name = device.get("name") or device.get("model") or self._device_mac
        
        self._attr_unique_id = f"{entry.entry_id}_firmware_{self._device_id}"
        self._attr_name = f"UniFi {self._device_name} Firmware Update"
        self._attr_device_class = BinarySensorDeviceClass.UPDATE

    @property
    def is_on(self) -> bool:
        """Return True if firmware update is available."""
        devices = self.coordinator.data.get("devices", [])
        for device in devices:
            device_id = device.get("_id") or device.get("id")
            if device_id == self._device_id:
                firmware_status = device.get("firmwareStatus") or device.get("upgradeable")
                if isinstance(firmware_status, bool):
                    return firmware_status
                elif isinstance(firmware_status, str):
                    return firmware_status.lower() in ("available", "pending", "true")
                # Check version fields
                current_version = device.get("version")
                available_version = device.get("upgrade_to_firmware") or device.get("available_version")
                if available_version and available_version != current_version:
                    return True
        return False

    @property
    def extra_state_attributes(self) -> dict[str, Any]:
        """Return extra attributes."""
        devices = self.coordinator.data.get("devices", [])
        for device in devices:
            device_id = device.get("_id") or device.get("id")
            if device_id == self._device_id:
                current_version = device.get("version")
                available_version = device.get("upgrade_to_firmware") or device.get("available_version")
                
                return {
                    "device_name": self._device_name,
                    "device_mac": self._device_mac,
                    "device_model": device.get("model"),
                    "device_type": device.get("type"),
                    "current_version": current_version,
                    "available_version": available_version,
                    "adopted": device.get("adopted"),
                    "status": device.get("status") or device.get("state"),
                }
        return {}

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
    def available(self) -> bool:
        """Return True if entity is available."""
        return self.coordinator.last_update_success
