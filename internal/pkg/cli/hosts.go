package cli

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// HostsCmd is the parent command for host operations
type HostsCmd struct {
	List    ListHostsCmd   `cmd:"" help:"List all hosts"`
	Get     GetHostCmd     `cmd:"" help:"Get a specific host"`
	Health  HostHealthCmd  `cmd:"" help:"Get host health"`
	Stats   HostStatsCmd   `cmd:"" help:"Get host statistics"`
	Restart RestartHostCmd `cmd:"" help:"Restart a host"`
}

// ListHostsCmd handles listing hosts
type ListHostsCmd struct {
	PageSize int    `help:"Number of hosts per page (0 = fetch all)" default:"50"`
	Search   string `help:"Filter hosts by name"`
}

func (c *ListHostsCmd) Run(ctx *CLIContext) error {
	var allHosts []api.Host
	nextToken := ""

	for {
		resp, err := ctx.Client.ListHosts(c.PageSize, nextToken)
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
			if strings.Contains(strings.ToLower(host.Name), searchLower) {
				filtered = append(filtered, host)
			}
		}
		allHosts = filtered
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(allHosts)
	}

	hostData := make([]output.HostData, len(allHosts))
	for i, host := range allHosts {
		hostData[i] = output.HostData{
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
		}
	}

	formatter.PrintHostsTable(hostData)
	return nil
}

// GetHostCmd handles getting a specific host
type GetHostCmd struct {
	HostID string `arg:"" help:"Host ID to retrieve"`
}

func (c *GetHostCmd) Run(ctx *CLIContext) error {
	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	resp, err := ctx.Client.GetHost(c.HostID)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// HostHealthCmd handles getting host health
type HostHealthCmd struct {
	HostID string `arg:"" help:"Host ID"`
}

func (c *HostHealthCmd) Run(ctx *CLIContext) error {
	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	resp, err := ctx.Client.GetHostHealth(c.HostID)
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

// HostStatsCmd handles getting host statistics
type HostStatsCmd struct {
	HostID string `arg:"" help:"Host ID"`
	Period string `help:"Stats period (day, week, month)" default:"day"`
}

func (c *HostStatsCmd) Run(ctx *CLIContext) error {
	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	resp, err := ctx.Client.GetHostStats(c.HostID, c.Period)
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

// RestartHostCmd handles restarting a host
type RestartHostCmd struct {
	HostID string `arg:"" help:"Host ID to restart"`
	Force  bool   `help:"Skip confirmation"`
}

func (c *RestartHostCmd) Run(ctx *CLIContext) error {
	if c.HostID == "" {
		return &api.ValidationError{Message: "host ID is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to restart host %s? (y/N): ", c.HostID)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restart cancelled")
			return nil
		}
	}

	if err := ctx.Client.RestartHost(c.HostID); err != nil {
		return err
	}

	fmt.Printf("Host %s restart initiated\n", c.HostID)
	return nil
}
