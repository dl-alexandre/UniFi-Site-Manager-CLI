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
	return nil
}

func (g *Globals) initClient() error {
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

	apiKey, err := config.GetAPIKey(g.APIKey)
	if err != nil {
		return err
	}

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

// ========== INIT COMMAND ==========

type InitCmd struct {
	Force bool `help:"Overwrite existing config"`
}

func (c *InitCmd) Run(g *Globals) error {
	if config.ConfigExists() && !c.Force {
		return fmt.Errorf("config already exists. Use --force to overwrite")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("UniFi Site Manager CLI - Configuration Setup")
	fmt.Println("==========================================")

	fmt.Print("Base URL [https://api.ui.com]: ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://api.ui.com"
	}

	fmt.Print("Default output format [table]: ")
	format, _ := reader.ReadString('\n')
	format = strings.TrimSpace(format)
	if format == "" {
		format = "table"
	}
	if err := output.ValidateFormat(format); err != nil {
		return err
	}

	fmt.Print("Color mode [auto]: ")
	color, _ := reader.ReadString('\n')
	color = strings.TrimSpace(color)
	if color == "" {
		color = "auto"
	}

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

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	output.PrintInitSuccess(config.GetConfigFilePath())
	return nil
}

// ========== SITES COMMANDS ==========

type SitesCmd struct {
	List   ListSitesCmd  `cmd:"" help:"List all sites"`
	Get    GetSiteCmd    `cmd:"" help:"Get a specific site"`
	Create CreateSiteCmd `cmd:"" help:"Create a new site"`
	Update UpdateSiteCmd `cmd:"" help:"Update a site"`
	Delete DeleteSiteCmd `cmd:"" help:"Delete a site"`
	Health SiteHealthCmd `cmd:"" help:"Get site health"`
	Stats  SiteStatsCmd  `cmd:"" help:"Get site statistics"`
}

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

	for {
		resp, err := g.appClient.ListSites(c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allSites = append(allSites, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

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

	siteData := make([]output.SiteData, len(allSites))
	for i, site := range allSites {
		siteData[i] = output.SiteData{
			ID:          site.ID,
			Name:        site.Name,
			Description: site.Description,
			HostID:      site.HostID,
			Status:      site.Status,
		}
	}

	formatter.PrintSitesTable(siteData)
	return nil
}

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
		Status:      site.Status,
	}}

	formatter.PrintSitesTable(siteData)
	return nil
}

type CreateSiteCmd struct {
	Name        string `arg:"" help:"Site name"`
	Description string `help:"Site description"`
}

func (c *CreateSiteCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.Name == "" {
		return &api.ValidationError{Message: "site name is required"}
	}

	req := api.CreateSiteRequest{
		Name:        c.Name,
		Description: c.Description,
	}

	resp, err := g.appClient.CreateSite(req)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	fmt.Printf("Site created successfully:\n")
	siteData := []output.SiteData{{
		ID:          resp.Data.ID,
		Name:        resp.Data.Name,
		Description: resp.Data.Description,
		HostID:      resp.Data.HostID,
		Status:      resp.Data.Status,
	}}
	formatter.PrintSitesTable(siteData)
	return nil
}

type UpdateSiteCmd struct {
	SiteID      string `arg:"" help:"Site ID to update"`
	Name        string `help:"New site name"`
	Description string `help:"New site description"`
}

func (c *UpdateSiteCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	req := api.UpdateSiteRequest{}
	if c.Name != "" {
		req.Name = c.Name
	}
	if c.Description != "" {
		req.Description = c.Description
	}

	resp, err := g.appClient.UpdateSite(c.SiteID, req)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	fmt.Printf("Site updated successfully:\n")
	siteData := []output.SiteData{{
		ID:          resp.Data.ID,
		Name:        resp.Data.Name,
		Description: resp.Data.Description,
		HostID:      resp.Data.HostID,
		Status:      resp.Data.Status,
	}}
	formatter.PrintSitesTable(siteData)
	return nil
}

type DeleteSiteCmd struct {
	SiteID string `arg:"" help:"Site ID to delete"`
	Force  bool   `help:"Skip confirmation"`
}

func (c *DeleteSiteCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to delete site %s? (y/N): ", c.SiteID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	if err := g.appClient.DeleteSite(c.SiteID); err != nil {
		return err
	}

	fmt.Printf("Site %s deleted successfully\n", c.SiteID)
	return nil
}

type SiteHealthCmd struct {
	SiteID string `arg:"" help:"Site ID"`
}

func (c *SiteHealthCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	resp, err := g.appClient.GetSiteHealth(c.SiteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	formatter.PrintHealthTable(resp.Data)
	return nil
}

type SiteStatsCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	Period string `help:"Stats period (day, week, month)" default:"day"`
}

func (c *SiteStatsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	resp, err := g.appClient.GetSiteStats(c.SiteID, c.Period)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	formatter.PrintPerformanceTable(&resp.Data)
	return nil
}

// ========== HOSTS COMMANDS ==========

type HostsCmd struct {
	List    ListHostsCmd   `cmd:"" help:"List all hosts"`
	Get     GetHostCmd     `cmd:"" help:"Get a specific host"`
	Health  HostHealthCmd  `cmd:"" help:"Get host health"`
	Stats   HostStatsCmd   `cmd:"" help:"Get host statistics"`
	Restart RestartHostCmd `cmd:"" help:"Restart a host"`
}

type ListHostsCmd struct {
	PageSize int    `help:"Number of hosts per page (0 = fetch all)" default:"50"`
	Search   string `help:"Filter hosts by name"`
}

func (c *ListHostsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	var allHosts []api.Host
	nextToken := ""

	for {
		resp, err := g.appClient.ListHosts(c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allHosts = append(allHosts, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	if c.Search != "" {
		filtered := make([]api.Host, 0)
		searchLower := strings.ToLower(c.Search)
		for _, host := range allHosts {
			if strings.Contains(strings.ToLower(host.Name), searchLower) ||
				strings.Contains(strings.ToLower(host.Description), searchLower) {
				filtered = append(filtered, host)
			}
		}
		allHosts = filtered
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allHosts)
	}

	hostData := make([]output.HostData, len(allHosts))
	for i, host := range allHosts {
		hostData[i] = output.HostData{
			ID:        host.ID,
			Name:      host.Name,
			Type:      host.Type,
			Model:     host.Model,
			Version:   host.Version,
			IPAddress: host.IPAddress,
			Status:    host.Status,
			SiteID:    host.SiteID,
			SiteName:  host.SiteName,
			Uptime:    host.Uptime,
		}
	}

	formatter.PrintHostsTable(hostData)
	return nil
}

type GetHostCmd struct {
	HostID string `arg:"" help:"Host ID to retrieve"`
}

func (c *GetHostCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	resp, err := g.appClient.GetHost(c.HostID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	host := resp.Data
	hostData := []output.HostData{{
		ID:         host.ID,
		Name:       host.Name,
		Type:       host.Type,
		Model:      host.Model,
		Version:    host.Version,
		IPAddress:  host.IPAddress,
		MACAddress: host.MACAddress,
		Status:     host.Status,
		SiteID:     host.SiteID,
		SiteName:   host.SiteName,
		Uptime:     host.Uptime,
	}}

	formatter.PrintHostsTable(hostData)
	return nil
}

type HostHealthCmd struct {
	HostID string `arg:"" help:"Host ID"`
}

func (c *HostHealthCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	resp, err := g.appClient.GetHostHealth(c.HostID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	formatter.PrintHealthTable(resp.Data)
	return nil
}

type HostStatsCmd struct {
	HostID string `arg:"" help:"Host ID"`
	Period string `help:"Stats period (day, week, month)" default:"day"`
}

func (c *HostStatsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	resp, err := g.appClient.GetHostStats(c.HostID, c.Period)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	formatter.PrintPerformanceTable(&resp.Data)
	return nil
}

type RestartHostCmd struct {
	HostID string `arg:"" help:"Host ID to restart"`
	Force  bool   `help:"Skip confirmation"`
}

func (c *RestartHostCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to restart host %s? (y/N): ", c.HostID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restart cancelled")
			return nil
		}
	}

	if err := g.appClient.RestartHost(c.HostID); err != nil {
		return err
	}

	fmt.Printf("Host %s restart initiated\n", c.HostID)
	return nil
}

// ========== DEVICES COMMANDS ==========

type DevicesCmd struct {
	List    ListDevicesCmd   `cmd:"" help:"List all devices for a site"`
	Get     GetDeviceCmd     `cmd:"" help:"Get a specific device"`
	Restart RestartDeviceCmd `cmd:"" help:"Restart a device"`
	Upgrade UpgradeDeviceCmd `cmd:"" help:"Upgrade device firmware"`
	Adopt   AdoptDeviceCmd   `cmd:"" help:"Adopt a new device"`
}

type ListDevicesCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	PageSize int    `help:"Number of devices per page (0 = fetch all)" default:"50"`
	Status   string `help:"Filter by status (online, offline, pending, etc.)"`
	Type     string `help:"Filter by device type (ap, switch, gateway, etc.)"`
}

func (c *ListDevicesCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allDevices []api.Device
	nextToken := ""

	for {
		resp, err := g.appClient.ListDevices(c.SiteID, c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allDevices = append(allDevices, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	// Filter by status and type if specified
	if c.Status != "" || c.Type != "" {
		filtered := make([]api.Device, 0)
		statusLower := strings.ToLower(c.Status)
		typeLower := strings.ToLower(c.Type)
		for _, device := range allDevices {
			if c.Status != "" && !strings.EqualFold(device.Status, statusLower) {
				continue
			}
			if c.Type != "" && !strings.EqualFold(device.Type, typeLower) {
				continue
			}
			filtered = append(filtered, device)
		}
		allDevices = filtered
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allDevices)
	}

	deviceData := make([]output.DeviceData, len(allDevices))
	for i, device := range allDevices {
		deviceData[i] = output.DeviceData{
			ID:           device.ID,
			Name:         device.Name,
			Type:         device.Type,
			Model:        device.Model,
			Version:      device.Version,
			MACAddress:   device.MACAddress,
			IPAddress:    device.IPAddress,
			Status:       device.Status,
			Adopted:      device.Adopted,
			Uptime:       device.Uptime,
			Clients:      device.Clients,
			Satisfaction: device.Satisfaction,
		}
	}

	formatter.PrintDevicesTable(deviceData)
	return nil
}

type GetDeviceCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	DeviceID string `arg:"" help:"Device ID"`
}

func (c *GetDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.DeviceID == "" {
		return &api.ValidationError{Message: "site ID and device ID are required"}
	}

	resp, err := g.appClient.GetDevice(c.SiteID, c.DeviceID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	device := resp.Data
	deviceData := []output.DeviceData{{
		ID:           device.ID,
		Name:         device.Name,
		Type:         device.Type,
		Model:        device.Model,
		Version:      device.Version,
		MACAddress:   device.MACAddress,
		IPAddress:    device.IPAddress,
		Status:       device.Status,
		Adopted:      device.Adopted,
		Uptime:       device.Uptime,
		Clients:      device.Clients,
		Satisfaction: device.Satisfaction,
		CPUUsage:     device.CPUUsage,
		MemoryUsage:  device.MemoryUsage,
	}}

	formatter.PrintDevicesTable(deviceData)
	return nil
}

type RestartDeviceCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	DeviceID string `arg:"" help:"Device ID to restart"`
	Force    bool   `help:"Skip confirmation"`
}

func (c *RestartDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.DeviceID == "" {
		return &api.ValidationError{Message: "site ID and device ID are required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to restart device %s? (y/N): ", c.DeviceID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restart cancelled")
			return nil
		}
	}

	if err := g.appClient.RestartDevice(c.SiteID, c.DeviceID); err != nil {
		return err
	}

	fmt.Printf("Device %s restart initiated\n", c.DeviceID)
	return nil
}

type UpgradeDeviceCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	DeviceID string `arg:"" help:"Device ID to upgrade"`
	Force    bool   `help:"Skip confirmation"`
}

func (c *UpgradeDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.DeviceID == "" {
		return &api.ValidationError{Message: "site ID and device ID are required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to upgrade firmware for device %s? (y/N): ", c.DeviceID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Upgrade cancelled")
			return nil
		}
	}

	if err := g.appClient.UpgradeDevice(c.SiteID, c.DeviceID); err != nil {
		return err
	}

	fmt.Printf("Device %s firmware upgrade initiated\n", c.DeviceID)
	return nil
}

type AdoptDeviceCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Device MAC address"`
}

func (c *AdoptDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.MACAddress == "" {
		return &api.ValidationError{Message: "site ID and MAC address are required"}
	}

	if err := g.appClient.AdoptDevice(c.SiteID, c.MACAddress); err != nil {
		return err
	}

	fmt.Printf("Device %s adoption initiated for site %s\n", c.MACAddress, c.SiteID)
	return nil
}

// ========== CLIENTS COMMANDS ==========

type ClientsCmd struct {
	List    ListClientsCmd   `cmd:"" help:"List all clients for a site"`
	Stats   ClientStatsCmd   `cmd:"" help:"Get client statistics"`
	Block   BlockClientCmd   `cmd:"" help:"Block a client"`
	Unblock UnblockClientCmd `cmd:"" help:"Unblock a client"`
}

type ListClientsCmd struct {
	SiteID       string `arg:"" help:"Site ID"`
	PageSize     int    `help:"Number of clients per page (0 = fetch all)" default:"50"`
	WiredOnly    bool   `help:"Show only wired clients"`
	WirelessOnly bool   `help:"Show only wireless clients"`
	Search       string `help:"Filter by hostname or name"`
}

func (c *ListClientsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	if c.WiredOnly && c.WirelessOnly {
		return &api.ValidationError{Message: "cannot use both --wired-only and --wireless-only"}
	}

	var allClients []api.NetworkClient
	nextToken := ""

	for {
		resp, err := g.appClient.ListClients(c.SiteID, c.PageSize, nextToken, c.WiredOnly, c.WirelessOnly)
		if err != nil {
			return err
		}

		allClients = append(allClients, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	if c.Search != "" {
		filtered := make([]api.NetworkClient, 0)
		searchLower := strings.ToLower(c.Search)
		for _, client := range allClients {
			if strings.Contains(strings.ToLower(client.Hostname), searchLower) ||
				strings.Contains(strings.ToLower(client.Name), searchLower) ||
				strings.Contains(strings.ToLower(client.MACAddress), searchLower) {
				filtered = append(filtered, client)
			}
		}
		allClients = filtered
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allClients)
	}

	clientData := make([]output.ClientData, len(allClients))
	for i, client := range allClients {
		connectionType := "Wireless"
		if client.IsWired {
			connectionType = "Wired"
		}
		clientData[i] = output.ClientData{
			ID:             client.ID,
			MACAddress:     client.MACAddress,
			IPAddress:      client.IPAddress,
			Hostname:       client.Hostname,
			Name:           client.Name,
			ConnectionType: connectionType,
			SSID:           client.SSID,
			Signal:         client.Signal,
			Satisfaction:   client.Satisfaction,
			Uptime:         client.Uptime,
			IsBlocked:      client.IsBlocked,
			IsGuest:        client.IsGuest,
		}
	}

	formatter.PrintClientsTable(clientData)
	return nil
}

type ClientStatsCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Client MAC address"`
}

func (c *ClientStatsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.MACAddress == "" {
		return &api.ValidationError{Message: "site ID and MAC address are required"}
	}

	resp, err := g.appClient.GetClientStats(c.SiteID, c.MACAddress)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	stats := resp.Data
	fmt.Printf("Client Statistics (%s):\n", c.MACAddress)
	fmt.Printf("  Rx Bytes: %s\n", formatBytes(stats.RxBytes))
	fmt.Printf("  Tx Bytes: %s\n", formatBytes(stats.TxBytes))
	fmt.Printf("  Rx Packets: %d\n", stats.RxPackets)
	fmt.Printf("  Tx Packets: %d\n", stats.TxPackets)
	fmt.Printf("  Signal Avg: %d dBm\n", stats.SignalAvg)
	fmt.Printf("  Satisfaction: %.1f%%\n", stats.Satisfaction)
	return nil
}

type BlockClientCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Client MAC address to block"`
	Force      bool   `help:"Skip confirmation"`
}

func (c *BlockClientCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.MACAddress == "" {
		return &api.ValidationError{Message: "site ID and MAC address are required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to block client %s? (y/N): ", c.MACAddress)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Block cancelled")
			return nil
		}
	}

	if err := g.appClient.BlockClient(c.SiteID, c.MACAddress, true); err != nil {
		return err
	}

	fmt.Printf("Client %s blocked successfully\n", c.MACAddress)
	return nil
}

type UnblockClientCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Client MAC address to unblock"`
}

func (c *UnblockClientCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.MACAddress == "" {
		return &api.ValidationError{Message: "site ID and MAC address are required"}
	}

	if err := g.appClient.BlockClient(c.SiteID, c.MACAddress, false); err != nil {
		return err
	}

	fmt.Printf("Client %s unblocked successfully\n", c.MACAddress)
	return nil
}

// ========== WLANS COMMANDS ==========

type WLANsCmd struct {
	List   ListWLANsCmd  `cmd:"" help:"List all WLANs for a site"`
	Get    GetWLANCmd    `cmd:"" help:"Get a specific WLAN"`
	Create CreateWLANCmd `cmd:"" help:"Create a new WLAN"`
	Update UpdateWLANCmd `cmd:"" help:"Update a WLAN"`
	Delete DeleteWLANCmd `cmd:"" help:"Delete a WLAN"`
}

type ListWLANsCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	PageSize int    `help:"Number of WLANs per page (0 = fetch all)" default:"50"`
}

func (c *ListWLANsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allWLANs []api.WLAN
	nextToken := ""

	for {
		resp, err := g.appClient.ListWLANs(c.SiteID, c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allWLANs = append(allWLANs, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allWLANs)
	}

	wlanData := make([]output.WLANData, len(allWLANs))
	for i, wlan := range allWLANs {
		security := wlan.Security
		if wlan.WPA3Support {
			security += " (WPA3)"
		}
		wlanData[i] = output.WLANData{
			ID:       wlan.ID,
			Name:     wlan.Name,
			SSID:     wlan.SSID,
			Security: security,
			Enabled:  wlan.Enabled,
			Hidden:   wlan.Hidden,
			VLAN:     wlan.VLAN,
			Band:     wlan.Band,
		}
	}

	formatter.PrintWLANsTable(wlanData)
	return nil
}

type GetWLANCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	WLANID string `arg:"" help:"WLAN ID"`
}

func (c *GetWLANCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.WLANID == "" {
		return &api.ValidationError{Message: "site ID and WLAN ID are required"}
	}

	resp, err := g.appClient.GetWLAN(c.SiteID, c.WLANID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	wlan := resp.Data
	security := wlan.Security
	if wlan.WPA3Support {
		security += " (WPA3)"
	}
	wlanData := []output.WLANData{{
		ID:       wlan.ID,
		Name:     wlan.Name,
		SSID:     wlan.SSID,
		Security: security,
		Enabled:  wlan.Enabled,
		Hidden:   wlan.Hidden,
		VLAN:     wlan.VLAN,
		Band:     wlan.Band,
	}}

	formatter.PrintWLANsTable(wlanData)
	return nil
}

type CreateWLANCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	Name     string `arg:"" help:"WLAN name"`
	SSID     string `arg:"" help:"Network SSID"`
	Password string `help:"Network password (for WPA/WPA2)"`
	Security string `help:"Security type (wpapsk, wpaeap, wep, open)" default:"wpapsk"`
	VLAN     int    `help:"VLAN ID"`
	Band     string `help:"WiFi band (2g, 5g, both)" default:"both"`
	Hidden   bool   `help:"Hide SSID"`
	WPA3     bool   `help:"Enable WPA3 support"`
}

func (c *CreateWLANCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.Name == "" || c.SSID == "" {
		return &api.ValidationError{Message: "site ID, name, and SSID are required"}
	}

	req := api.CreateWLANRequest{
		Name:        c.Name,
		SSID:        c.SSID,
		Security:    c.Security,
		Password:    c.Password,
		VLAN:        c.VLAN,
		Band:        c.Band,
		Hidden:      c.Hidden,
		WPA3Support: c.WPA3,
	}

	resp, err := g.appClient.CreateWLAN(c.SiteID, req)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	fmt.Printf("WLAN created successfully:\n")
	wlan := resp.Data
	security := wlan.Security
	if wlan.WPA3Support {
		security += " (WPA3)"
	}
	wlanData := []output.WLANData{{
		ID:       wlan.ID,
		Name:     wlan.Name,
		SSID:     wlan.SSID,
		Security: security,
		Enabled:  wlan.Enabled,
		Hidden:   wlan.Hidden,
		VLAN:     wlan.VLAN,
		Band:     wlan.Band,
	}}
	formatter.PrintWLANsTable(wlanData)
	return nil
}

type UpdateWLANCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	WLANID   string `arg:"" help:"WLAN ID"`
	Name     string `help:"New WLAN name"`
	Password string `help:"New password"`
	Enabled  *bool  `help:"Enable/disable WLAN"`
	Hidden   *bool  `help:"Hide/unhide SSID"`
}

func (c *UpdateWLANCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.WLANID == "" {
		return &api.ValidationError{Message: "site ID and WLAN ID are required"}
	}

	req := api.UpdateWLANRequest{}
	if c.Name != "" {
		req.Name = c.Name
	}
	if c.Password != "" {
		req.Password = c.Password
	}
	if c.Enabled != nil {
		req.Enabled = c.Enabled
	}
	if c.Hidden != nil {
		req.Hidden = c.Hidden
	}

	resp, err := g.appClient.UpdateWLAN(c.SiteID, c.WLANID, req)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	fmt.Printf("WLAN updated successfully:\n")
	wlan := resp.Data
	security := wlan.Security
	if wlan.WPA3Support {
		security += " (WPA3)"
	}
	wlanData := []output.WLANData{{
		ID:       wlan.ID,
		Name:     wlan.Name,
		SSID:     wlan.SSID,
		Security: security,
		Enabled:  wlan.Enabled,
		Hidden:   wlan.Hidden,
		VLAN:     wlan.VLAN,
		Band:     wlan.Band,
	}}
	formatter.PrintWLANsTable(wlanData)
	return nil
}

type DeleteWLANCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	WLANID string `arg:"" help:"WLAN ID to delete"`
	Force  bool   `help:"Skip confirmation"`
}

func (c *DeleteWLANCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.WLANID == "" {
		return &api.ValidationError{Message: "site ID and WLAN ID are required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to delete WLAN %s? (y/N): ", c.WLANID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	if err := g.appClient.DeleteWLAN(c.SiteID, c.WLANID); err != nil {
		return err
	}

	fmt.Printf("WLAN %s deleted successfully\n", c.WLANID)
	return nil
}

// ========== ALERTS COMMANDS ==========

type AlertsCmd struct {
	List    ListAlertsCmd   `cmd:"" help:"List alerts"`
	Ack     AckAlertCmd     `cmd:"" help:"Acknowledge an alert"`
	Archive ArchiveAlertCmd `cmd:"" help:"Archive an alert"`
}

type ListAlertsCmd struct {
	SiteID   string `help:"Filter by site ID (optional)"`
	PageSize int    `help:"Number of alerts per page (0 = fetch all)" default:"50"`
	Archived bool   `help:"Show archived alerts"`
}

func (c *ListAlertsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	var allAlerts []api.Alert
	nextToken := ""

	for {
		resp, err := g.appClient.ListAlerts(c.SiteID, c.PageSize, nextToken, c.Archived)
		if err != nil {
			return err
		}

		allAlerts = append(allAlerts, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allAlerts)
	}

	alertData := make([]output.AlertData, len(allAlerts))
	for i, alert := range allAlerts {
		alertData[i] = output.AlertData{
			ID:           alert.ID,
			Type:         alert.Type,
			Message:      alert.Message,
			SiteID:       alert.SiteID,
			DeviceID:     alert.DeviceID,
			DeviceName:   alert.DeviceName,
			Severity:     alert.Severity,
			Timestamp:    alert.Timestamp,
			Acknowledged: alert.Acknowledged,
			Archived:     alert.Archived,
		}
	}

	formatter.PrintAlertsTable(alertData)
	return nil
}

type AckAlertCmd struct {
	SiteID  string `arg:"" help:"Site ID"`
	AlertID string `arg:"" help:"Alert ID to acknowledge"`
}

func (c *AckAlertCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.AlertID == "" {
		return &api.ValidationError{Message: "site ID and alert ID are required"}
	}

	if err := g.appClient.AcknowledgeAlert(c.SiteID, c.AlertID); err != nil {
		return err
	}

	fmt.Printf("Alert %s acknowledged\n", c.AlertID)
	return nil
}

type ArchiveAlertCmd struct {
	SiteID  string `arg:"" help:"Site ID"`
	AlertID string `arg:"" help:"Alert ID to archive"`
}

func (c *ArchiveAlertCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" || c.AlertID == "" {
		return &api.ValidationError{Message: "site ID and alert ID are required"}
	}

	if err := g.appClient.ArchiveAlert(c.SiteID, c.AlertID); err != nil {
		return err
	}

	fmt.Printf("Alert %s archived\n", c.AlertID)
	return nil
}

// ========== EVENTS COMMANDS ==========

type EventsCmd struct {
	List ListEventsCmd `cmd:"" help:"List events"`
}

type ListEventsCmd struct {
	SiteID   string `help:"Filter by site ID (optional)"`
	PageSize int    `help:"Number of events per page (0 = fetch all)" default:"50"`
}

func (c *ListEventsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	var allEvents []api.Event
	nextToken := ""

	for {
		resp, err := g.appClient.ListEvents(c.SiteID, c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allEvents = append(allEvents, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allEvents)
	}

	eventData := make([]output.EventData, len(allEvents))
	for i, event := range allEvents {
		eventData[i] = output.EventData{
			ID:        event.ID,
			Type:      event.Type,
			Message:   event.Message,
			SiteID:    event.SiteID,
			DeviceID:  event.DeviceID,
			ClientID:  event.ClientID,
			UserID:    event.UserID,
			Timestamp: event.Timestamp,
		}
	}

	formatter.PrintEventsTable(eventData)
	return nil
}

// ========== NETWORKS COMMANDS ==========

type NetworksCmd struct {
	List ListNetworksCmd `cmd:"" help:"List networks for a site"`
}

type ListNetworksCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	PageSize int    `help:"Number of networks per page (0 = fetch all)" default:"50"`
}

func (c *ListNetworksCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allNetworks []api.Network
	nextToken := ""

	for {
		resp, err := g.appClient.ListNetworks(c.SiteID, c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allNetworks = append(allNetworks, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(allNetworks)
	}

	networkData := make([]output.NetworkData, len(allNetworks))
	for i, network := range allNetworks {
		networkData[i] = output.NetworkData{
			ID:           network.ID,
			Name:         network.Name,
			Purpose:      network.Purpose,
			VLAN:         network.VLAN,
			Subnet:       network.Subnet,
			GatewayIP:    network.GatewayIP,
			DHCPEnabled:  network.DHCPEnabled,
			DHCPStart:    network.DHCPStart,
			DHCPStop:     network.DHCPStop,
			NetworkGroup: network.NetworkGroup,
			DomainName:   network.DomainName,
		}
	}

	formatter.PrintNetworksTable(networkData)
	return nil
}

// ========== WHOAMI COMMAND ==========

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

// ========== VERSION COMMAND ==========

type VersionCmd struct {
	Check bool `help:"Check for updates"`
}

func (c *VersionCmd) Run(g *Globals) error {
	version := "dev"
	gitCommit := "unknown"
	buildTime := "unknown"

	output.PrintVersion(version, gitCommit, buildTime, c.Check)
	return nil
}

// ========== RUN FUNCTION ==========

// Run parses CLI args and executes the appropriate command
func Run(args []string, version, gitCommit, buildTime string) (int, error) {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("usm"),
		kong.Description("UniFi Site Manager CLI - Full-featured management for UniFi networks"),
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

	err = ctx.Run(&cli.Globals)
	if err != nil {
		return api.GetExitCode(err), err
	}

	return api.ExitSuccess, nil
}

// ========== UTILITY FUNCTIONS ==========

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
