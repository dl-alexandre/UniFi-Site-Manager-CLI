"""Config flow for UniFi Controller integration."""

import logging
import voluptuous as vol

from homeassistant import config_entries
from homeassistant.const import (
    CONF_HOST,
    CONF_PASSWORD,
    CONF_PORT,
    CONF_USERNAME,
)
from homeassistant.core import callback
from homeassistant.data_entry_flow import FlowResult

from .const import (
    CONF_SITE_ID,
    CONF_UNIFI_OS,
    CONF_VERIFY_SSL,
    DEFAULT_PORT,
    DEFAULT_SITE_ID,
    DEFAULT_VERIFY_SSL,
    DOMAIN,
)
from .api import UniFiAPI, UniFiAuthenticationError, UniFiConnectionError

_LOGGER = logging.getLogger(__name__)


class UniFiConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Handle a config flow for UniFi Controller."""

    VERSION = 1

    async def async_step_user(
        self, user_input: dict | None = None
    ) -> FlowResult:
        """Handle the initial step."""
        errors = {}

        if user_input is not None:
            # Validate the connection
            try:
                api = UniFiAPI(
                    host=user_input[CONF_HOST],
                    port=user_input[CONF_PORT],
                    username=user_input[CONF_USERNAME],
                    password=user_input[CONF_PASSWORD],
                    site_id=user_input.get(CONF_SITE_ID, DEFAULT_SITE_ID),
                    unifi_os=user_input.get(CONF_UNIFI_OS, True),
                    verify_ssl=user_input.get(CONF_VERIFY_SSL, DEFAULT_VERIFY_SSL),
                )
                
                # Test the connection
                await api.login()
                
                # Get sites for the user to choose from
                sites = await api.get_sites()
                await api.close()
                
                if not sites:
                    errors["base"] = "no_sites"
                else:
                    # Create a unique ID based on host and site
                    await self.async_set_unique_id(
                        f"{user_input[CONF_HOST]}_{user_input.get(CONF_SITE_ID, DEFAULT_SITE_ID)}"
                    )
                    self._abort_if_unique_id_configured()
                    
                    return self.async_create_entry(
                        title=f"UniFi ({user_input[CONF_HOST]})",
                        data=user_input,
                    )
                    
            except UniFiAuthenticationError:
                errors["base"] = "invalid_auth"
            except UniFiConnectionError:
                errors["base"] = "cannot_connect"
            except Exception:
                _LOGGER.exception("Unexpected exception")
                errors["base"] = "unknown"

        # Show the form
        data_schema = vol.Schema({
            vol.Required(CONF_HOST): str,
            vol.Required(CONF_PORT, default=DEFAULT_PORT): int,
            vol.Required(CONF_USERNAME): str,
            vol.Required(CONF_PASSWORD): str,
            vol.Optional(CONF_SITE_ID, default=DEFAULT_SITE_ID): str,
            vol.Optional(CONF_UNIFI_OS, default=True): bool,
            vol.Optional(CONF_VERIFY_SSL, default=DEFAULT_VERIFY_SSL): bool,
        })

        return self.async_show_form(
            step_id="user",
            data_schema=data_schema,
            errors=errors,
        )

    @staticmethod
    @callback
    def async_get_options_flow(
        config_entry: config_entries.ConfigEntry,
    ) -> config_entries.OptionsFlow:
        """Create the options flow."""
        return UniFiOptionsFlowHandler(config_entry)


class UniFiOptionsFlowHandler(config_entries.OptionsFlow):
    """Handle options flow for UniFi Controller."""

    def __init__(self, config_entry: config_entries.ConfigEntry) -> None:
        """Initialize options flow."""
        self.config_entry = config_entry

    async def async_step_init(
        self, user_input: dict | None = None
    ) -> FlowResult:
        """Manage the options."""
        if user_input is not None:
            return self.async_create_entry(title="", data=user_input)

        options_schema = vol.Schema({
            vol.Optional(
                "scan_interval",
                default=self.config_entry.options.get("scan_interval", 30),
            ): vol.All(vol.Coerce(int), vol.Range(min=10, max=300)),
        })

        return self.async_show_form(
            step_id="init",
            data_schema=options_schema,
        )
