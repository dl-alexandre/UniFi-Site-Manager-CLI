# Dalton Alexandre's Homebrew Tap

This is a Homebrew tap containing formulas for UniFi CLI tools.

## Available Formulas

### `usm` - UniFi Site Manager CLI

Command-line interface for managing UniFi sites via cloud API.

**Installation:**
```bash
brew tap dl-alexandre/unifi-cli
brew install usm
```

**Features:**
- Manage UniFi sites from the command line
- Device management and monitoring
- Network configuration
- Cloud API integration

## Usage

```bash
# Show help
usm --help

# Login to your UniFi account
usm account login

# List your sites
usm site list

# Show current configuration
usm config view
```

## Updates

To update to the latest version:
```bash
brew upgrade usm
```

## Uninstallation

```bash
brew uninstall usm
brew untap dl-alexandre/unifi-cli
```

## Development

This tap is maintained by [Dalton Alexandre](https://github.com/dl-alexandre).

## License

These formulas are provided under the MIT License.
