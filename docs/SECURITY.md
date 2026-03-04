# Security Best Practices

Security guidelines for using UniFi Site Manager CLI safely.

## Table of Contents

- [Overview](#overview)
- [API Key Security](#api-key-security)
- [Authentication Best Practices](#authentication-best-practices)
- [Credential Management](#credential-management)
- [Network Security](#network-security)
- [Local Controller Security](#local-controller-security)
- [Scripting Security](#scripting-security)
- [Container Security](#container-security)
- [Audit and Monitoring](#audit-and-monitoring)
- [Reporting Vulnerabilities](#reporting-vulnerabilities)

## Overview

UniFi Site Manager CLI handles sensitive network credentials and API keys. This guide provides best practices for secure usage.

### Security Principles

1. **Never expose credentials** in logs, scripts, or version control
2. **Use least privilege** - only necessary permissions
3. **Rotate credentials** regularly
4. **Monitor and audit** all API access
5. **Encrypt at rest** and **in transit**

## API Key Security

### Generating API Keys

1. Log into [unifi.ui.com](https://unifi.ui.com)
2. Navigate to: Settings → Control Plane → Integrations → API Keys
3. Click **Create API Key**
4. Copy immediately (shown only once)

### API Key Storage

**DO:**
```bash
# Use environment variables
export USM_API_KEY="your-api-key"

# Use password managers
export USM_API_KEY="$(secret-tool lookup service usm)"

# Use dedicated secrets services (AWS Secrets Manager, HashiCorp Vault)
```

**DON'T:**
```bash
# Never use in command line
usm --api-key="your-api-key" sites list  # VISIBLE in process list!

# Never commit to git
echo "USM_API_KEY=xxx" >> config.yaml && git add config.yaml

# Never log to files
usm sites list 2>&1 | tee log.txt  # May contain debug info
```

### API Key Rotation

**Schedule**: Every 90 days minimum

```bash
#!/bin/bash
# rotate-key.sh

echo "Step 1: Generate new key at https://unifi.ui.com"
read -p "Enter new API key: " NEW_KEY

# Test new key
export USM_API_KEY="$NEW_KEY"
if usm whoami > /dev/null; then
    echo "✅ New key works!"
    
    # Update environment
    echo "export USM_API_KEY=$NEW_KEY" >> ~/.bashrc
    
    echo "Step 2: Revoke old key at https://unifi.ui.com"
    echo "Step 3: Update any CI/CD pipelines"
else
    echo "❌ New key failed!"
    exit 1
fi
```

### Multiple API Keys

Create separate keys for different environments:

| Environment | Key Purpose | Permissions |
|-------------|-------------|-------------|
| Development | Testing | Read-only |
| Staging | Integration testing | Read/Write |
| Production | Live operations | Minimal required |
| CI/CD | Automated deployment | Site-specific |
| Monitoring | Health checks | Read-only |

## Authentication Best Practices

### Environment Variables

**Secure approach**:
```bash
# ~/.bashrc or ~/.zshrc
export USM_API_KEY="$(cat ~/.usm/api-key)"
chmod 600 ~/.usm/api-key
```

**Docker approach**:
```bash
# Use Docker secrets or environment files
docker run --env-file .env usm:latest sites list
# .env is in .gitignore!
```

### Configuration Files

```yaml
# ~/.config/usm/config.yaml
api:
  base_url: https://api.ui.com
  timeout: 30

# NOTE: API keys are NOT stored here for security!
```

Permissions:
```bash
chmod 600 ~/.config/usm/config.yaml
chmod 700 ~/.config/usm
```

### Interactive Input

```bash
# For one-time operations
read -s -p "API Key: " USM_API_KEY
export USM_API_KEY
usm sites list
```

## Credential Management

### 1Password Integration

```bash
#!/bin/bash
# Get credentials from 1Password

export USM_API_KEY=$(op read "op://Private/USM/API Key")
export USM_PASSWORD=$(op read "op://Private/UniFi/password")

usm sites list
```

### HashiCorp Vault

```bash
# Read from Vault
export USM_API_KEY=$(vault kv get -field=api_key secret/usm)

usm sites list
```

### AWS Secrets Manager

```bash
# Using AWS CLI
export USM_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id usm/api-key \
  --query SecretString \
  --output text)

usm sites list
```

### macOS Keychain

```bash
# Store credential
security add-generic-password -s "usm-api-key" -a "user" -w "your-key"

# Retrieve
export USM_API_KEY=$(security find-generic-password -s "usm-api-key" -w)
```

## Network Security

### TLS/SSL

Always use HTTPS:
```bash
# Good
export USM_BASE_URL="https://api.ui.com"

# Bad - never use HTTP
export USM_BASE_URL="http://api.ui.com"  # ❌
```

### Certificate Verification

**Cloud API**: Always verify certificates (default)

**Local Controllers**: May require skipping verification for self-signed certs:
```bash
# For UniFi OS with self-signed certificates
export USM_INSECURE=true  # Only if necessary!
```

### Proxy Configuration

```bash
# If behind corporate proxy
export HTTPS_PROXY="http://proxy.company.com:8080"
export NO_PROXY="localhost,127.0.0.1,192.168.0.0/16"

usm sites list
```

### Firewall Rules

Allow outbound HTTPS only:
```
# Cloud API
ALLOW OUTBOUND TO api.ui.com:443

# Local Controller
ALLOW OUTBOUND TO 192.168.1.1:443  # UDM
```

## Local Controller Security

### Account Setup

1. Create dedicated **local admin account** (not SSO)
2. Use strong password (16+ characters)
3. Enable 2FA on UniFi account if possible
4. Restrict account permissions

### Password Security

```bash
# NEVER in command line
usm --local --password="secret123" sites list  # ❌ VISIBLE in ps aux!

# ALWAYS environment variable
export USM_PASSWORD="secret123"  # Better
usm --local sites list

# Best: secrets manager
export USM_PASSWORD="$(secret-tool lookup service unifi)"
```

### Network Segmentation

```
[Management Network: 192.168.10.0/24]
    |
    |-- Admin Workstation
    |-- Automation Server
    |
    +-- Firewall Rule: Allow HTTPS to UDM only
    |
[UDM: 192.168.1.1]
```

## Scripting Security

### Shebang and Permissions

```bash
#!/bin/bash
# usm-backup.sh
set -euo pipefail  # Strict mode

# Permissions
chmod 700 usm-backup.sh
```

### Avoid Hardcoded Credentials

**Bad**:
```bash
#!/bin/bash
API_KEY="sk-1234567890abcdef"  # ❌ Hardcoded!
usm --api-key="$API_KEY" sites list
```

**Good**:
```bash
#!/bin/bash
API_KEY="${USM_API_KEY:?API key not set}"
usm sites list
```

### Secure Temporary Files

```bash
# Create secure temp file
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

usm sites list --output json > "$TEMP_FILE"
# Process...
```

### Logging Safely

```bash
#!/bin/bash
# Redact sensitive data in logs

LOG_FILE="usm-$(date +%Y%m%d).log"

# Run with debug but filter output
usm --debug sites list 2>&1 | \
  sed -E 's/(api-key|password|token|cookie)=[^[:space:]]+/\1=[REDACTED]/gi' | \
  tee "$LOG_FILE"
```

### Cron Jobs

```bash
# Secure cron entry
# /etc/cron.d/usm-monitor
USM_API_KEY=/etc/usm/api-key
*/5 * * * * root /usr/local/bin/usm sites health SITE_ID > /dev/null 2>&1 || echo "USM check failed" | mail -s "USM Alert" admin@example.com
```

Store API key:
```bash
# Only root can read
sudo chmod 600 /etc/usm/api-key
sudo chown root:root /etc/usm/api-key
```

## Container Security

### Docker Best Practices

```dockerfile
# Dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY usm /usr/local/bin/usm
ENTRYPOINT ["usm"]
```

**Run securely**:
```bash
# Don't pass credentials as build args
docker build -t usm:local .

# Use environment at runtime
docker run --rm \
  -e USM_API_KEY="$(cat ~/.usm/api-key)" \
  -v ~/.config/usm:/root/.config/usm:ro \
  usm:latest sites list
```

### Kubernetes Secrets

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: usm-credentials
type: Opaque
stringData:
  api-key: "your-api-key"
---
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: usm-monitor
spec:
  template:
    spec:
      containers:
      - name: usm
        image: usm:latest
        env:
        - name: USM_API_KEY
          valueFrom:
            secretKeyRef:
              name: usm-credentials
              key: api-key
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  usm:
    image: usm:latest
    environment:
      - USM_API_KEY=${USM_API_KEY}  # From .env file
    env_file:
      - .env  # Not committed to git!
    volumes:
      - ./config:/root/.config/usm:ro
```

```bash
# .env (in .gitignore!)
USM_API_KEY=your-api-key
```

## Audit and Monitoring

### Access Logging

Enable audit logging:
```bash
# Log all API calls
export USM_DEBUG=true
usm sites list 2>&1 | tee -a /var/log/usm-audit.log
```

### Monitoring API Usage

```bash
#!/bin/bash
# monitor-usage.sh

LOG_FILE="/var/log/usm-usage.log"

# Log with timestamp and user
echo "$(date '+%Y-%m-%d %H:%M:%S') $(whoami) $1" >> "$LOG_FILE"

usm "$@"
```

### Rate Limit Monitoring

```bash
# Check if approaching limits
usm sites list --debug 2>&1 | grep -i "rate\|limit\|retry"
```

### Failed Authentication Alerts

```bash
#!/bin/bash
# auth-monitor.sh

if ! usm whoami > /dev/null 2>&1; then
  echo "USM authentication failed at $(date)" | \
    mail -s "Security Alert: USM Auth Failure" security@example.com
fi
```

## Reporting Vulnerabilities

If you discover a security vulnerability:

1. **DO NOT** open a public issue
2. Email: security@example.com (replace with actual address)
3. Include:
   - Description of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will:
- Acknowledge receipt within 48 hours
- Investigate and respond within 7 days
- Coordinate disclosure timeline
- Credit you (if desired)

## Security Checklist

Before using in production:

- [ ] API keys stored securely (not in code)
- [ ] Environment variables used for credentials
- [ ] Config file permissions set to 600
- [ ] HTTPS only (no HTTP)
- [ ] Certificate verification enabled (Cloud API)
- [ ] Rate limiting understood
- [ ] Audit logging configured
- [ ] Credential rotation scheduled
- [ ] Dedicated service accounts created
- [ ] Principle of least privilege applied
- [ ] Debug mode disabled in production
- [ ] Secrets in .gitignore
- [ ] Container images scanned
- [ ] Network segmentation implemented

## Resources

- [OWASP Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [Go Security Guidelines](https://golang.org/security)
- [Docker Security](https://docs.docker.com/engine/security/)
- [UniFi Security Best Practices](https://help.ui.com/hc/en-us/articles/360012282453)

---

**Remember**: Security is everyone's responsibility. When in doubt, choose the more secure option.
