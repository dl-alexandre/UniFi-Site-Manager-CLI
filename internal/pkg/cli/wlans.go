package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// WLANsCmd is the parent command for WLAN operations
type WLANsCmd struct {
	List   ListWLANsCmd  `cmd:"" help:"List all WLANs"`
	Get    GetWLANCmd    `cmd:"" help:"Get a specific WLAN"`
	Create CreateWLANCmd `cmd:"" help:"Create a new WLAN"`
	Update UpdateWLANCmd `cmd:"" help:"Update a WLAN"`
	Delete DeleteWLANCmd `cmd:"" help:"Delete a WLAN"`
}

// ListWLANsCmd handles listing WLANs
type ListWLANsCmd struct {
	SiteID   string `arg:"" help:"Site ID to list WLANs for"`
	PageSize int    `help:"Number of WLANs per page (0 = fetch all)" default:"50"`
	Search   string `help:"Filter WLANs by name or SSID"`
}

func (c *ListWLANsCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allWLANs []api.WLAN
	nextToken := ""

	for {
		resp, err := ctx.Client.ListWLANs(c.SiteID, c.PageSize, nextToken)
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

	if c.Search != "" {
		filtered := make([]api.WLAN, 0)
		searchLower := strings.ToLower(c.Search)
		for _, wlan := range allWLANs {
			if strings.Contains(strings.ToLower(wlan.Name), searchLower) ||
				strings.Contains(strings.ToLower(wlan.SSID), searchLower) {
				filtered = append(filtered, wlan)
			}
		}
		allWLANs = filtered
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(allWLANs)
	}

	wlanData := make([]output.WLANData, len(allWLANs))
	for i, wlan := range allWLANs {
		wlanData[i] = output.WLANData{
			ID:       wlan.ID,
			Name:     wlan.Name,
			SSID:     wlan.SSID,
			Security: wlan.Security,
			Enabled:  wlan.Enabled,
			Hidden:   wlan.Hidden,
			VLAN:     wlan.VLAN,
			Band:     wlan.Band,
		}
	}

	formatter.PrintWLANsTable(wlanData)
	return nil
}

// GetWLANCmd handles getting a specific WLAN
type GetWLANCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	WLANID string `arg:"" help:"WLAN ID to retrieve"`
}

func (c *GetWLANCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.WLANID == "" {
		return &api.ValidationError{Message: "WLAN ID is required"}
	}

	resp, err := ctx.Client.GetWLAN(c.SiteID, c.WLANID)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	wlan := resp.Data
	wlanData := []output.WLANData{{
		ID:       wlan.ID,
		Name:     wlan.Name,
		SSID:     wlan.SSID,
		Security: wlan.Security,
		Enabled:  wlan.Enabled,
		Hidden:   wlan.Hidden,
		VLAN:     wlan.VLAN,
		Band:     wlan.Band,
	}}

	formatter.PrintWLANsTable(wlanData)
	return nil
}

// CreateWLANCmd handles creating a new WLAN
type CreateWLANCmd struct {
	SiteID          string `arg:"" help:"Site ID to create WLAN in"`
	Name            string `arg:"" help:"WLAN name"`
	SSID            string `help:"Network SSID (defaults to name if not specified)"`
	Security        string `help:"Security type (wpapsk, wpaeap, etc.)" default:"wpapsk"`
	Password        string `help:"Network password (for PSK security)"`
	VLAN            int    `help:"VLAN ID"`
	Band            string `help:"Band (2g, 5g, both)" default:"both"`
	Hidden          bool   `help:"Hide SSID"`
	PMFMode         string `help:"PMF mode (optional, required, disabled)" default:"optional"`
	WPA3Support     bool   `help:"Enable WPA3 support"`
	MACFilter       bool   `help:"Enable MAC filtering"`
	MACFilterPolicy string `help:"MAC filter policy (allow, deny)" default:"allow"`
}

func (c *CreateWLANCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.Name == "" {
		return &api.ValidationError{Message: "WLAN name is required"}
	}

	ssid := c.SSID
	if ssid == "" {
		ssid = c.Name
	}

	req := api.CreateWLANRequest{
		Name:            c.Name,
		SSID:            ssid,
		Security:        c.Security,
		Password:        c.Password,
		VLAN:            c.VLAN,
		Band:            c.Band,
		Hidden:          c.Hidden,
		PMFMode:         c.PMFMode,
		WPA3Support:     c.WPA3Support,
		MACFilter:       c.MACFilter,
		MACFilterPolicy: c.MACFilterPolicy,
	}

	resp, err := ctx.Client.CreateWLAN(c.SiteID, req)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	fmt.Printf("WLAN created successfully:\n")
	wlan := resp.Data
	wlanData := []output.WLANData{{
		ID:       wlan.ID,
		Name:     wlan.Name,
		SSID:     wlan.SSID,
		Security: wlan.Security,
		Enabled:  wlan.Enabled,
		Hidden:   wlan.Hidden,
		VLAN:     wlan.VLAN,
		Band:     wlan.Band,
	}}
	formatter.PrintWLANsTable(wlanData)
	return nil
}

// UpdateWLANCmd handles updating a WLAN
type UpdateWLANCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	WLANID   string `arg:"" help:"WLAN ID to update"`
	Name     string `help:"New WLAN name"`
	Security string `help:"New security type"`
	Password string `help:"New password"`
	Enabled  bool   `help:"Enable/disable WLAN"`
	Hidden   bool   `help:"Hide/show SSID"`
	VLAN     int    `help:"VLAN ID"`
	Band     string `help:"Band (2g, 5g, both)"`
	PMFMode  string `help:"PMF mode"`
}

func (c *UpdateWLANCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.WLANID == "" {
		return &api.ValidationError{Message: "WLAN ID is required"}
	}

	req := api.UpdateWLANRequest{}
	if c.Name != "" {
		req.Name = c.Name
	}
	if c.Security != "" {
		req.Security = c.Security
	}
	if c.Password != "" {
		req.Password = c.Password
	}
	if c.Band != "" {
		req.Band = c.Band
	}
	if c.PMFMode != "" {
		req.PMFMode = c.PMFMode
	}

	// Boolean fields are always sent when the flag is provided
	enabled := c.Enabled
	req.Enabled = &enabled
	hidden := c.Hidden
	req.Hidden = &hidden
	req.VLAN = c.VLAN

	resp, err := ctx.Client.UpdateWLAN(c.SiteID, c.WLANID, req)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	fmt.Printf("WLAN updated successfully:\n")
	wlan := resp.Data
	wlanData := []output.WLANData{{
		ID:       wlan.ID,
		Name:     wlan.Name,
		SSID:     wlan.SSID,
		Security: wlan.Security,
		Enabled:  wlan.Enabled,
		Hidden:   wlan.Hidden,
		VLAN:     wlan.VLAN,
		Band:     wlan.Band,
	}}
	formatter.PrintWLANsTable(wlanData)
	return nil
}

// DeleteWLANCmd handles deleting a WLAN
type DeleteWLANCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	WLANID string `arg:"" help:"WLAN ID to delete"`
	Force  bool   `help:"Skip confirmation"`
}

func (c *DeleteWLANCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.WLANID == "" {
		return &api.ValidationError{Message: "WLAN ID is required"}
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

	if err := ctx.Client.DeleteWLAN(c.SiteID, c.WLANID); err != nil {
		return err
	}

	fmt.Printf("WLAN %s deleted successfully\n", c.WLANID)
	return nil
}
