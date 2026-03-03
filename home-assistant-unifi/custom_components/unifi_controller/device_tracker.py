"""Support for UniFi Controller device trackers."""

import logging
from typing import Any

from homeassistant.components.device_tracker import SourceType
from homeassistant.components.device_tracker.config_entry import TrackerEntity
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant, callback
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from . import UniFiDataUpdateCoordinator
from .const import DOMAIN

_LOGGER = logging.getLogger(__name__)


async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up UniFi Controller device trackers."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    
    @callback
    def async_add_client_entities():
        """Add client tracker entities."""
        entities = []
        clients = coordinator.data.get("clients", [])
        
        for client in clients:
            mac = client.get("mac")
            if not mac:
                continue
                
            # Check if entity already exists
            entity_id = f"{DOMAIN}_{entry.entry_id}_client_{mac.replace(':', '')}"
            if entity_id not in hass.data.get("entity_registry", {}):
                entities.append(UniFiClientTracker(coordinator, entry, client))
        
        if entities:
            async_add_entities(entities, True)
    
    # Add initial entities
    async_add_client_entities()
    
    # Listen for updates to add new clients
    entry.async_on_unload(
        coordinator.async_add_listener(async_add_client_entities)
    )


class UniFiClientTracker(CoordinatorEntity, TrackerEntity):
    """Representation of a UniFi client device tracker."""

    def __init__(
        self,
        coordinator: UniFiDataUpdateCoordinator,
        entry: ConfigEntry,
        client: dict[str, Any],
    ) -> None:
        """Initialize the tracker."""
        super().__init__(coordinator)
        self._entry = entry
        self._mac = client.get("mac", "")
        self._attr_unique_id = f"{entry.entry_id}_client_{self._mac.replace(':', '')}"
        
        # Set name from client data
        name = client.get("name") or client.get("hostname") or self._mac
        self._attr_name = f"UniFi Client {name}"
        
        # Set entity ID
        self.entity_id = f"device_tracker.unifi_client_{self._mac.replace(':', '').lower()}"
        
        # Store initial client data
        self._client_data = client

    @property
    def source_type(self) -> SourceType:
        """Return the source type."""
        return SourceType.ROUTER

    @property
    def is_connected(self) -> bool:
        """Return true if the client is connected."""
        # Find current client data
        clients = self.coordinator.data.get("clients", [])
        for client in clients:
            if client.get("mac") == self._mac:
                return True
        return False

    @property
    def mac_address(self) -> str:
        """Return the mac address."""
        return self._mac

    @property
    def ip_address(self) -> str | None:
        """Return the IP address."""
        clients = self.coordinator.data.get("clients", [])
        for client in clients:
            if client.get("mac") == self._mac:
                return client.get("ip")
        return None

    @property
    def hostname(self) -> str | None:
        """Return the hostname."""
        clients = self.coordinator.data.get("clients", [])
        for client in clients:
            if client.get("mac") == self._mac:
                return client.get("hostname") or client.get("name")
        return None

    @property
    def extra_state_attributes(self) -> dict[str, Any]:
        """Return extra attributes."""
        clients = self.coordinator.data.get("clients", [])
        for client in clients:
            if client.get("mac") == self._mac:
                return {
                    "mac_address": self._mac,
                    "ip_address": client.get("ip"),
                    "hostname": client.get("hostname") or client.get("name"),
                    "is_wired": client.get("is_wired") or client.get("isWired"),
                    "ssid": client.get("essid") or client.get("ssid"),
                    "ap_mac": client.get("ap_mac") or client.get("apMac"),
                    "signal": client.get("signal"),
                    "rssi": client.get("rssi"),
                    "rx_bytes": client.get("rx_bytes") or client.get("rxBytes"),
                    "tx_bytes": client.get("tx_bytes") or client.get("txBytes"),
                    "uptime": client.get("uptime"),
                    "first_seen": client.get("first_seen") or client.get("firstSeen"),
                    "is_guest": client.get("is_guest") or client.get("isGuest"),
                    "is_blocked": client.get("blocked") or client.get("isBlocked"),
                    "oui": client.get("oui"),
                    "network": client.get("network"),
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
    def icon(self):
        """Return the icon."""
        clients = self.coordinator.data.get("clients", [])
        for client in clients:
            if client.get("mac") == self._mac:
                is_wired = client.get("is_wired") or client.get("isWired")
                if is_wired:
                    return "mdi:ethernet"
                else:
                    return "mdi:wifi"
        return "mdi:lan-disconnect"

    @property
    def available(self) -> bool:
        """Return True if entity is available."""
        return self.coordinator.last_update_success
