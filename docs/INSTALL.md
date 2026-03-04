# Installation Guide

Complete installation instructions for UniFi Site Manager CLI on all supported platforms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [macOS Installation](#macos-installation)
- [Linux Installation](#linux-installation)
- [Windows Installation](#windows-installation)
- [Docker Installation](#docker-installation)
- [Build from Source](#build-from-source)
- [Package Managers](#package-managers)
- [Post-Installation](#post-installation)
- [Uninstallation](#uninstallation)

## Prerequisites

### System Requirements

- **Operating System**: macOS 10.15+, Linux (x86_64, ARM64), Windows 10+
- **Memory**: 50 MB RAM (minimal)
- **Disk Space**: 20 MB
- **Network**: Internet connection for cloud API mode

### For Building from Source

- **Go**: Version 1.24 or later
- **Git**: For cloning the repository
- **Make**: For using the Makefile

## macOS Installation

### Option 1: Homebrew (Recommended)

```bash
# Add the tap
brew tap dl-alexandre/tap

# Install usm
brew install usm

# Verify installation
usm version
```

### Option 2: Direct Download

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-darwin-arm64 -o usm

# Intel Macs
curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-darwin-amd64 -o usm

# Make executable and move to PATH
chmod +x usm
sudo mv usm /usr/local/bin/

# Verify
usm version
```

### Option 3: MacPorts

```bash
# Coming soon
sudo port install usm
```

## Linux Installation

### Option 1: Download Binary

```bash
# Detect architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64) BINARY="usm-linux-amd64" ;;
  aarch64) BINARY="usm-linux-arm64" ;;
  armv7l) BINARY="usm-linux-armv7" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Download
curl -L "https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/${BINARY}" -o usm

# Install
chmod +x usm
sudo mv usm /usr/local/bin/

# Verify
usm version
```

### Option 2: Package Managers

#### Debian/Ubuntu (apt)

```bash
# Download .deb package
wget https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm_amd64.deb

# Install
sudo dpkg -i usm_amd64.deb

# Fix dependencies if needed
sudo apt-get install -f
```

#### RHEL/CentOS/Fedora (rpm)

```bash
# Download .rpm package
wget https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm_x86_64.rpm

# Install
sudo rpm -i usm_x86_64.rpm
```

#### Arch Linux (AUR)

```bash
# Using yay
yay -S usm

# Or using paru
paru -S usm
```

### Option 3: Snap

```bash
# Coming soon
sudo snap install usm
```

## Windows Installation

### Option 1: Scoop (Recommended)

```powershell
# Add bucket
scoop bucket add unifi https://github.com/dl-alexandre/scoop-bucket

# Install
scoop install usm

# Verify
usm version
```

### Option 2: Chocolatey

```powershell
# Coming soon
choco install usm
```

### Option 3: Manual Download

```powershell
# Download latest release
# https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-windows-amd64.exe

# Move to a directory in PATH
# Example: C:\Program Files\usm\

# Add to PATH environment variable
# System Properties → Environment Variables → Path → Edit → New
```

### Option 4: Winget

```powershell
# Coming soon
winget install dl-alexandre.usm
```

## Docker Installation

### Using Pre-built Image

```bash
# Pull the image
docker pull ghcr.io/dl-alexandre/usm:latest

# Run with API key
docker run --rm -e USM_API_KEY="your-key" ghcr.io/dl-alexandre/usm sites list

# With config volume
docker run --rm \
  -v ~/.config/usm:/root/.config/usm \
  ghcr.io/dl-alexandre/usm sites list
```

### Building Docker Image

```bash
# Clone repository
git clone https://github.com/dl-alexandre/UniFi-Site-Manager-CLI.git
cd UniFi-Site-Manager-CLI

# Build image
docker build -t usm:local .

# Run
docker run --rm usm:local version
```

### Docker Compose

```yaml
version: '3.8'
services:
  usm:
    image: ghcr.io/dl-alexandre/usm:latest
    environment:
      - USM_API_KEY=${USM_API_KEY}
    volumes:
      - ./config:/root/.config/usm
    command: sites list
```

## Build from Source

### Prerequisites

- Go 1.24 or later
- Make (optional)
- Git

### Steps

```bash
# Clone repository
git clone https://github.com/dl-alexandre/UniFi-Site-Manager-CLI.git
cd UniFi-Site-Manager-CLI

# Install dependencies
make deps
# Or: go mod download

# Build binary
make build
# Or: go build -o usm ./cmd/usm

# Install to system
sudo cp usm /usr/local/bin/

# Verify
usm version
```

### Cross-Compilation

```bash
# Build for multiple platforms
make release

# Or manually:
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o usm-linux-amd64 ./cmd/usm

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o usm-darwin-amd64 ./cmd/usm

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o usm-darwin-arm64 ./cmd/usm

# Windows
GOOS=windows GOARCH=amd64 go build -o usm-windows-amd64.exe ./cmd/usm
```

## Package Managers

### Homebrew (macOS/Linux)

```bash
# Install
brew install dl-alexandre/tap/usm

# Update
brew upgrade usm

# Uninstall
brew uninstall usm
```

### Go Install

```bash
# Install directly
go install github.com/dl-alexandre/UniFi-Site-Manager-CLI/cmd/usm@latest

# Binary will be in $GOPATH/bin or $HOME/go/bin
# Add to PATH if needed
```

## Post-Installation

### 1. Verify Installation

```bash
# Check version
usm version

# Expected output:
# Version: v1.0.0
# Git Commit: abc123
# Build Time: 2024-01-01T00:00:00Z
# Go Version: go1.24.0
```

### 2. Initialize Configuration

```bash
# Interactive setup
usm init

# This creates ~/.config/usm/config.yaml
```

### 3. Test Connection

```bash
# Cloud API mode
usm whoami

# Local controller mode
usm --local --host=192.168.1.1 --username=admin --password=secret whoami
```

### 4. Shell Completion (Optional)

```bash
# Bash
echo 'eval "$(usm completion bash)"' >> ~/.bashrc

# Zsh
echo 'eval "$(usm completion zsh)"' >> ~/.zshrc

# Fish
usm completion fish > ~/.config/fish/completions/usm.fish
```

## Uninstallation

### macOS

```bash
# Homebrew
brew uninstall usm
brew untap dl-alexandre/tap

# Manual
sudo rm /usr/local/bin/usm
rm -rf ~/.config/usm
```

### Linux

```bash
# Binary installation
sudo rm /usr/local/bin/usm

# Package manager
# Debian/Ubuntu
sudo apt-get remove usm

# RHEL/CentOS/Fedora
sudo rpm -e usm

# Arch
yay -R usm

# Remove config
rm -rf ~/.config/usm
```

### Windows

```powershell
# Scoop
scoop uninstall usm

# Manual
# Delete C:\Program Files\usm\usm.exe
# Remove from PATH
```

### Docker

```bash
# Remove image
docker rmi ghcr.io/dl-alexandre/usm:latest
```

## Troubleshooting Installation

### Permission Denied

```bash
# Fix permissions
chmod +x usm
sudo mv usm /usr/local/bin/

# Or install to user directory
mkdir -p ~/.local/bin
mv usm ~/.local/bin/
# Add ~/.local/bin to PATH
```

### Command Not Found

```bash
# Check if in PATH
which usm

# Add to PATH if needed
export PATH=$PATH:/usr/local/bin

# Or for Go install
export PATH=$PATH:$HOME/go/bin
```

### macOS Gatekeeper

```bash
# If macOS blocks the binary
xattr -d com.apple.quarantine /usr/local/bin/usm

# Or allow manually in System Preferences → Security & Privacy
```

## Next Steps

After installation, proceed to:
- [Usage Guide](../docs/USAGE.md) - Learn how to use the CLI
- [Configuration](../docs/CONFIGURATION.md) - Configure authentication and settings
- [FAQ](../docs/FAQ.md) - Common questions and answers
