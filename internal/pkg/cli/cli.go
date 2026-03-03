// Package cli provides command-line interface using Kong
package cli

import (
	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/config"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// CLIContext holds the dependencies passed to commands
// This decouples commands from the concrete client implementation,
// allowing the same commands to work with Cloud or Local controllers
type CLIContext struct {
	Client  api.SiteManager // Interface, not concrete type
	Config  *config.Config
	Format  string
	Color   string
	Verbose bool
	Debug   bool
}

// RootCmd is the root command structure
// Domain subcommands are defined in their respective files (sites.go, devices.go, etc.)
type RootCmd struct {
	// Output formatting (applies to both modes)
	Format    string `help:"Output format: table, json" default:"table" enum:"table,json" env:"USM_FORMAT"`
	Color     string `help:"Color mode: auto, always, never" default:"auto" enum:"auto,always,never" env:"USM_COLOR"`
	NoHeaders bool   `help:"Disable table headers" env:"USM_NO_HEADERS"`
	Verbose   bool   `help:"Enable verbose output" short:"v"`
	Debug     bool   `help:"Enable debug output"`

	// --- Cloud Mode Authentication (Default) ---
	// Uses UniFi Site Manager Cloud API (api.ui.com)
	APIKey     string `name:"api-key" env:"USM_API_KEY" help:"Cloud Site Manager API Key" group:"Cloud Mode"`
	BaseURL    string `name:"base-url" env:"USM_BASE_URL" help:"Cloud API base URL (default: https://api.ui.com)" group:"Cloud Mode"`
	Timeout    int    `env:"USM_TIMEOUT" help:"Request timeout in seconds" default:"30" group:"Cloud Mode"`
	ConfigFile string `short:"c" env:"USM_CONFIG" help:"Config file path" group:"Cloud Mode"`

	// --- Local Controller Mode ---
	// Direct connection to UniFi OS controllers (UDM, UDM-Pro, UDR, etc.)
	Local    bool   `name:"local" env:"USM_LOCAL" help:"Enable Local Controller mode" group:"Local Mode"`
	Host     string `name:"host" env:"USM_HOST" help:"Local Controller IP or hostname (e.g., 192.168.1.1 or unifi.local)" group:"Local Mode"`
	Username string `name:"username" env:"USM_USERNAME" help:"Local Controller username" group:"Local Mode"`
	Password string `name:"password" env:"USM_PASSWORD" help:"Local Controller password (use env var for security)" group:"Local Mode"`

	// --- Commands ---
	Init     InitCmd     `cmd:"" help:"Interactive configuration setup"`
	Sites    SitesCmd    `cmd:"" help:"Manage sites"`
	Hosts    HostsCmd    `cmd:"" help:"Manage hosts/consoles"`
	Devices  DevicesCmd  `cmd:"" help:"Manage devices"`
	Clients  ClientsCmd  `cmd:"" help:"Manage clients"`
	WLANs    WLANsCmd    `cmd:"" help:"Manage wireless networks"`
	Alerts   AlertsCmd   `cmd:"" help:"Manage alerts"`
	Events   EventsCmd   `cmd:"" help:"View events"`
	Networks NetworksCmd `cmd:"" help:"Manage networks"`
	Whoami   WhoamiCmd   `cmd:"" help:"Show authenticated user information"`
	Version  VersionCmd  `cmd:"" help:"Show version information"`
}

// IsLocalMode returns true if local controller mode is enabled
func (r *RootCmd) IsLocalMode() bool {
	return r.Local
}

// ValidateLocalAuth checks if all required local auth parameters are present
func (r *RootCmd) ValidateLocalAuth() error {
	if r.Host == "" {
		return &api.ValidationError{Message: "--host is required for local mode (or set USM_HOST)"}
	}
	if r.Username == "" {
		return &api.ValidationError{Message: "--username is required for local mode (or set USM_USERNAME)"}
	}
	if r.Password == "" {
		return &api.ValidationError{Message: "--password is required for local mode (or set USM_PASSWORD env var)"}
	}
	return nil
}

// ValidateCloudAuth checks if cloud auth parameters are present
func (r *RootCmd) ValidateCloudAuth() error {
	if r.APIKey == "" {
		return &api.ValidationError{Message: "--api-key is required for cloud mode (or set USM_API_KEY env var)"}
	}
	return nil
}

// createCloudContext initializes CLI context for Cloud mode
func (r *RootCmd) createCloudContext() (*CLIContext, error) {
	flags := config.GlobalFlags{
		APIKey:     r.APIKey,
		BaseURL:    r.BaseURL,
		Timeout:    r.Timeout,
		Format:     r.Format,
		Color:      r.Color,
		NoHeaders:  r.NoHeaders,
		Verbose:    r.Verbose,
		Debug:      r.Debug,
		ConfigFile: r.ConfigFile,
	}

	cfg, err := config.Load(flags)
	if err != nil {
		return nil, err
	}

	apiKey, err := config.GetAPIKey(r.APIKey)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClient(api.ClientOptions{
		BaseURL: cfg.API.BaseURL,
		APIKey:  apiKey,
		Timeout: cfg.API.Timeout,
		Verbose: r.Verbose,
		Debug:   r.Debug,
	})
	if err != nil {
		return nil, err
	}

	return &CLIContext{
		Client:  client,
		Config:  cfg,
		Format:  cfg.Output.Format,
		Color:   cfg.Output.Color,
		Verbose: r.Verbose,
		Debug:   r.Debug,
	}, nil
}

// createLocalContext initializes CLI context for Local mode
func (r *RootCmd) createLocalContext() (*CLIContext, error) {
	// Local mode doesn't use the config file system
	// It directly uses the provided credentials
	client, err := api.NewLocalClient(api.LocalClientOptions{
		Host:          r.Host,
		Username:      r.Username,
		Password:      r.Password,
		AllowInsecure: true, // Local controllers use self-signed certs
	})
	if err != nil {
		return nil, err
	}

	// Create a minimal config for output formatting
	cfg := &config.Config{
		Output: config.OutputConfig{
			Format:    r.Format,
			Color:     r.Color,
			NoHeaders: r.NoHeaders,
		},
	}

	return &CLIContext{
		Client:  client,
		Config:  cfg,
		Format:  r.Format,
		Color:   r.Color,
		Verbose: r.Verbose,
		Debug:   r.Debug,
	}, nil
}

// getFormatter creates a formatter from context
func (ctx *CLIContext) getFormatter() *output.Formatter {
	return output.NewFormatter(ctx.Format, ctx.Color, ctx.Config.Output.NoHeaders)
}

// Run parses CLI args and executes the appropriate command
func Run(args []string, version, gitCommit, buildTime string) (int, error) {
	var root RootCmd
	parser, err := kong.New(&root,
		kong.Name("usm"),
		kong.Description("UniFi Site Manager CLI - Full-featured management for UniFi networks\n\nMode Selection:\n  Cloud Mode (default): Connects to UniFi Site Manager Cloud API\n  Local Mode (--local): Connects directly to UniFi OS controllers"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Tree:    true,
		}),
	)
	if err != nil {
		return api.ExitGeneralError, err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return api.ExitValidationError, err
	}

	// Validate mode selection and authentication
	var cliCtx *CLIContext
	if root.IsLocalMode() {
		// Local Controller Mode
		if err := root.ValidateLocalAuth(); err != nil {
			return api.ExitValidationError, err
		}
		cliCtx, err = root.createLocalContext()
		if err != nil {
			return api.GetExitCode(err), err
		}
	} else {
		// Cloud Mode (default)
		if err := root.ValidateCloudAuth(); err != nil {
			return api.ExitValidationError, err
		}
		cliCtx, err = root.createCloudContext()
		if err != nil {
			return api.GetExitCode(err), err
		}
	}

	// Enable debug logging if requested (with credential redaction)
	if cliCtx.Debug {
		cliCtx.Client.EnableDebug()
	}

	err = ctx.Run(cliCtx)
	if err != nil {
		return api.GetExitCode(err), err
	}

	return api.ExitSuccess, nil
}
