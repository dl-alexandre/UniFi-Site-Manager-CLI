# API Documentation

Complete reference for UniFi Site Manager API endpoints used by the CLI.

## Table of Contents

- [Overview](#overview)
- [Base URLs](#base-urls)
- [Authentication](#authentication)
- [Sites API](#sites-api)
- [Hosts API](#hosts-api)
- [Devices API](#devices-api)
- [Clients API](#clients-api)
- [WLANs API](#wlans-api)
- [Alerts API](#alerts-api)
- [Events API](#events-api)
- [Networks API](#networks-api)
- [Error Responses](#error-responses)
- [Rate Limiting](#rate-limiting)

## Overview

The CLI supports two API modes:

1. **Cloud API**: Site Manager API at `api.ui.com`
2. **Local API**: Direct UniFi OS controller API

## Base URLs

| Mode | Base URL | Notes |
|------|----------|-------|
| Cloud | `https://api.ui.com/v1` | Requires API key |
| Local UniFi OS | `https://{host}/proxy/network/api` | Requires username/password |
| Local Standalone | `https://{host}:8443/api` | Requires username/password |

## Authentication

### Cloud API (API Key)

```http
GET /v1/sites HTTP/1.1
Host: api.ui.com
X-API-Key: your-api-key-here
Accept: application/json
```

### Local API (Session Cookie)

```http
POST /api/auth/login HTTP/1.1
Host: 192.168.1.1
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}

# Response includes cookie:
Set-Cookie: unifises=xxx; csrf_token=xxx
```

Subsequent requests:
```http
GET /proxy/network/api/s/default/stat/device HTTP/1.1
Host: 192.168.1.1
Cookie: unifises=xxx
X-CSRF-Token: xxx
```

## Sites API

### List Sites

**Endpoint**: `GET /v1/sites`

**CLI Command**: `usm sites list`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Max results (default: 50, max: 1000) |
| `offset` | integer | Pagination offset |
| `search` | string | Filter by name/description |

**Response**:
```json
{
  "sites": [
    {
      "id": "60abcdef1234567890abcdef",
      "name": "Main Office",
      "description": "Headquarters location",
      "role": "owner",
      "hosts_count": 1,
      "devices_count": 12,
      "clients_count": 45,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T12:30:00Z"
    }
  ],
  "total": 3,
  "limit": 50,
  "offset": 0
}
```

### Get Site

**Endpoint**: `GET /v1/sites/{site_id}`

**CLI Command**: `usm sites get <site-id>`

**Response**:
```json
{
  "id": "60abcdef1234567890abcdef",
  "name": "Main Office",
  "description": "Headquarters location",
  "role": "owner",
  "timezone": "America/Los_Angeles",
  "hosts": [...],
  "settings": {...}
}
```

### Create Site

**Endpoint**: `POST /v1/sites`

**CLI Command**: `usm sites create <name>`

**Request Body**:
```json
{
  "name": "New Office",
  "description": "New branch location",
  "timezone": "America/Los_Angeles"
}
```

**Response**: Site object with generated ID

### Update Site

**Endpoint**: `PUT /v1/sites/{site_id}`

**CLI Command**: `usm sites update <site-id>`

**Request Body**:
```json
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

### Delete Site

**Endpoint**: `DELETE /v1/sites/{site_id}`

**CLI Command**: `usm sites delete <site-id>`

### Site Health

**Endpoint**: `GET /v1/sites/{site_id}/health`

**CLI Command**: `usm sites health <site-id>`

**Response**:
```json
{
  "status": "healthy",
  "devices": {
    "total": 12,
    "online": 11,
    "offline": 1,
    "pending": 0
  },
  "wan": {
    "status": "up",
    "latency_ms": 15
  },
  "lan": {
    "status": "healthy"
  },
  "wlan": {
    "status": "healthy",
    "aps_online": 4,
    "aps_offline": 0
  }
}
```

### Site Statistics

**Endpoint**: `GET /v1/sites/{site_id}/stats`

**CLI Command**: `usm sites stats <site-id>`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `period` | string | `day`, `week`, `month` |

**Response**:
```json
{
  "period": "day",
  "clients": {
    "total": 45,
    "wireless": 32,
    "wired": 13,
    "guest": 5
  },
  "traffic": {
    "rx_bytes": 10737418240,
    "tx_bytes": 5368709120,
    "rx_rate": 104857600,
    "tx_rate": 52428800
  },
  "uptime": 86400
}
```

## Hosts API

### List Hosts

**Endpoint**: `GET /v1/hosts`

**CLI Command**: `usm hosts list`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Max results |
| `offset` | integer | Pagination offset |
| `search` | string | Filter by name |

**Response**:
```json
{
  "hosts": [
    {
      "id": "60abcdef1234567890abcdef",
      "name": "UDM-Pro Main",
      "model": "UDMPRO",
      "version": "2.4.0",
      "ip_address": "192.168.1.1",
      "status": "online",
      "sites_count": 3
    }
  ]
}
```

### Get Host

**Endpoint**: `GET /v1/hosts/{host_id}`

**CLI Command**: `usm hosts get <host-id>`

### Restart Host

**Endpoint**: `POST /v1/hosts/{host_id}/restart`

**CLI Command**: `usm hosts restart <host-id>`

### Host Health

**Endpoint**: `GET /v1/hosts/{host_id}/health`

**CLI Command**: `usm hosts health <host-id>`

**Response**:
```json
{
  "status": "online",
  "cpu": {
    "usage_percent": 25
  },
  "memory": {
    "total_bytes": 4294967296,
    "used_bytes": 2147483648,
    "usage_percent": 50
  },
  "storage": {
    "total_bytes": 107374182400,
    "used_bytes": 53687091200,
    "usage_percent": 50
  }
}
```

### Host Statistics

**Endpoint**: `GET /v1/hosts/{host_id}/stats`

**CLI Command**: `usm hosts stats <host-id>`

## Devices API

### List Devices

**Endpoint**: `GET /v1/sites/{site_id}/devices`

**CLI Command**: `usm devices list <site-id>`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Max results |
| `offset` | integer | Pagination offset |
| `status` | string | Filter by status: `online`, `offline`, `pending` |
| `type` | string | Filter by type: `ap`, `switch`, `gateway` |

**Response**:
```json
{
  "devices": [
    {
      "id": "60abcdef1234567890abcdef",
      "name": "Office AP 1",
      "mac": "aa:bb:cc:dd:ee:ff",
      "type": "ap",
      "model": "U6-Pro",
      "status": "online",
      "version": "6.2.0",
      "ip_address": "192.168.1.10",
      "clients_count": 15,
      "upgrade_available": false
    }
  ]
}
```

### Get Device

**Endpoint**: `GET /v1/sites/{site_id}/devices/{device_id}`

**CLI Command**: `usm devices get <site-id> <device-id>`

### Restart Device

**Endpoint**: `POST /v1/sites/{site_id}/devices/{device_id}/restart`

**CLI Command**: `usm devices restart <site-id> <device-id>`

### Upgrade Firmware

**Endpoint**: `POST /v1/sites/{site_id}/devices/{device_id}/upgrade`

**CLI Command**: `usm devices upgrade <site-id> <device-id>`

### Adopt Device

**Endpoint**: `POST /v1/sites/{site_id}/devices/adopt`

**CLI Command**: `usm devices adopt <site-id> <mac-address>`

**Request Body**:
```json
{
  "mac": "aa:bb:cc:dd:ee:ff"
}
```

## Clients API

### List Clients

**Endpoint**: `GET /v1/sites/{site_id}/clients`

**CLI Command**: `usm clients list <site-id>`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Max results |
| `type` | string | Filter: `wired`, `wireless` |
| `search` | string | Search by name, MAC, or IP |

**Response**:
```json
{
  "clients": [
    {
      "mac": "aa:bb:cc:dd:ee:ff",
      "name": "Johns-iPhone",
      "hostname": "johns-iphone",
      "ip": "192.168.1.100",
      "type": "wireless",
      "ssid": "Office-WiFi",
      "ap_mac": "11:22:33:44:55:66",
      "signal": -65,
      "rssi": -65,
      "noise": -90,
      "tx_rate": 866700,
      "rx_rate": 780000,
      "uptime": 3600,
      "blocked": false
    }
  ]
}
```

### Get Client Statistics

**Endpoint**: `GET /v1/sites/{site_id}/clients/{mac}/stats`

**CLI Command**: `usm clients stats <site-id> <mac>`

**Response**:
```json
{
  "mac": "aa:bb:cc:dd:ee:ff",
  "traffic": {
    "rx_bytes": 104857600,
    "tx_bytes": 52428800,
    "rx_packets": 100000,
    "tx_packets": 50000
  },
  "signal_history": [...],
  "uptime": 86400
}
```

### Block Client

**Endpoint**: `POST /v1/sites/{site_id}/clients/{mac}/block`

**CLI Command**: `usm clients block <site-id> <mac>`

### Unblock Client

**Endpoint**: `POST /v1/sites/{site_id}/clients/{mac}/unblock`

**CLI Command**: `usm clients unblock <site-id> <mac>`

## WLANs API

### List WLANs

**Endpoint**: `GET /v1/sites/{site_id}/wlans`

**CLI Command**: `usm wlans list <site-id>`

**Response**:
```json
{
  "wlans": [
    {
      "id": "60abcdef1234567890abcdef",
      "name": "Office WiFi",
      "ssid": "Office-5G",
      "enabled": true,
      "security": "wpapsk",
      "wpa_mode": "wpa2",
      "wpa3": false,
      "vlan": 10,
      "band": "both",
      "hide_ssid": false,
      "is_guest": false
    }
  ]
}
```

### Get WLAN

**Endpoint**: `GET /v1/sites/{site_id}/wlans/{wlan_id}`

**CLI Command**: `usm wlans get <site-id> <wlan-id>`

### Create WLAN

**Endpoint**: `POST /v1/sites/{site_id}/wlans`

**CLI Command**: `usm wlans create <site-id> <name> <ssid>`

**Request Body**:
```json
{
  "name": "Guest WiFi",
  "ssid": "Guest-Network",
  "enabled": true,
  "security": "wpapsk",
  "wpa_mode": "wpa2",
  "wpa3": false,
  "password": "guest-password",
  "vlan": 20,
  "band": "both",
  "hide_ssid": false,
  "is_guest": true,
  "user_group_id": ""
}
```

### Update WLAN

**Endpoint**: `PUT /v1/sites/{site_id}/wlans/{wlan_id}`

**CLI Command**: `usm wlans update <site-id> <wlan-id>`

### Delete WLAN

**Endpoint**: `DELETE /v1/sites/{site_id}/wlans/{wlan_id}`

**CLI Command**: `usm wlans delete <site-id> <wlan-id>`

## Alerts API

### List Alerts

**Endpoint**: `GET /v1/alerts`

**CLI Command**: `usm alerts list`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `site_id` | string | Filter by site |
| `archived` | boolean | Include archived |

**Response**:
```json
{
  "alerts": [
    {
      "id": "60abcdef1234567890abcdef",
      "site_id": "60fedcba0987654321fedcba",
      "type": "device_offline",
      "severity": "high",
      "message": "Device Office AP 1 is offline",
      "device_mac": "aa:bb:cc:dd:ee:ff",
      "timestamp": "2024-01-01T12:00:00Z",
      "acknowledged": false,
      "archived": false
    }
  ]
}
```

### Acknowledge Alert

**Endpoint**: `POST /v1/sites/{site_id}/alerts/{alert_id}/ack`

**CLI Command**: `usm alerts ack <site-id> <alert-id>`

### Archive Alert

**Endpoint**: `POST /v1/sites/{site_id}/alerts/{alert_id}/archive`

**CLI Command**: `usm alerts archive <site-id> <alert-id>`

## Events API

### List Events

**Endpoint**: `GET /v1/events`

**CLI Command**: `usm events list`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `site_id` | string | Filter by site |
| `limit` | integer | Max results |
| `start` | string | Start time (ISO 8601) |
| `end` | string | End time (ISO 8601) |

**Response**:
```json
{
  "events": [
    {
      "id": "60abcdef1234567890abcdef",
      "site_id": "60fedcba0987654321fedcba",
      "type": "device_connected",
      "severity": "info",
      "message": "Device Office AP 1 connected",
      "timestamp": "2024-01-01T12:00:00Z",
      "device_mac": "aa:bb:cc:dd:ee:ff"
    }
  ]
}
```

## Networks API

### List Networks

**Endpoint**: `GET /v1/sites/{site_id}/networks`

**CLI Command**: `usm networks list <site-id>`

**Response**:
```json
{
  "networks": [
    {
      "id": "60abcdef1234567890abcdef",
      "name": "Default",
      "purpose": "corporate",
      "vlan": 1,
      "subnet": "192.168.1.0/24",
      "dhcpd_enabled": true,
      "dhcpd_start": "192.168.1.6",
      "dhcpd_stop": "192.168.1.254"
    }
  ]
}
```

## Error Responses

### Common Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Invalid or missing API key |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |

### Error Response Format

```json
{
  "error": {
    "code": "validation_error",
    "message": "Invalid site ID format",
    "details": {
      "field": "site_id",
      "value": "invalid-id"
    }
  }
}
```

## Rate Limiting

The Site Manager API implements rate limiting:

- **Default**: 100 requests per minute per API key
- **Burst**: 20 requests per second
- **Headers**:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset timestamp
  - `Retry-After`: Seconds to wait (on 429)

### Handling Rate Limits

The CLI automatically implements exponential backoff when rate limited. For scripts, add delays:

```bash
# Add delay between requests
usm sites list
sleep 1
usm devices list $SITE_ID
sleep 1
```

## Local API Differences

When using `--local` mode, the API paths differ:

| Resource | Cloud Path | Local Path |
|----------|-----------|-----------|
| Sites | `/v1/sites` | `/api/s/{site}/self` |
| Devices | `/v1/sites/{id}/devices` | `/api/s/{site}/stat/device` |
| Clients | `/v1/sites/{id}/clients` | `/api/s/{site}/stat/sta` |
| WLANs | `/v1/sites/{id}/wlans` | `/api/s/{site}/rest/wlanconf` |

### Local API Authentication Flow

1. Login: `POST /api/auth/login` → Session cookie
2. All requests include:
   - `Cookie: unifises={token}`
   - `X-CSRF-Token: {token}`

## API Versioning

Current API version: **v1**

The API follows semantic versioning. Breaking changes will increment the major version.

## Official Documentation

For the most up-to-date API reference, visit:
- [Site Manager API Documentation](https://developer.ui.com/site-manager/v1.0.0/gettingstarted)

## Next Steps

- See [Usage Guide](USAGE.md) for CLI commands
- Read [Configuration](CONFIGURATION.md) for authentication setup
- Check [FAQ](FAQ.md) for common questions
