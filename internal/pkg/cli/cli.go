// Package cli provides command-line interface using Kong
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/config"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// CLI is the main command-line interface structure using Kong
type CLI struct {
	Globals

	Init    InitCmd    `cmd:"" help:"Interactive configuration setup"`
	Sites   SitesCmd   `cmd:"" help:"Manage sites"`
	Whoami  WhoamiCmd  `cmd:"" help:"Show authenticated user information"`
	Version VersionCmd `cmd:"" help:"Show version information"`
}

// Globals contains global flags available to all commands
type Globals struct {
	APIKey     string `help:"API key for authentication" env:"USM_API_KEY"`
	BaseURL    string `help:"API base URL" env:"USM_BASE_URL"`
	Timeout    int    `help:"Request timeout in seconds" default:"30" env:"USM_TIMEOUT"`
	Format     string `help:"Output format: table, json" default:"table" enum:"table,json" env:"USM_FORMAT"`
	Color      string `help:"Color mode: auto, always, never" default:"auto" enum:"auto,always,never" env:"USM_COLOR"`
	NoHeaders  bool   `help:"Disable table headers" env:"USM_NO_HEADERS"`
	Verbose    bool   `help:"Enable verbose output" short:"v"`
	Debug      bool   `help:"Enable debug output"`
	ConfigFile string `help:"Config file path" short:"c" env:"USM_CONFIG"`

	appConfig *config.Config
	appClient *api.Client
}

func (g *Globals) AfterApply() error {
	// Skip initialization for init command (which doesn't need API client)
	return nil
}

func (g *Globals) initClient() error {
	// Load configuration
	flags := config.GlobalFlags{
		APIKey:     g.APIKey,
		BaseURL:    g.BaseURL,
		Timeout:    g.Timeout,
		Format:     g.Format,
		Color:      g.Color,
		NoHeaders:  g.NoHeaders,
		Verbose:    g.Verbose,
		Debug:      g.Debug,
		ConfigFile: g.ConfigFile,
	}

	cfg, err := config.Load(flags)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	g.appConfig = cfg

	// Get API key (from flag or env)
	apiKey, err := config.GetAPIKey(g.APIKey)
	if err != nil {
		return err
	}

	// Initialize API client
	client, err := api.NewClient(api.ClientOptions{
		BaseURL: cfg.API.BaseURL,
		APIKey:  apiKey,
		Timeout: cfg.API.Timeout,
		Verbose: g.Verbose,
		Debug:   g.Debug,
	})
	if err != nil {
		return err
	}
	g.appClient = client

	return nil
}

func (g *Globals) getFormatter() *output.Formatter {
	return output.NewFormatter(g.appConfig.Output.Format, g.appConfig.Output.Color, g.appConfig.Output.NoHeaders)
}

// InitCmd handles the init command
type InitCmd struct {
	Force bool `help:"Overwrite existing config"`
}

func (c *InitCmd) Run(g *Globals) error {
	// Check if config already exists
	if config.ConfigExists() && !c.Force {
		return fmt.Errorf("config already exists. Use --force to overwrite")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("UniFi Site Manager CLI - Configuration Setup")
	fmt.Println("==========================================")

	// Base URL
	fmt.Print("Base URL [https://api.ui.com]: ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://api.ui.com"
	}

	// Output format
	fmt.Print("Default output format [table]: ")
	format, _ := reader.ReadString('\n')
	format = strings.TrimSpace(format)
	if format == "" {
		format = "table"
	}
	if err := output.ValidateFormat(format); err != nil {
		return err
	}

	// Color mode
	fmt.Print("Color mode [auto]: ")
	color, _ := reader.ReadString('\n')
	color = strings.TrimSpace(color)
	if color == "" {
		color = "auto"
	}

	// Create config
	cfg := &config.Config{
		API: config.APIConfig{
			BaseURL: baseURL,
			Timeout: 30,
		},
		Output: config.OutputConfig{
			Format:    format,
			Color:     color,
			NoHeaders: false,
		},
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	output.PrintInitSuccess(config.GetConfigFilePath())
	return nil
}

// SitesCmd groups site-related commands
type SitesCmd struct {
	List ListSitesCmd `cmd:"" help:"List all sites"`
	Get  GetSiteCmd   `cmd:"" help:"Get a specific site"`
}

// ListSitesCmd handles the sites list command
type ListSitesCmd struct {
	PageSize int    `help:"Number of sites per page (0 = fetch all)" default:"50"`
	Search   string `help:"Filter sites by name/description"`
}

func (c *ListSitesCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	var allSites []api.Site
	nextToken := ""

	// Fetch all sites (handle pagination)
	for {
		resp, err := g.appClient.ListSites(c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allSites = append(allSites, resp.Data...)

		// Check if we should fetch all pages
		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			// Only fetch one page
			break
		}
	}

	// Filter by search term if provided
	if c.Search != "" {
		filtered := make([]api.Site, 0)
		searchLower := strings.ToLower(c.Search)
		for _, site := range allSites {
			if strings.Contains(strings.ToLower(site.Name), searchLower) ||
				strings.Contains(strings.ToLower(site.Description), searchLower) {
				filtered = append(filtered, site)
			}
		}
		allSites = filtered
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allSites)
	}

	// Convert to output format
	siteData := make([]output.SiteData, len(allSites))
	for i, site := range allSites {
		siteData[i] = output.SiteData{
			ID:          site.ID,
			Name:        site.Name,
			Description: site.Description,
			HostID:      site.HostID,
		}
	}

	formatter.PrintSitesTable(siteData)
	return nil
}

// GetSiteCmd handles the sites get command
type GetSiteCmd struct {
	SiteID string `arg:"" help:"Site ID to retrieve"`
}

func (c *GetSiteCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	resp, err := g.appClient.GetSite(c.SiteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	site := resp.Data
	siteData := []output.SiteData{{
		ID:          site.ID,
		Name:        site.Name,
		Description: site.Description,
		HostID:      site.HostID,
	}}

	formatter.PrintSitesTable(siteData)
	return nil
}

// WhoamiCmd handles the whoami command
type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.Whoami()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	user := resp.Data
	userData := output.UserData{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		IsOwner:   user.IsOwner,
	}

	formatter.PrintUserTable(userData)
	return nil
}

// VersionCmd handles the version command
type VersionCmd struct {
	Check bool `help:"Check for updates"`
}

func (c *VersionCmd) Run(g *Globals) error {
	// These will be set at build time
	version := "dev"
	gitCommit := "unknown"
	buildTime := "unknown"

	// Try to get from main package vars (these are set via ldflags)
	// For now, just print placeholder
	output.PrintVersion(version, gitCommit, buildTime, c.Check)
	return nil
}

// Run parses CLI args and executes the appropriate command
func Run(args []string, version, gitCommit, buildTime string) (int, error) {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("usm"),
		kong.Description("UniFi Site Manager CLI"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	if err != nil {
		return api.ExitGeneralError, err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return api.ExitValidationError, err
	}

	// Update version info for version command
	if ctx.Command() == "version" {
		// Would update global vars here
	}

	err = ctx.Run(&cli.Globals)
	if err != nil {
		return api.GetExitCode(err), err
	}

	return api.ExitSuccess, nil
}
