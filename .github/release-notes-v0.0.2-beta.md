## v0.0.2-beta - Local Controller Support

🚀 **Beta Release**: This version introduces experimental support for direct connection to UniFi OS local controllers (UDM, UDM-Pro, UDR).

### 🆕 What's New

#### Local Controller Mode (Beta)

Connect directly to your UniFi hardware without using the cloud API:

```bash
# Basic usage
usm --local --host=192.168.1.1 --username=admin --password=xxx devices list

# With environment variables (recommended for security)
export USM_LOCAL=true
export USM_HOST=192.168.1.1
export USM_USERNAME=admin
export USM_PASSWORD=yourpassword

usm devices list
usm clients list
usm wlans list
```

#### Debug Mode

Verbose API logging with **automatic credential redaction**:

```bash
usm --local --host=192.168.1.1 --debug devices list
```

Debug output format:
```
[DEBUG] === REQUEST ===
[DEBUG] Method: GET URL: https://192.168.1.1/proxy/network/api/s/default/stat/device
[DEBUG] Header: X-CSRF-Token: [REDACTED]
[DEBUG] Body: {"name":"Test","essid":"TestSSID"}
[DEBUG] =================
[DEBUG] === RESPONSE ===
[DEBUG] Status: 200 OK
[DEBUG] Raw Payload: {"data":[{"_id":"abc123"...
```

**Security**: Passwords, CSRF tokens, cookies, and API keys are automatically redacted. Safe to share in bug reports!

### ✅ Working Features

**Local Controller (Beta)**:
- ✅ Site listing (`sites list`)
- ✅ Device listing and details (`devices list`, `devices get`)
- ✅ Client listing (`clients list`)
- ✅ WLAN CRUD operations
- ✅ Device restart

**Cloud API**:
- ✅ All previous features maintained
- ✅ Enhanced retry logic
- ✅ Debug logging support

### 🚧 Known Limitations

27 methods are stubbed and return "not yet implemented":
- Site creation/update/delete
- Host/Console management  
- Device adoption and firmware upgrades
- Client blocking/unblocking
- Alerts and events
- Network configuration

### 🔧 Chrome Dev Tools Debugging

If local controller commands fail:

1. Open UniFi web UI in Chrome
2. Press F12 → Network tab
3. Perform action in web UI (e.g., create WLAN)
4. Find the `POST` request to `wlanconf`
5. Check **Payload** tab for exact JSON structure
6. Run CLI with `--debug` and compare

### 🐛 Reporting Issues

Please test and report issues!

1. Run with `--debug`: `usm --local --debug devices list`
2. Copy the `[DEBUG]` output (already redacted)
3. Open a GitHub issue with:
   - Command you ran
   - Expected vs actual behavior
   - Debug output
   - UniFi OS version (if known)

### 📦 Installation

Download from assets below or install via Homebrew (when tap is configured):

```bash
# macOS/Linux
brew tap dl-alexandre/usm
brew install usm

# Or download binary
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/download/v0.0.2/usm-$(uname -s)_$(uname -m).tar.gz | tar xz
```

### 🏗️ Architecture

```
CLI Commands (sites, devices, clients, wlans)
         │
         ▼
┌─────────────────┐
│  SiteManager    │  ← Interface abstraction
│  Interface      │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
Cloud      Local
Client     Client
(API Key)  (Cookie + CSRF)
```

### 🔒 Security Notes

- **Never use `--password` flag in production** - use `USM_PASSWORD` env var
- Debug logs automatically redact all credentials
- Local mode uses TLS with `InsecureSkipVerify` for self-signed certs
- Session cookies and CSRF tokens are managed automatically

### 📝 Changes

See [CHANGELOG.md](https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/blob/main/CHANGELOG.md) for full details.

---

**Full Changelog**: https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/compare/v0.0.1...v0.0.2
