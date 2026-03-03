"""The UniFi Controller integration."""

import asyncio
import logging

from homeassistant.config_entries import ConfigEntry
from homeassistant.const import (
    CONF_HOST,
    CONF_PASSWORD,
    CONF_PORT,
    CONF_USERNAME,
    Platform,
)
from homeassistant.core import HomeAssistant, ServiceCall
from homeassistant.helpers import device_registry as dr
from homeassistant.helpers.update_coordinator import DataUpdateCoordinator, UpdateFailed

from .api import UniFiAPI, UniFiControllerError
from .const import (
    CONF_SITE_ID,
    CONF_UNIFI_OS,
    CONF_VERIFY_SSL,
    DEFAULT_SCAN_INTERVAL,
    DEFAULT_SITE_ID,
    DOMAIN,
    SERVICE_BLOCK_CLIENT,
    SERVICE_RECONNECT_CLIENT,
    SERVICE_RESTART_DEVICE,
    ATTR_DEVICE_MAC,
    ATTR_CLIENT_MAC,
    ATTR_BLOCK,
)

_LOGGER = logging.getLogger(__name__)

PLATFORMS = [Platform.SENSOR, Platform.DEVICE_TRACKER, Platform.BINARY_SENSOR]


async def async_setup_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Set up UniFi Controller from a config entry."""
    hass.data.setdefault(DOMAIN, {})
    
    # Create API instance
    api = UniFiAPI(
        host=entry.data[CONF_HOST],
        port=entry.data[CONF_PORT],
        username=entry.data[CONF_USERNAME],
        password=entry.data[CONF_PASSWORD],
        site_id=entry.data.get(CONF_SITE_ID, DEFAULT_SITE_ID),
        unifi_os=entry.data.get(CONF_UNIFI_OS, True),
        verify_ssl=entry.data.get(CONF_VERIFY_SSL, False),
    )
    
    # Create update coordinator
    coordinator = UniFiDataUpdateCoordinator(
        hass,
        api,
        entry,
    )
    
    # Fetch initial data
    await coordinator.async_config_entry_first_refresh()
    
    # Store coordinator
    hass.data[DOMAIN][entry.entry_id] = coordinator
    
    # Setup platforms
    await hass.config_entries.async_forward_entry_setups(entry, PLATFORMS)
    
    # Setup services
    async def async_restart_device(call: ServiceCall) -> None:
        """Restart a UniFi device."""
        mac_address = call.data.get(ATTR_DEVICE_MAC)
        if not mac_address:
            _LOGGER.error("No device MAC provided")
            return
        
        try:
            await api.restart_device(mac_address)
            _LOGGER.info("Restarted device %s", mac_address)
        except Exception as err:
            _LOGGER.error("Failed to restart device %s: %s", mac_address, err)
    
    async def async_block_client(call: ServiceCall) -> None:
        """Block or unblock a client."""
        mac_address = call.data.get(ATTR_CLIENT_MAC)
        block = call.data.get(ATTR_BLOCK, True)
        
        if not mac_address:
            _LOGGER.error("No client MAC provided")
            return
        
        try:
            await api.block_client(mac_address, block)
            action = "blocked" if block else "unblocked"
            _LOGGER.info("%s client %s", action, mac_address)
            
            # Refresh coordinator to update state
            await coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Failed to block client %s: %s", mac_address, err)
    
    async def async_reconnect_client(call: ServiceCall) -> None:
        """Force a client to reconnect."""
        mac_address = call.data.get(ATTR_CLIENT_MAC)
        
        if not mac_address:
            _LOGGER.error("No client MAC provided")
            return
        
        try:
            await api.reconnect_client(mac_address)
            _LOGGER.info("Reconnected client %s", mac_address)
        except Exception as err:
            _LOGGER.error("Failed to reconnect client %s: %s", mac_address, err)
    
    hass.services.async_register(
        DOMAIN, SERVICE_RESTART_DEVICE, async_restart_device
    )
    hass.services.async_register(
        DOMAIN, SERVICE_BLOCK_CLIENT, async_block_client
    )
    hass.services.async_register(
        DOMAIN, SERVICE_RECONNECT_CLIENT, async_reconnect_client
    )
    
    return True


async def async_unload_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Unload a config entry."""
    unload_ok = await hass.config_entries.async_unload_platforms(entry, PLATFORMS)
    
    if unload_ok:
        coordinator = hass.data[DOMAIN].pop(entry.entry_id)
        await coordinator.api.close()
    
    return unload_ok


class UniFiDataUpdateCoordinator(DataUpdateCoordinator):
    """Class to manage fetching data from UniFi Controller."""

    def __init__(
        self,
        hass: HomeAssistant,
        api: UniFiAPI,
        entry: ConfigEntry,
    ) -> None:
        """Initialize the coordinator."""
        self.api = api
        self.entry = entry
        
        scan_interval = entry.options.get("scan_interval", DEFAULT_SCAN_INTERVAL)
        
        super().__init__(
            hass,
            _LOGGER,
            name=f"UniFi {entry.data[CONF_HOST]}",
            update_interval=asyncio.timedelta(seconds=scan_interval),
        )

    async def _async_update_data(self):
        """Fetch data from UniFi Controller."""
        try:
            data = {
                "devices": await self.api.get_devices(),
                "clients": await self.api.get_clients(),
                "health": await self.api.get_health(),
                "alerts": await self.api.get_alerts(),
            }
            return data
        except UniFiControllerError as err:
            raise UpdateFailed(f"Error communicating with API: {err}")
