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
	List   ListSitesCmd  `cmd:"" help:"List all sites"`
	Get    GetSiteCmd    `cmd:"" help:"Get a specific site"`
	Create CreateSiteCmd `cmd:"" help:"Create a new site"`
	Update UpdateSiteCmd `cmd:"" help:"Update a site"`
	Delete DeleteSiteCmd `cmd:"" help:"Delete a site"`
	Health SiteHealthCmd `cmd:"" help:"Get site health"`
	Stats  SiteStatsCmd  `cmd:"" help:"Get site statistics"`
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
