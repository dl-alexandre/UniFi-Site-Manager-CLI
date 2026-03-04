# Installation Guide

Complete installation instructions for UniFi Controller Home Assistant integration.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
- [Configuration](#configuration)
- [Post-Installation](#post-installation)
- [Updating](#updating)
- [Uninstallation](#uninstallation)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Requirements

- **Home Assistant**: 2023.1.0 or newer
- **UniFi Controller**: 
  - UDM/UDR/UDM-Pro/UDM-SE (UniFi OS)
  - Cloud Key Gen2+
  - Standalone Controller 7.x+
- **Network**: Home Assistant must reach controller on port 443/8443
- **Account**: Local admin account (not SSO/cloud)

### Verify Home Assistant Version

```yaml
# Developer Tools → Info, or configuration.yaml
homeassistant:
  # Check version in logs or UI
```

### Verify Controller Access

```bash
# From Home Assistant terminal
ping 192.168.1.1  # Your controller IP
curl -k https://192.168.1.1  # Should return UniFi login page
```

## Installation Methods

### Method 1: HACS (Recommended)

HACS (Home Assistant Community Store) is the easiest way to install custom integrations.

#### Step 1: Install HACS

If not already installed:
1. Follow [HACS installation guide](https://hacs.xyz/docs/setup/download)
2. Or use [HACS install script](https://hacs.xyz/docs/setup/prereqs)

#### Step 2: Add Custom Repository

1. Open Home Assistant
2. Go to **HACS** → **Integrations**
3. Click **⋮ (menu)** → **Custom repositories**
4. Add:
   - Repository: `https://github.com/dl-alexandre/Local-UniFi-CLI`
   - Category: **Integration**
5. Click **Add**

#### Step 3: Install Integration

1. In HACS Integrations, click **+ Explore & Download Repositories**
2. Search for "UniFi Controller"
3. Click **Download**
4. Restart Home Assistant:
   - **Settings** → **System** → **Restart**

#### Step 4: Configure

1. **Settings** → **Devices & Services**
2. Click **+ Add Integration**
3. Search for "UniFi Controller"
4. Enter your controller details
5. Click **Submit**

### Method 2: Manual Installation

For users who prefer manual installation or cannot use HACS.

#### Step 1: Download Files

```bash
# Clone repository
git clone https://github.com/dl-alexandre/Local-UniFi-CLI.git

# Or download ZIP from GitHub
# https://github.com/dl-alexandre/Local-UniFi-CLI/archive/refs/heads/main.zip
```

#### Step 2: Copy Files

```bash
# SSH into Home Assistant (or use Terminal add-on)
# Create directory
mkdir -p /config/custom_components/unifi_controller

# Copy files
cp -r /path/to/Local-UniFi-CLI/home-assistant-unifi/custom_components/unifi_controller/* \
      /config/custom_components/unifi_controller/

# Verify
ls -la /config/custom_components/unifi_controller/
# Should show: __init__.py, config_flow.py, const.py, etc.
```

#### Step 3: Restart Home Assistant

```bash
# From terminal
ha core restart

# Or via UI: Settings → System → Restart
```

#### Step 4: Configure

1. **Settings** → **Devices & Services**
2. Click **+ Add Integration**
3. Search "UniFi Controller"
4. Complete setup

### Method 3: Terminal & SSH Add-on

For Home Assistant OS/Supervised installations:

#### Using Terminal Add-on

1. Install **Terminal & SSH** add-on from add-on store
2. Start the add-on
3. Open Web UI or SSH to HA

```bash
# In terminal
cd /config/custom_components

# Download latest release
wget https://github.com/dl-alexandre/Local-UniFi-CLI/releases/latest/download/unifi_controller.zip

# Extract
unzip unifi_controller.zip -d unifi_controller/
rm unifi_controller.zip

# Fix permissions
chmod -R 755 unifi_controller/

# Restart HA
ha core restart
```

### Method 4: Samba/Network Share

If using Samba share to access HA config:

1. Mount `\<HA-IP>\config` share
2. Navigate to `custom_components` folder
3. Create `unifi_controller` folder
4. Copy all integration files
5. Restart Home Assistant

## Configuration

### Initial Setup

After installation:

1. **Settings** → **Devices & Services**
2. Click **+ Add Integration**
3. Search for and select **"UniFi Controller"**

### Configuration Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| **Host** | Yes | - | Controller IP or hostname |
| **Port** | Yes | 443/8443 | Controller port |
| **Username** | Yes | - | Local admin username |
| **Password** | Yes | - | Account password |
| **Site ID** | Yes | default | Site identifier |
| **UniFi OS** | Yes | false | Check for UDM/UDR/Cloud Key |
| **Verify SSL** | Yes | false | Verify TLS certificate |

### Configuration Examples

#### UniFi Dream Machine (UDM)

```
Host: 192.168.1.1
Port: 443
Username: admin
Password: your-password
Site ID: default
UniFi OS: ✅ (checked)
Verify SSL: ❌ (unchecked)
```

#### UniFi Cloud Key Gen2+

```
Host: 192.168.1.10
Port: 443
Username: admin
Password: your-password
Site ID: default
UniFi OS: ✅ (checked)
Verify SSL: ❌ (unchecked)
```

#### Standalone Controller

```
Host: 192.168.1.100
Port: 8443
Username: admin
Password: your-password
Site ID: default
UniFi OS: ❌ (unchecked)
Verify SSL: ❌ (unchecked)
```

### Finding Your Site ID

1. Log into controller web UI
2. URL shows: `https://192.168.1.1/network/default/dashboard`
3. Site ID is `default` (or `/abc123/dashboard` → site ID is `abc123`)

### Creating a Local Account

The integration requires a local admin account (not Ubiquiti SSO):

1. Log into controller: `https://192.168.1.1`
2. **Settings** → **Admins**
3. Click **+ Invite Admin**
4. Enter:
   - Name: `Home Assistant`
   - Username: `hass`
   - Password: (generate strong password)
5. Role: **Administrator**
6. **Save**

## Post-Installation

### Verify Installation

1. **Settings** → **Devices & Services**
2. Look for **"UniFi Controller"** card
3. Should show:
   - Number of entities
   - Configuration button
   - System options

### Check Entities

1. **Settings** → **Devices & Services**
2. Click **Entities** tab
3. Filter by "unifi"
4. Verify sensors appear:
   - `sensor.unifi_devices_online`
   - `sensor.unifi_clients_connected`
   - etc.

### Add to Dashboard

```yaml
# Example card
type: entities
title: Network Status
entities:
  - sensor.unifi_devices_online
  - sensor.unifi_clients_connected
  - sensor.unifi_cpu_usage
```

### Enable/Disable Entities

1. **Settings** → **Devices & Services**
2. Click **Entities**
3. Filter "unifi"
4. Select entity
5. Click ⚙️ (settings)
6. Toggle **Enabled**

## Updating

### Update via HACS

1. **HACS** → **Integrations**
2. Find "UniFi Controller"
3. Click **Update** if available
4. Restart Home Assistant

### Update Manually

```bash
# Backup current version
cd /config/custom_components
cp -r unifi_controller unifi_controller.backup

# Download new version
wget https://github.com/dl-alexandre/Local-UniFi-CLI/releases/latest/download/unifi_controller.zip

# Replace files
rm -rf unifi_controller/*
unzip unifi_controller.zip -d unifi_controller/
rm unifi_controller.zip

# Restart
ha core restart
```

### Check for Updates

Integration version shown in:
- **Settings** → **Devices & Services** → UniFi Controller
- Or check HACS for available updates

## Uninstallation

### Remove via HACS

1. **HACS** → **Integrations**
2. Find "UniFi Controller"
3. Click **⋮** → **Uninstall**
4. Restart Home Assistant

### Remove Manually

```bash
# SSH into Home Assistant
rm -rf /config/custom_components/unifi_controller/

# Restart
ha core restart
```

### Clean Up Entities

1. **Settings** → **Devices & Services**
2. Find "UniFi Controller" card
3. Click **⋮** → **Delete**
4. Confirm deletion

### Remove Device Trackers

If device trackers remain after uninstall:

1. **Settings** → **People & Zones**
2. Click **Devices** tab
3. Find and delete stale trackers

## Troubleshooting

### Integration Not Found

**Symptoms**: "UniFi Controller" not in integration list

**Solutions**:
1. Verify files copied to correct location:
   ```bash
   ls /config/custom_components/unifi_controller/
   ```
2. Check for missing `__init__.py`
3. Restart Home Assistant again
4. Clear browser cache, hard refresh

### Setup Fails

**Symptoms**: "Failed to connect" or "Invalid credentials"

**Checklist**:
- [ ] Host IP/hostname is correct
- [ ] Port is correct (443 for UniFi OS, 8443 for standalone)
- [ ] Using local account (not Ubiquiti cloud account)
- [ ] Username and password are correct
- [ ] "UniFi OS" checkbox matches your controller type
- [ ] Try with "Verify SSL" unchecked
- [ ] Controller firmware is up to date
- [ ] Home Assistant can reach controller IP (ping test)

### No Entities Created

**Symptoms**: Integration configured but no entities

**Solutions**:
1. Check Home Assistant logs:
   ```yaml
   # configuration.yaml
   logger:
     logs:
       custom_components.unifi_controller: debug
   ```
2. Verify Site ID is correct
3. Check controller has devices/clients
4. Reload integration: Settings → Integration → ⋮ → Reload

### High CPU/Memory Usage

**Symptoms**: HA slow or unresponsive

**Solutions**:
1. Reduce polling frequency:
   - Settings → Integration → Configure → Scan interval
2. Disable unused entities
3. Restart Home Assistant
4. Check for duplicate integrations

### Authentication Errors

**Symptoms**: Repeated auth failures in logs

**Solutions**:
1. Create new local admin account
2. Avoid special characters in password
3. Check account isn't locked
4. Verify not using SSO/cloud account

### SSL/TLS Errors

**Symptoms**: SSL certificate errors

**Solutions**:
1. Disable "Verify SSL" in configuration
2. Or install controller's certificate in Home Assistant

### Diagnostic Steps

```bash
# Test controller connectivity from HA
ping 192.168.1.1

# Test API access
curl -k -u "admin:password" \
  https://192.168.1.1/proxy/network/api/s/default/self

# Check Home Assistant logs
cat /config/home-assistant.log | grep unifi

# Enable debug logging and restart
echo "logger:\n  logs:\n    custom_components.unifi_controller: debug" >> /config/configuration.yaml
ha core restart
```

## Support

For installation issues:
- [GitHub Issues](https://github.com/dl-alexandre/Local-UniFi-CLI/issues)
- [Home Assistant Community Forum](https://community.home-assistant.io/)
- [UniFi Community Forums](https://community.ui.com/)

## Next Steps

After successful installation:
- [Usage Guide](USAGE.md) - How to use the integration
- [API Documentation](API.md) - Technical details
- [FAQ](FAQ.md) - Common questions
- Create dashboards and automations!
