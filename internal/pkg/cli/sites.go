package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// SitesCmd is the parent command for site operations
type SitesCmd struct {
	List    ListSitesCmd    `cmd:"" help:"List all sites"`
	Get     GetSiteCmd      `cmd:"" help:"Get a specific site"`
	Create  CreateSiteCmd   `cmd:"" help:"Create a new site"`
	Update  UpdateSiteCmd   `cmd:"" help:"Update a site"`
	Delete  DeleteSiteCmd   `cmd:"" help:"Delete a site"`
	Health  SiteHealthCmd   `cmd:"" help:"Get site health"`
	Stats   SiteStatsCmd    `cmd:"" help:"Get site statistics"`
	Exec    SitesExecCmd    `cmd:"" help:"Execute a command across multiple sites"`
	Compare SitesCompareCmd `cmd:"" help:"Compare configurations between sites"`
	Report  SitesReportCmd  `cmd:"" help:"Generate a summary report of all sites"`
}

// ListSitesCmd handles listing sites
type ListSitesCmd struct {
	PageSize int    `help:"Number of sites per page (0 = fetch all)" default:"50"`
	Search   string `help:"Filter sites by name/description"`
}

func (c *ListSitesCmd) Run(ctx *CLIContext) error {
	var allSites []api.Site
	nextToken := ""

	for {
		resp, err := ctx.Client.ListSites(c.PageSize, nextToken)
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

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// GetSiteCmd handles getting a specific site
type GetSiteCmd struct {
	SiteID string `arg:"" help:"Site ID to retrieve"`
}

func (c *GetSiteCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	resp, err := ctx.Client.GetSite(c.SiteID)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// CreateSiteCmd handles creating a new site
type CreateSiteCmd struct {
	Name        string `arg:"" help:"Site name"`
	Description string `help:"Site description"`
}

func (c *CreateSiteCmd) Run(ctx *CLIContext) error {
	if c.Name == "" {
		return &api.ValidationError{Message: "site name is required"}
	}

	req := api.CreateSiteRequest{
		Name:        c.Name,
		Description: c.Description,
	}

	resp, err := ctx.Client.CreateSite(req)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// UpdateSiteCmd handles updating a site
type UpdateSiteCmd struct {
	SiteID      string `arg:"" help:"Site ID to update"`
	Name        string `help:"New site name"`
	Description string `help:"New site description"`
}

func (c *UpdateSiteCmd) Run(ctx *CLIContext) error {
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

	resp, err := ctx.Client.UpdateSite(c.SiteID, req)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// DeleteSiteCmd handles deleting a site
type DeleteSiteCmd struct {
	SiteID string `arg:"" help:"Site ID to delete"`
	Force  bool   `help:"Skip confirmation"`
}

func (c *DeleteSiteCmd) Run(ctx *CLIContext) error {
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

	if err := ctx.Client.DeleteSite(c.SiteID); err != nil {
		return err
	}

	fmt.Printf("Site %s deleted successfully\n", c.SiteID)
	return nil
}

// SiteHealthCmd handles getting site health
type SiteHealthCmd struct {
	SiteID string `arg:"" help:"Site ID"`
}

func (c *SiteHealthCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	resp, err := ctx.Client.GetSiteHealth(c.SiteID)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	formatter.PrintHealthTable(resp.Data)
	return nil
}

// SiteStatsCmd handles getting site statistics
type SiteStatsCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	Period string `help:"Stats period (day, week, month)" default:"day"`
}

func (c *SiteStatsCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	resp, err := ctx.Client.GetSiteStats(c.SiteID, c.Period)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	formatter.PrintPerformanceTable(&resp.Data)
	return nil
}

// ========== MULTI-SITE COMMANDS ==========

// SitesExecCmd executes a command across multiple sites
type SitesExecCmd struct {
	Command  string   `help:"Command to execute (e.g., 'devices list', 'clients list')" required:""`
	Sites    []string `help:"Comma-separated list of site IDs"`
	AllSites bool     `help:"Execute on all sites"`
}

func (c *SitesExecCmd) Run(ctx *CLIContext) error {
	// Get sites to execute on
	var targetSites []api.Site

	if c.AllSites {
		// Fetch all sites
		resp, err := ctx.Client.ListSites(0, "")
		if err != nil {
			return fmt.Errorf("failed to list sites: %w", err)
		}
		targetSites = resp.Data
	} else if len(c.Sites) > 0 {
		// Validate and fetch specified sites
		for _, siteID := range c.Sites {
			resp, err := ctx.Client.GetSite(siteID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to get site %s: %v\n", siteID, err)
				continue
			}
			targetSites = append(targetSites, resp.Data)
		}
	} else {
		return &api.ValidationError{Message: "either --sites or --all-sites must be specified"}
	}

	if len(targetSites) == 0 {
		return &api.ValidationError{Message: "no valid sites specified"}
	}

	fmt.Printf("Executing: %s\n\n", c.Command)

	// Parse command
	parts := strings.Fields(c.Command)
	if len(parts) == 0 {
		return &api.ValidationError{Message: "command cannot be empty"}
	}

	commandType := strings.ToLower(parts[0])
	subCommand := ""
	if len(parts) > 1 {
		subCommand = strings.ToLower(parts[1])
	}

	successCount := 0
	failCount := 0

	// Execute command on each site
	for _, site := range targetSites {
		fmt.Printf("=== Site: %s (%s) ===\n", site.Name, site.ID)

		err := c.executeCommand(ctx, site.ID, commandType, subCommand)
		if err != nil {
			fmt.Printf("✗ ERROR: %v\n", err)
			failCount++
		} else {
			successCount++
		}
		fmt.Println()
	}

	fmt.Printf("Execution complete. %d/%d sites succeeded.\n", successCount, len(targetSites))
	return nil
}

func (c *SitesExecCmd) executeCommand(ctx *CLIContext, siteID, commandType, subCommand string) error {
	switch commandType {
	case "devices":
		return c.executeDevicesCommand(ctx, siteID, subCommand)
	case "clients":
		return c.executeClientsCommand(ctx, siteID, subCommand)
	case "alerts":
		return c.executeAlertsCommand(ctx, siteID, subCommand)
	case "networks":
		return c.executeNetworksCommand(ctx, siteID, subCommand)
	case "wlans":
		return c.executeWLANsCommand(ctx, siteID, subCommand)
	default:
		return fmt.Errorf("unsupported command type: %s", commandType)
	}
}

func (c *SitesExecCmd) executeDevicesCommand(ctx *CLIContext, siteID, subCommand string) error {
	switch subCommand {
	case "list", "":
		resp, err := ctx.Client.ListDevices(siteID, 0, "")
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			fmt.Println("No devices found")
			return nil
		}
		for _, device := range resp.Data {
			status := "✓"
			if device.Status == "OFFLINE" || device.Status == "offline" {
				status = "✗"
			}
			fmt.Printf("%s %-15s %-15s %s\n", status, device.Name, device.IPAddress, device.Status)
		}
	default:
		return fmt.Errorf("unsupported devices subcommand: %s", subCommand)
	}
	return nil
}

func (c *SitesExecCmd) executeClientsCommand(ctx *CLIContext, siteID, subCommand string) error {
	switch subCommand {
	case "list", "":
		resp, err := ctx.Client.ListClients(siteID, 0, "", false, false)
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			fmt.Println("No clients found")
			return nil
		}
		for _, client := range resp.Data {
			name := client.Name
			if name == "" {
				name = client.Hostname
			}
			if name == "" {
				name = client.MACAddress
			}
			connType := "wireless"
			if client.IsWired {
				connType = "wired"
			}
			fmt.Printf("• %-20s %-15s %s\n", name, client.IPAddress, connType)
		}
	default:
		return fmt.Errorf("unsupported clients subcommand: %s", subCommand)
	}
	return nil
}

func (c *SitesExecCmd) executeAlertsCommand(ctx *CLIContext, siteID, subCommand string) error {
	switch subCommand {
	case "list", "":
		resp, err := ctx.Client.ListAlerts(siteID, 0, "", false)
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			fmt.Println("No alerts found")
			return nil
		}
		for _, alert := range resp.Data {
			fmt.Printf("• [%s] %s: %s\n", alert.Severity, alert.Type, alert.Message)
		}
	default:
		return fmt.Errorf("unsupported alerts subcommand: %s", subCommand)
	}
	return nil
}

func (c *SitesExecCmd) executeNetworksCommand(ctx *CLIContext, siteID, subCommand string) error {
	switch subCommand {
	case "list", "":
		resp, err := ctx.Client.ListNetworks(siteID, 0, "")
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			fmt.Println("No networks found")
			return nil
		}
		for _, network := range resp.Data {
			fmt.Printf("• %-20s VLAN:%d %s\n", network.Name, network.VLAN, network.Subnet)
		}
	default:
		return fmt.Errorf("unsupported networks subcommand: %s", subCommand)
	}
	return nil
}

func (c *SitesExecCmd) executeWLANsCommand(ctx *CLIContext, siteID, subCommand string) error {
	switch subCommand {
	case "list", "":
		resp, err := ctx.Client.ListWLANs(siteID, 0, "")
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			fmt.Println("No WLANs found")
			return nil
		}
		for _, wlan := range resp.Data {
			status := "enabled"
			if !wlan.Enabled {
				status = "disabled"
			}
			fmt.Printf("• %-20s (SSID: %s) [%s]\n", wlan.Name, wlan.SSID, status)
		}
	default:
		return fmt.Errorf("unsupported wlans subcommand: %s", subCommand)
	}
	return nil
}

// SitesCompareCmd compares configurations between sites
type SitesCompareCmd struct {
	Sites  []string `help:"Comma-separated list of site IDs to compare (exactly 2 required)" required:""`
	Output string   `help:"Output format: table, json" default:"table" enum:"table,json"`
}

func (c *SitesCompareCmd) Run(ctx *CLIContext) error {
	if len(c.Sites) < 2 {
		return &api.ValidationError{Message: "at least 2 sites required for comparison"}
	}

	// Get site details
	siteData := make(map[string]*siteComparisonData)
	for _, siteID := range c.Sites {
		resp, err := ctx.Client.GetSite(siteID)
		if err != nil {
			return fmt.Errorf("failed to get site %s: %w", siteID, err)
		}
		site := resp.Data

		// Fetch devices
		devicesResp, err := ctx.Client.ListDevices(siteID, 0, "")
		if err != nil {
			return fmt.Errorf("failed to list devices for site %s: %w", siteID, err)
		}

		// Fetch clients
		clientsResp, err := ctx.Client.ListClients(siteID, 0, "", false, false)
		if err != nil {
			return fmt.Errorf("failed to list clients for site %s: %w", siteID, err)
		}

		// Count device types
		deviceCounts := make(map[string]int)
		onlineDevices := 0
		offlineDevices := 0
		firmwareVersions := make(map[string]string)
		for _, device := range devicesResp.Data {
			deviceCounts[device.Type]++
			if device.Status == "ONLINE" || device.Status == "online" {
				onlineDevices++
			} else {
				offlineDevices++
			}
			if device.Version != "" {
				firmwareVersions[device.Model] = device.Version
			}
		}

		siteData[siteID] = &siteComparisonData{
			SiteID:           siteID,
			SiteName:         site.Name,
			DeviceCount:      len(devicesResp.Data),
			OnlineDevices:    onlineDevices,
			OfflineDevices:   offlineDevices,
			ClientCount:      len(clientsResp.Data),
			DeviceCounts:     deviceCounts,
			FirmwareVersions: firmwareVersions,
		}
	}

	if c.Output == "json" {
		formatter := ctx.getFormatter()
		return formatter.PrintJSON(siteData)
	}

	// Print comparison table
	fmt.Println("Site Comparison")
	fmt.Println("================")
	fmt.Println()

	// Print overview
	fmt.Println("Overview:")
	for _, siteID := range c.Sites {
		data := siteData[siteID]
		fmt.Printf("  %-20s: %d devices (%d online, %d offline) | %d clients\n",
			data.SiteName, data.DeviceCount, data.OnlineDevices, data.OfflineDevices, data.ClientCount)
	}
	fmt.Println()

	// Print device type comparison
	fmt.Println("Device Types:")
	allTypes := make(map[string]bool)
	for _, data := range siteData {
		for deviceType := range data.DeviceCounts {
			allTypes[deviceType] = true
		}
	}

	typeList := make([]string, 0, len(allTypes))
	for t := range allTypes {
		typeList = append(typeList, t)
	}

	for _, deviceType := range typeList {
		fmt.Printf("  %-15s:", deviceType)
		for _, siteID := range c.Sites {
			data := siteData[siteID]
			count := data.DeviceCounts[deviceType]
			fmt.Printf("  %-8d", count)
		}
		fmt.Println()
	}
	fmt.Println()

	// Check for mismatches
	fmt.Println("Firmware Versions:")
	firmwareMismatch := false
	for model, versions := range c.collectFirmwareVersions(siteData) {
		if len(versions) > 1 {
			firmwareMismatch = true
			fmt.Printf("  ⚠ %-20s: ", model)
			first := true
			for siteID, version := range versions {
				if !first {
					fmt.Print(" vs ")
				}
				fmt.Printf("%s@%s", siteData[siteID].SiteName, version)
				first = false
			}
			fmt.Println(" ← MISMATCH")
		}
	}

	if !firmwareMismatch {
		fmt.Println("  ✓ All firmware versions match across sites")
	}

	return nil
}

type siteComparisonData struct {
	SiteID           string            `json:"site_id"`
	SiteName         string            `json:"site_name"`
	DeviceCount      int               `json:"device_count"`
	OnlineDevices    int               `json:"online_devices"`
	OfflineDevices   int               `json:"offline_devices"`
	ClientCount      int               `json:"client_count"`
	DeviceCounts     map[string]int    `json:"device_counts"`
	FirmwareVersions map[string]string `json:"firmware_versions"`
}

func (c *SitesCompareCmd) collectFirmwareVersions(siteData map[string]*siteComparisonData) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for siteID, data := range siteData {
		for model, version := range data.FirmwareVersions {
			if result[model] == nil {
				result[model] = make(map[string]string)
			}
			result[model][siteID] = version
		}
	}
	return result
}

// SitesReportCmd generates a summary report of all sites
type SitesReportCmd struct {
	Output string `help:"Output format: table, json" default:"table" enum:"table,json"`
}

type siteReportData struct {
	SiteID         string `json:"site_id"`
	SiteName       string `json:"site_name"`
	DeviceCount    int    `json:"device_count"`
	OnlineDevices  int    `json:"online_devices"`
	OfflineDevices int    `json:"offline_devices"`
	ClientCount    int    `json:"client_count"`
	HasIssues      bool   `json:"has_issues"`
	IssueCount     int    `json:"issue_count"`
}

type multiSiteReport struct {
	TotalSites      int               `json:"total_sites"`
	TotalDevices    int               `json:"total_devices"`
	TotalClients    int               `json:"total_clients"`
	OnlineDevices   int               `json:"online_devices"`
	OfflineDevices  int               `json:"offline_devices"`
	SitesWithIssues int               `json:"sites_with_issues"`
	Sites           []*siteReportData `json:"sites"`
}

func (c *SitesReportCmd) Run(ctx *CLIContext) error {
	// Fetch all sites
	resp, err := ctx.Client.ListSites(0, "")
	if err != nil {
		return fmt.Errorf("failed to list sites: %w", err)
	}

	if len(resp.Data) == 0 {
		return &api.ValidationError{Message: "no sites found"}
	}

	report := &multiSiteReport{
		TotalSites: len(resp.Data),
		Sites:      make([]*siteReportData, 0, len(resp.Data)),
	}

	// Collect data from each site
	for _, site := range resp.Data {
		siteReport := &siteReportData{
			SiteID:   site.ID,
			SiteName: site.Name,
		}

		// Fetch devices
		devicesResp, err := ctx.Client.ListDevices(site.ID, 0, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to list devices for site %s: %v\n", site.ID, err)
		} else {
			siteReport.DeviceCount = len(devicesResp.Data)
			report.TotalDevices += len(devicesResp.Data)

			for _, device := range devicesResp.Data {
				if device.Status == "ONLINE" || device.Status == "online" {
					siteReport.OnlineDevices++
					report.OnlineDevices++
				} else {
					siteReport.OfflineDevices++
					report.OfflineDevices++
					siteReport.HasIssues = true
					siteReport.IssueCount++
				}
			}
		}

		// Fetch clients
		clientsResp, err := ctx.Client.ListClients(site.ID, 0, "", false, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to list clients for site %s: %v\n", site.ID, err)
		} else {
			siteReport.ClientCount = len(clientsResp.Data)
			report.TotalClients += len(clientsResp.Data)
		}

		// Fetch alerts for additional issue detection
		alertsResp, err := ctx.Client.ListAlerts(site.ID, 0, "", false)
		if err == nil && len(alertsResp.Data) > 0 {
			unacknowledgedCount := 0
			for _, alert := range alertsResp.Data {
				if !alert.Acknowledged {
					unacknowledgedCount++
				}
			}
			if unacknowledgedCount > 0 {
				siteReport.HasIssues = true
				siteReport.IssueCount += unacknowledgedCount
			}
		}

		if siteReport.HasIssues {
			report.SitesWithIssues++
		}

		report.Sites = append(report.Sites, siteReport)
	}

	if c.Output == "json" {
		formatter := ctx.getFormatter()
		return formatter.PrintJSON(report)
	}

	// Print table report
	fmt.Println("Multi-Site Report")
	fmt.Println("=================")
	fmt.Printf("Sites monitored: %d\n", report.TotalSites)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total devices:    %d\n", report.TotalDevices)
	fmt.Printf("  Total clients:    %d\n", report.TotalClients)
	fmt.Printf("  Online devices:   %d\n", report.OnlineDevices)
	fmt.Printf("  Offline devices:  %d\n", report.OfflineDevices)
	if report.SitesWithIssues > 0 {
		fmt.Printf("  Alerts:           %d site(s) have issues\n", report.SitesWithIssues)
	} else {
		fmt.Println("  Alerts:           ✓ All sites healthy")
	}
	fmt.Println()
	fmt.Println("Per-Site Breakdown:")
	for _, site := range report.Sites {
		status := "✓ healthy"
		if site.HasIssues {
			status = fmt.Sprintf("⚠ %d issue(s)", site.IssueCount)
		}
		fmt.Printf("  %-15s %3d devices | %4d clients | %s\n",
			site.SiteName, site.DeviceCount, site.ClientCount, status)
	}

	return nil
}
