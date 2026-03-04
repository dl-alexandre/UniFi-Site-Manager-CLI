# Fish completion for usm (UniFi Site Manager CLI)

# Main commands
complete -c usm -f -n __fish_use_subcommand -a "account" -d "Manage UniFi account"
complete -c usm -f -n __fish_use_subcommand -a "site" -d "Manage UniFi sites"
complete -c usm -f -n __fish_use_subcommand -a "device" -d "Manage UniFi devices"
complete -c usm -f -n __fish_use_subcommand -a "network" -d "Manage UniFi networks"
complete -c usm -f -n __fish_use_subcommand -a "config" -d "Manage configuration"
complete -c usm -f -n __fish_use_subcommand -a "help" -d "Show help"

# Account subcommands
complete -c usm -f -n '__fish_seen_subcommand_from account' -a "login" -d "Login to UniFi account"
complete -c usm -f -n '__fish_seen_subcommand_from account' -a "logout" -d "Logout from UniFi account"
complete -c usm -f -n '__fish_seen_subcommand_from account' -a "status" -d "Show account status"

# Site subcommands
complete -c usm -f -n '__fish_seen_subcommand_from site' -a "list" -d "List all sites"
complete -c usm -f -n '__fish_seen_subcommand_from site' -a "create" -d "Create new site"
complete -c usm -f -n '__fish_seen_subcommand_from site' -a "delete" -d "Delete a site"
complete -c usm -f -n '__fish_seen_subcommand_from site' -a "select" -d "Select current site"
complete -c usm -f -n '__fish_seen_subcommand_from site' -a "current" -d "Show current site"

# Device subcommands
complete -c usm -f -n '__fish_seen_subcommand_from device' -a "list" -d "List all devices"
complete -c usm -f -n '__fish_seen_subcommand_from device' -a "adopt" -d "Adopt a new device"
complete -c usm -f -n '__fish_seen_subcommand_from device' -a "restart" -d "Restart a device"
complete -c usm -f -n '__fish_seen_subcommand_from device' -a "upgrade" -d "Upgrade device firmware"

# Network subcommands
complete -c usm -f -n '__fish_seen_subcommand_from network' -a "status" -d "Show network status"
complete -c usm -f -n '__fish_seen_subcommand_from network' -a "clients" -d "List connected clients"

# Config subcommands
complete -c usm -f -n '__fish_seen_subcommand_from config' -a "init" -d "Initialize configuration"
complete -c usm -f -n '__fish_seen_subcommand_from config' -a "view" -d "View configuration"
complete -c usm -f -n '__fish_seen_subcommand_from config' -a "set" -d "Set configuration value"

# Options
complete -c usm -f -s v -l version -d "Show version"
complete -c usm -f -s h -l help -d "Show help"
