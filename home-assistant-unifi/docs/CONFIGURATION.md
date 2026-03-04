# Configuration Reference

Complete configuration guide for UniFi Controller Home Assistant integration.

## Table of Contents

- [Configuration Methods](#configuration-methods)
- [Configuration Options](#configuration-options)
- [UI Configuration](#ui-configuration)
- [YAML Configuration](#yaml-configuration)
- [Options Flow](#options-flow)
- [Advanced Configuration](#advanced-configuration)
- [Troubleshooting](#troubleshooting)

## Configuration Methods

### UI Configuration (Recommended)

All configuration is done through Home Assistant UI:

1. **Settings** → **Devices & Services**
2. Click **+ Add Integration**
3. Search "UniFi Controller"
4. Fill in the form
5. Submit

### YAML Configuration (Legacy)

Available but not recommended:

```yaml
# configuration.yaml
unifi_controller:
  host: 192.168.1.1
  port: 443
  username: admin
  password: !secret unifi_password
  site_id: default
  unifi_os: true
  verify_ssl: false
  scan_interval: 30
```

## Configuration Options

### Required Options

| Option | Description | Example |
|--------|-------------|---------|
| **Host** | Controller IP or hostname | `192.168.1.1` |
| **Username** | Local admin username | `admin` |
| **Password** | Account password | `SecurePass123!` |

### Optional Options

| Option | Default | Description |
|--------|---------|-------------|
| **Port** | `443` (UniFi OS) / `8443` (Standalone) | Controller port |
| **Site ID** | `default` | Site identifier |
| **UniFi OS** | `false` | Enable for UDM/UDR/Cloud Key |
| **Verify SSL** | `false` | Verify TLS certificate |
| **Scan Interval** | `30` | Seconds between updates |

## UI Configuration

### Initial Setup

1. **Add Integration**
   ```
   Settings → Devices & Services → + Add Integration
   ```

2. **Search and Select**
   - Type: "UniFi Controller"
   - Click on the result

3. **Configure Connection**
   
   **For UniFi OS (UDM/UDR/Cloud Key)**:
   ```
   Host: 192.168.1.1
   Port: 443
   Username: admin
   Password: ••••••••
   Site ID: default
   UniFi OS: ✅ Checked
   Verify SSL: ❌ Unchecked (for self-signed certs)
   ```
   
   **For Standalone Controller**:
   ```
   Host: 192.168.1.100
   Port: 8443
   Username: admin
   Password: ••••••••
   Site ID: default
   UniFi OS: ❌ Unchecked
   Verify SSL: ❌ Unchecked
   ```

4. **Test Connection**
   - Integration validates credentials
   - Discovers sites
   - Creates entities

### Multiple Sites

Add multiple instances for different sites:

1. **Settings** → **Devices & Services**
2. **+ Add Integration** → "UniFi Controller"
3. Configure different Site ID or controller

## YAML Configuration

### Basic Setup

```yaml
# configuration.yaml

# Secrets in secrets.yaml
unifi_controller:
  host: 192.168.1.1
  port: 443
  username: admin
  password: !secret unifi_password
  site_id: default
  unifi_os: true
  verify_ssl: false
```

### Advanced Setup

```yaml
# configuration.yaml

unifi_controller:
  host: 192.168.1.1
  port: 443
  username: admin
  password: !secret unifi_password
  site_id: default
  unifi_os: true
  verify_ssl: false
  scan_interval: 60  # Update every 60 seconds
  
  # Entity filtering
  monitored_conditions:
    - devices_online
    - clients_connected
    - cpu_usage
    - memory_usage
    - alerts
```

### secrets.yaml

```yaml
# secrets.yaml
unifi_password: "YourSecurePassword123!"
```

**Important**: Never commit secrets.yaml to version control!

Add to `.gitignore`:
```
secrets.yaml
```

### Multiple Controllers

```yaml
# configuration.yaml

unifi_controller:
  - host: 192.168.1.1
    port: 443
    username: admin
    password: !secret unifi_password_main
    site_id: default
    unifi_os: true
  
  - host: 192.168.2.1
    port: 443
    username: admin
    password: !secret unifi_password_branch
    site_id: branch_office
    unifi_os: true
```

## Options Flow

After initial setup, reconfigure anytime:

### Accessing Options

1. **Settings** → **Devices & Services**
2. Find **"UniFi Controller"** card
3. Click **Configure**

### Available Options

| Option | Default | Description |
|--------|---------|-------------|
| **Scan Interval** | 30 | Seconds between data updates |
| **Enable Device Trackers** | true | Create device trackers for clients |
| **Track Wired Clients** | true | Include wired clients |
| **Track Wireless Clients** | true | Include wireless clients |
| **Block Clients** | true | Enable blocking service |
| **Disable SSL Verification** | true | Skip SSL verification |
| **Debug Logging** | false | Enable debug output |

### Reconfiguring

```
1. Click Configure
2. Change desired options
3. Click Submit
4. Changes apply immediately
```

## Advanced Configuration

### Custom Update Intervals

```yaml
# Different intervals for different data types
unifi_controller:
  host: 192.168.1.1
  
  # Fast updates for critical sensors
  scan_interval: 10
  
  # Slow updates for device trackers
  device_tracker_scan_interval: 60
```

### Entity Filtering

```yaml
# Only create specific entity types
unifi_controller:
  host: 192.168.1.1
  
  # Disable device trackers
  device_trackers: false
  
  # Only monitor specific MACs
  tracked_devices:
    - "aa:bb:cc:dd:ee:ff"
    - "11:22:33:44:55:66"
  
  # Ignore guest clients
  track_guests: false
```

### SSL/TLS Configuration

```yaml
# For custom certificates
unifi_controller:
  host: 192.168.1.1
  verify_ssl: true
  
  # Custom CA certificate
  ssl_ca: /config/ssl/ca.crt
  
  # Or disable verification
  verify_ssl: false
```

### Performance Tuning

```yaml
# Reduce load for large networks
unifi_controller:
  host: 192.168.1.1
  
  # Less frequent updates
  scan_interval: 120
  
  # Batch updates
  batch_updates: true
  
  # Limit device trackers
  max_device_trackers: 50
  
  # Exclude by MAC prefix
  exclude_macs:
    - "00:00:00"  # Virtual machines
```

### Security Hardening

```yaml
unifi_controller:
  host: 192.168.1.1
  
  # Use dedicated service account
  username: homeassistant
  password: !secret ha_unifi_password
  
  # Read-only account recommended
  # Create in UniFi: Role = Read Only Admin
```

## Configuration Examples

### Example 1: Home Network (UDM)

```yaml
# Home setup with UDM
unifi_controller:
  host: 192.168.1.1
  port: 443
  username: ha_service
  password: !secret udm_password
  site_id: default
  unifi_os: true
  verify_ssl: false
```

### Example 2: Office Network (Standalone)

```yaml
# Office with standalone controller
unifi_controller:
  host: 192.168.10.5
  port: 8443
  username: admin
  password: !secret office_unifi_pass
  site_id: main_office
  unifi_os: false
  verify_ssl: false
  scan_interval: 60
```

### Example 3: Multiple Sites

```yaml
# Multiple site monitoring
unifi_controller:
  - host: 192.168.1.1
    port: 443
    username: admin
    password: !secret site1_pass
    site_id: headquarters
    unifi_os: true
    
  - host: 192.168.2.1
    port: 443
    username: admin
    password: !secret site2_pass
    site_id: branch_east
    unifi_os: true
    
  - host: 192.168.3.1
    port: 443
    username: admin
    password: !secret site3_pass
    site_id: branch_west
    unifi_os: true
```

### Example 4: Minimal Setup (Sensors Only)

```yaml
# Minimal - only sensors, no device trackers
unifi_controller:
  host: 192.168.1.1
  username: admin
  password: !secret unifi_pass
  unifi_os: true
  device_trackers: false  # Disable all trackers
```

### Example 5: Maximum Entities

```yaml
# Complete monitoring
unifi_controller:
  host: 192.168.1.1
  username: admin
  password: !secret unifi_pass
  unifi_os: true
  
  # Enable everything
  device_trackers: true
  track_wired: true
  track_wireless: true
  track_guests: true
  
  # All sensors
  monitored_conditions:
    - all
```

## Troubleshooting

### Configuration Not Saving

**Symptoms**: Changes don't persist

**Solutions**:
1. Check YAML syntax: `ha core check`
2. Restart Home Assistant
3. Check file permissions

### Options Not Applying

**Symptoms**: Changed options don't take effect

**Solutions**:
1. Reload integration:
   - Settings → Integration → ⋮ → Reload
2. Restart Home Assistant
3. Check logs for errors

### Invalid Configuration

**Symptoms**: "Invalid config" error

**Solutions**:
1. Validate YAML:
   ```bash
   ha core check
   ```
2. Check configuration.yaml syntax
3. Verify secrets exist in secrets.yaml

### Connection Refused

**Symptoms**: Cannot connect to controller

**Solutions**:
1. Verify host and port
2. Check UniFi OS checkbox matches controller type
3. Test connectivity:
   ```bash
   ping 192.168.1.1
   curl -k https://192.168.1.1
   ```
4. Check firewall rules

### Authentication Failed

**Symptoms**: "Invalid credentials"

**Solutions**:
1. Verify username/password
2. Ensure local account (not SSO)
3. Check account has admin privileges
4. Try logging into UniFi web UI directly

## Migration Guide

### From YAML to UI Configuration

1. **Note current settings** from configuration.yaml
2. **Remove YAML config**:
   ```yaml
   # Remove this section
   # unifi_controller:
   #   host: ...
   ```
3. **Restart Home Assistant**
4. **Add via UI**:
   - Settings → Devices & Services
   - + Add Integration
   - Re-enter settings

### From Old Integration

If migrating from another UniFi integration:
1. **Remove old integration**
2. **Restart Home Assistant**
3. **Add new integration**
4. **Update automations** to use new entity names

## Next Steps

After configuration:
- [Usage Guide](USAGE.md) - How to use the integration
- [API Documentation](API.md) - Technical details
- [FAQ](FAQ.md) - Common questions
- Create dashboards and automations!
