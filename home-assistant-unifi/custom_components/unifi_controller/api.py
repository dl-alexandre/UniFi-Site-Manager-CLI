"""UniFi Controller API client for Home Assistant."""

import json
import logging
import ssl
from typing import Any
from urllib.parse import urljoin

import aiohttp
import async_timeout

_LOGGER = logging.getLogger(__name__)


class UniFiControllerError(Exception):
    """Base exception for UniFi Controller errors."""
    pass


class UniFiAuthenticationError(UniFiControllerError):
    """Authentication failed."""
    pass


class UniFiConnectionError(UniFiControllerError):
    """Connection error."""
    pass


class UniFiAPI:
    """UniFi Controller API client."""

    def __init__(
        self,
        host: str,
        port: int,
        username: str,
        password: str,
        site_id: str = "default",
        unifi_os: bool = True,
        verify_ssl: bool = False,
    ) -> None:
        """Initialize the API client."""
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.site_id = site_id
        self.unifi_os = unifi_os
        self.verify_ssl = verify_ssl
        
        self._session: aiohttp.ClientSession | None = None
        self._csrf_token: str | None = None
        self._logged_in = False
        
        # Build base URL
        protocol = "https" if port == 443 else "http"
        self._base_url = f"{protocol}://{host}:{port}"
        
        # Select endpoint set
        from .const import API_ENDPOINTS, API_ENDPOINTS_STANDALONE
        self._endpoints = API_ENDPOINTS if unifi_os else API_ENDPOINTS_STANDALONE

    def _get_endpoint(self, key: str, **kwargs) -> str:
        """Get API endpoint with formatting."""
        endpoint = self._endpoints[key]
        if kwargs:
            endpoint = endpoint.format(**kwargs)
        return urljoin(self._base_url, endpoint)

    async def _get_session(self) -> aiohttp.ClientSession:
        """Get or create aiohttp session."""
        if self._session is None or self._session.closed:
            ssl_context = ssl.create_default_context()
            if not self.verify_ssl:
                ssl_context.check_hostname = False
                ssl_context.verify_mode = ssl.CERT_NONE
            
            connector = aiohttp.TCPConnector(ssl=ssl_context)
            timeout = aiohttp.ClientTimeout(total=30)
            self._session = aiohttp.ClientSession(
                connector=connector,
                timeout=timeout,
                headers={"Content-Type": "application/json"},
            )
        return self._session

    async def login(self) -> bool:
        """Authenticate with the UniFi Controller."""
        try:
            session = await self._get_session()
            url = self._get_endpoint("login")
            
            payload = {
                "username": self.username,
                "password": self.password,
            }
            
            _LOGGER.debug("Logging in to %s", url)
            
            async with async_timeout.timeout(30):
                async with session.post(
                    url,
                    json=payload,
                ) as response:
                    if response.status == 401:
                        raise UniFiAuthenticationError("Invalid credentials")
                    elif response.status != 200:
                        text = await response.text()
                        raise UniFiConnectionError(
                            f"Login failed: {response.status} - {text}"
                        )
                    
                    # Extract CSRF token for UniFi OS
                    if self.unifi_os:
                        self._csrf_token = response.headers.get(
                            "X-CSRF-Token"
                        ) or response.headers.get("x-csrf-token")
                        if self._csrf_token:
                            session.headers["X-CSRF-Token"] = self._csrf_token
                    
                    self._logged_in = True
                    _LOGGER.debug("Successfully logged in")
                    return True
                    
        except aiohttp.ClientError as err:
            raise UniFiConnectionError(f"Connection error: {err}")
        except asyncio.TimeoutError:
            raise UniFiConnectionError("Connection timeout")

    async def logout(self) -> None:
        """Logout from the UniFi Controller."""
        if self._session and self._logged_in:
            try:
                url = self._get_endpoint("logout")
                async with self._session.post(url) as response:
                    pass
            except Exception:
                pass
            finally:
                self._logged_in = False
                self._csrf_token = None

    async def _request(
        self,
        method: str,
        endpoint_key: str,
        **kwargs
    ) -> dict[str, Any]:
        """Make an authenticated API request."""
        if not self._logged_in:
            await self.login()
        
        session = await self._get_session()
        url = self._get_endpoint(endpoint_key, site=self.site_id, **kwargs)
        
        headers = {}
        if self._csrf_token and method in ("POST", "PUT", "DELETE"):
            headers["X-CSRF-Token"] = self._csrf_token
        
        try:
            async with async_timeout.timeout(30):
                async with session.request(
                    method,
                    url,
                    headers=headers,
                    **kwargs
                ) as response:
                    if response.status == 401:
                        # Token expired, try to re-login
                        self._logged_in = False
                        await self.login()
                        # Retry the request
                        async with session.request(
                            method,
                            url,
                            headers=headers,
                            **kwargs
                        ) as retry_response:
                            if retry_response.status != 200:
                                text = await retry_response.text()
                                raise UniFiControllerError(
                                    f"API error: {retry_response.status} - {text}"
                                )
                            return await retry_response.json()
                    
                    if response.status != 200:
                        text = await response.text()
                        raise UniFiControllerError(
                            f"API error: {response.status} - {text}"
                        )
                    
                    return await response.json()
                    
        except aiohttp.ClientError as err:
            raise UniFiConnectionError(f"Request failed: {err}")
        except asyncio.TimeoutError:
            raise UniFiConnectionError("Request timeout")

    async def get_sites(self) -> list[dict[str, Any]]:
        """Get list of sites."""
        data = await self._request("GET", "sites")
        return data.get("data", [])

    async def get_devices(self) -> list[dict[str, Any]]:
        """Get list of devices."""
        data = await self._request("GET", "devices")
        return data.get("data", [])

    async def get_clients(self) -> list[dict[str, Any]]:
        """Get list of connected clients."""
        data = await self._request("GET", "clients")
        return data.get("data", [])

    async def get_health(self) -> list[dict[str, Any]]:
        """Get site health status."""
        data = await self._request("GET", "health")
        return data.get("data", [])

    async def get_alerts(self) -> list[dict[str, Any]]:
        """Get list of alerts."""
        data = await self._request("GET", "alerts")
        return data.get("data", [])

    async def get_wlans(self) -> list[dict[str, Any]]:
        """Get list of wireless networks."""
        data = await self._request("GET", "wlans")
        return data.get("data", [])

    async def restart_device(self, mac_address: str) -> bool:
        """Restart a device by MAC address."""
        url = self._get_endpoint("restart_device", site=self.site_id)
        session = await self._get_session()
        
        payload = {
            "cmd": "restart",
            "mac": mac_address,
        }
        
        headers = {}
        if self._csrf_token:
            headers["X-CSRF-Token"] = self._csrf_token
        
        async with session.post(url, json=payload, headers=headers) as response:
            return response.status == 200

    async def block_client(self, mac_address: str, block: bool = True) -> bool:
        """Block or unblock a client by MAC address."""
        url = self._get_endpoint("block_client", site=self.site_id, mac=mac_address)
        session = await self._get_session()
        
        payload = {
            "blocked": block,
        }
        
        headers = {}
        if self._csrf_token:
            headers["X-CSRF-Token"] = self._csrf_token
        
        async with session.put(url, json=payload, headers=headers) as response:
            return response.status == 200

    async def reconnect_client(self, mac_address: str) -> bool:
        """Force a client to reconnect."""
        url = self._get_endpoint("reconnect_client", site=self.site_id)
        session = await self._get_session()
        
        payload = {
            "cmd": "kick-sta",
            "mac": mac_address,
        }
        
        headers = {}
        if self._csrf_token:
            headers["X-CSRF-Token"] = self._csrf_token
        
        async with session.post(url, json=payload, headers=headers) as response:
            return response.status == 200

    async def close(self) -> None:
        """Close the session."""
        await self.logout()
        if self._session and not self._session.closed:
            await self._session.close()
            self._session = None


import asyncio  # noqa: E402
