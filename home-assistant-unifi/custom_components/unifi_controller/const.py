"""Constants for the UniFi Controller integration."""

DOMAIN = "unifi_controller"

# Configuration keys
CONF_HOST = "host"
CONF_PORT = "port"
CONF_USERNAME = "username"
CONF_PASSWORD = "password"
CONF_SITE_ID = "site_id"
CONF_UNIFI_OS = "unifi_os"
CONF_VERIFY_SSL = "verify_ssl"
CONF_SCAN_INTERVAL = "scan_interval"

# Defaults
DEFAULT_PORT = 443
DEFAULT_SITE_ID = "default"
DEFAULT_SCAN_INTERVAL = 30
DEFAULT_VERIFY_SSL = False

# API Endpoints for UniFi OS (local controllers)
API_ENDPOINTS = {
    "login": "/api/auth/login",
    "logout": "/api/auth/logout",
    "whoami": "/api/auth/user-info",
    "sites": "/proxy/network/api/self/sites",
    "devices": "/proxy/network/api/s/{site}/stat/device",
    "clients": "/proxy/network/api/s/{site}/stat/sta",
    "wlans": "/proxy/network/api/s/{site}/rest/wlanconf",
    "health": "/proxy/network/api/s/{site}/stat/health",
    "alerts": "/proxy/network/api/s/{site}/rest/alarm",
    "events": "/proxy/network/api/s/{site}/rest/event",
    "networks": "/proxy/network/api/s/{site}/rest/networkconf",
    "restart_device": "/proxy/network/api/s/{site}/cmd/devmgr",
    "block_client": "/proxy/network/api/s/{site}/rest/user/{mac}",
    "reconnect_client": "/proxy/network/api/s/{site}/cmd/stamgr",
}

# API Endpoints for standalone controllers (non-UniFi OS)
API_ENDPOINTS_STANDALONE = {
    "login": "/api/login",
    "logout": "/api/logout",
    "sites": "/api/self/sites",
    "devices": "/api/s/{site}/stat/device",
    "clients": "/api/s/{site}/stat/sta",
    "wlans": "/api/s/{site}/rest/wlanconf",
    "health": "/api/s/{site}/stat/health",
    "alerts": "/api/s/{site}/rest/alarm",
    "events": "/api/s/{site}/rest/event",
    "networks": "/api/s/{site}/rest/networkconf",
    "restart_device": "/api/s/{site}/cmd/devmgr",
    "block_client": "/api/s/{site}/rest/user/{mac}",
    "reconnect_client": "/api/s/{site}/cmd/stamgr",
}

# Sensor types
SENSOR_DEVICES_ONLINE = "devices_online"
SENSOR_CLIENTS_CONNECTED = "clients_connected"
SENSOR_GATEWAY_UP_SPEED = "gateway_up_speed"
SENSOR_GATEWAY_DOWN_SPEED = "gateway_down_speed"
SENSOR_CPU_USAGE = "cpu_usage"
SENSOR_MEMORY_USAGE = "memory_usage"
SENSOR_ALERTS_COUNT = "alerts_count"

# Binary sensor types
BINARY_SENSOR_FIRMWARE_UPDATE = "firmware_update_available"

# Device classes
DEVICE_CLASS_CONNECTIVITY = "connectivity"
DEVICE_CLASS_UPDATE = "update"

# Services
SERVICE_RESTART_DEVICE = "restart_device"
SERVICE_BLOCK_CLIENT = "block_client"
SERVICE_RECONNECT_CLIENT = "reconnect_client"

# Service attributes
ATTR_DEVICE_MAC = "device_mac"
ATTR_CLIENT_MAC = "client_mac"
ATTR_BLOCK = "block"

# Update intervals (seconds)
MIN_SCAN_INTERVAL = 10
MAX_SCAN_INTERVAL = 300
