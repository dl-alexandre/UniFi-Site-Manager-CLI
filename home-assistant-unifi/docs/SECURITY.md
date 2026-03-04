# Security Best Practices

Security guidelines for using UniFi Controller Home Assistant integration.

## Table of Contents

- [Overview](#overview)
- [Account Security](#account-security)
- [Network Security](#network-security)
- [Home Assistant Security](#home-assistant-security)
- [Credential Management](#credential-management)
- [SSL/TLS](#ssltls)
- [Best Practices](#best-practices)
- [Vulnerability Reporting](#vulnerability-reporting)

## Overview

This integration connects to your UniFi controller using local API calls. Understanding and implementing security best practices is essential to protect your network.

### Security Principles

1. **Least Privilege**: Use accounts with minimal required permissions
2. **Defense in Depth**: Multiple layers of security
3. **Secure by Default**: Safe configurations out of the box
4. **Regular Updates**: Keep software current

## Account Security

### Create a Dedicated Service Account

**Don't** use your personal admin account:

```yaml
# ❌ Bad
username: admin  # Your personal account
password: YourPersonalPassword123!
```

**Do** create a dedicated Home Assistant account:

1. Log into UniFi controller
2. Settings → Admins → + Invite Admin
3. Enter:
   - Name: `Home Assistant`
   - Username: `ha_service`
   - Password: Randomly generated strong password
4. Role: **Administrator** (or Read-Only for monitoring only)

### Use Strong Passwords

Generate strong password:
```bash
# macOS
openssl rand -base64 24

# Linux
pwgen -s 24 1

# Or use password manager
```

Password requirements:
- Minimum 16 characters
- Mix of uppercase, lowercase, numbers, symbols
- No dictionary words
- Unique (not used elsewhere)

### Rotate Passwords Regularly

Set calendar reminder every 90 days:
1. Generate new password
2. Update integration configuration
3. Update password in UniFi controller
4. Verify integration still works
5. Revoke old password

## Network Security

### Network Segmentation

Isolate controller management:

```
┌─────────────────┐
│ Management VLAN │ 10
│ 192.168.10.0/24 │
│                 │
│ • Home Assistant│
│ • Admin PCs     │
└────────┬────────┘
         │
    ┌────┴────┐
    │ Gateway │
    └────┬────┘
         │
┌────────┴────────┐
│ User VLANs        │ 20, 30, 40
│ • LAN             │
│ • IoT             │
│ • Guest           │
└─────────────────┘
```

### Firewall Rules

Allow only necessary traffic:

```bash
# Allow HA to reach controller
iptables -A INPUT -p tcp -s 192.168.10.5 --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j DROP

# Allow WebSocket (same port)
# Already covered by above rule

# Block management from other VLANs
iptables -A INPUT -p tcp -s 192.168.20.0/24 --dport 443 -j DROP
iptables -A INPUT -p tcp -s 192.168.30.0/24 --dport 443 -j DROP
```

### Disable Remote Access (If Not Needed)

In UniFi controller:
1. Settings → System → Advanced
2. Disable "Remote Access"
3. Use VPN for remote management instead

## Home Assistant Security

### Secure Home Assistant

```yaml
# configuration.yaml

# Enable HTTPS
http:
  ssl_certificate: /ssl/fullchain.pem
  ssl_key: /ssl/privkey.pem
  ip_ban_enabled: true
  login_attempts_threshold: 5

# Require authentication for API
api:
  auth_required: true
```

### Protect Secrets

**Use secrets.yaml**:
```yaml
# ❌ Bad - Never do this
configuration.yaml:
  unifi_controller:
    password: "SecretPassword123!"

# ✅ Good
configuration.yaml:
  unifi_controller:
    password: !secret unifi_password

secrets.yaml:
  unifi_password: "SecretPassword123!"
```

**Add to .gitignore**:
```
secrets.yaml
.uuid
.env
```

### File Permissions

```bash
# Set proper permissions
chmod 600 /config/secrets.yaml
chmod 700 /config/custom_components/unifi_controller/
chown -R homeassistant:homeassistant /config/
```

### Enable 2FA

Enable two-factor authentication on Home Assistant:
1. User profile → Enable 2FA
2. Use authenticator app
3. Store backup codes securely

## Credential Management

### Password Managers

Store credentials in password manager:

**1Password**:
```bash
# Get password for automation
op read "op://Private/UniFi Controller/password"
```

**Bitwarden**:
```bash
bw get password "UniFi Controller"
```

**HashiCorp Vault**:
```bash
vault kv get -field=password secret/homeassistant/unifi
```

### Environment Variables

```bash
# For Docker deployments
export UNIFI_PASSWORD="$(cat /run/secrets/unifi_password)"

# In docker-compose
environment:
  - UNIFI_PASSWORD_FILE=/run/secrets/unifi_password
```

### Home Assistant Secrets

```yaml
# Option 1: secrets.yaml (recommended for simple setups)
unifi_password: "SecurePassword123"

# Option 2: Command line (for dynamic retrieval)
shell_command:
  get_unifi_password: 'cat /secure/password/file'

# Then use in template
unifi_controller:
  password: "{{ states('sensor.unifi_password') }}"
```

## SSL/TLS

### Certificate Verification

**Default**: Disabled (for self-signed certs)

**Enable only if**:
- Using valid certificate
- Configured custom domain
- Using reverse proxy with valid SSL

```yaml
unifi_controller:
  verify_ssl: true  # Only with valid certificate
```

### Install Valid Certificate

On UniFi controller:
1. Get certificate from Let's Encrypt or CA
2. Install in controller settings
3. Update integration to verify SSL

### Reverse Proxy SSL

If using reverse proxy (nginx, traefik):
```yaml
# configuration.yaml
http:
  use_x_forwarded_for: true
  trusted_proxies:
    - 172.30.33.0/24  # Docker network
```

## Best Practices

### Regular Security Audits

Checklist (monthly):
- [ ] Review integration logs for anomalies
- [ ] Verify no unauthorized access
- [ ] Check for entity/entity ID changes
- [ ] Confirm password rotation schedule
- [ ] Review automation/service calls
- [ ] Check for integration updates

### Monitoring

```yaml
# Alert on failed login attempts
automation:
  - alias: "UniFi Login Alert"
    trigger:
      - platform: state
        entity_id: persistent_notification.httplogin
    condition:
      - condition: template
        value_template: '{{ "UniFi Controller" in trigger.to_state.state }}'
    action:
      - service: notify.mobile_app_phone
        data:
          message: "Suspicious UniFi Controller activity detected"
```

### Backup Strategy

Regular backups include:
- Home Assistant configuration
- Secrets (encrypted)
- UniFi controller configuration

```bash
# Automated backup script
#!/bin/bash
date=$(date +%Y%m%d)
tar -czf "/backup/ha-config-$date.tar.gz" /config/
# Encrypt with GPG
gpg --encrypt --recipient admin@example.com "/backup/ha-config-$date.tar.gz"
```

### Disable Unused Features

Reduce attack surface:
```yaml
unifi_controller:
  # Only enable what you need
  device_trackers: false  # If not using presence detection
  track_guests: false     # If not monitoring guests
  services:              # Only enable required services
    - restart_device
    # - block_client  # Disable if not needed
```

## Vulnerability Reporting

### How to Report

If you discover a security vulnerability:

1. **DO NOT** create a public GitHub issue
2. **DO** email: security@example.com (replace with actual)
3. **Include**:
   - Description of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Investigation**: Within 7 days
- **Fix/Response**: Within 30 days (critical), 90 days (non-critical)
- **Disclosure**: Coordinated with reporter

### Security Updates

Critical security fixes will be:
1. Released as soon as possible
2. Announced via GitHub Security Advisories
3. Mentioned in release notes
4. Backported to supported versions

## Security Checklist

Before deploying to production:

- [ ] Created dedicated service account
- [ ] Using strong, unique password
- [ ] Password stored in secrets.yaml
- [ ] secrets.yaml in .gitignore
- [ ] secrets.yaml has 600 permissions
- [ ] Home Assistant requires authentication
- [ ] HTTPS enabled (if remote access needed)
- [ ] 2FA enabled on Home Assistant
- [ ] Controller on isolated network/VLAN
- [ ] Firewall rules restrict access
- [ ] Remote access disabled (if not needed)
- [ ] SSL verification appropriate for setup
- [ ] Regular backup strategy in place
- [ ] Monitoring for anomalies configured
- [ ] Password rotation scheduled

## Resources

### UniFi Security
- [UniFi Security Guide](https://help.ui.com/hc/en-us/articles/360012282453)
- [UniFi Network Security](https://help.ui.com/hc/en-us/articles/115004872787)

### Home Assistant Security
- [Home Assistant Security](https://www.home-assistant.io/docs/security/)
- [Securing Home Assistant](https://www.home-assistant.io/docs/configuration/securing/)

### General Security
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Controls](https://www.cisecurity.org/controls/)

---

**Remember**: Security is an ongoing process, not a one-time setup. Regular reviews and updates are essential.

**When in doubt**: Choose the more secure option. It's easier to relax security later than to recover from a breach.
