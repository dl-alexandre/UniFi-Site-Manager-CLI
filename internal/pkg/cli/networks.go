package cli

import (
	"fmt"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// NetworksCmd is the parent command for network operations
type NetworksCmd struct {
	List    ListNetworksCmd   `cmd:"" help:"List all networks"`
	Enable  EnableNetworkCmd  `cmd:"" help:"Enable a network"`
	Disable DisableNetworkCmd `cmd:"" help:"Disable a network"`
}

// ListNetworksCmd handles listing networks
type ListNetworksCmd struct {
	SiteID   string `arg:"" help:"Site ID to list networks for"`
	PageSize int    `help:"Number of networks per page (0 = fetch all)" default:"50"`
}

func (c *ListNetworksCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allNetworks []api.Network
	nextToken := ""

	for {
		resp, err := ctx.Client.ListNetworks(c.SiteID, c.PageSize, nextToken)
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

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// EnableNetworkCmd handles enabling a network
type EnableNetworkCmd struct {
	SiteID    string `arg:"" help:"Site ID containing the network"`
	NetworkID string `arg:"" help:"Network ID to enable"`
}

func (c *EnableNetworkCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.NetworkID == "" {
		return &api.ValidationError{Message: "network ID is required"}
	}

	if err := ctx.Client.EnableNetwork(c.SiteID, c.NetworkID); err != nil {
		return err
	}

	fmt.Printf("✓ Network %s enabled successfully\n", c.NetworkID)
	return nil
}

// DisableNetworkCmd handles disabling a network
type DisableNetworkCmd struct {
	SiteID    string `arg:"" help:"Site ID containing the network"`
	NetworkID string `arg:"" help:"Network ID to disable"`
	Force     bool   `help:"Force disable without confirmation"`
}

func (c *DisableNetworkCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.NetworkID == "" {
		return &api.ValidationError{Message: "network ID is required"}
	}

	if !c.Force {
		// Get network details to warn about default network
		resp, err := ctx.Client.ListNetworks(c.SiteID, 100, "")
		if err != nil {
			return err
		}

		var targetNetwork *api.Network
		for i := range resp.Data {
			if resp.Data[i].ID == c.NetworkID {
				targetNetwork = &resp.Data[i]
				break
			}
		}

		if targetNetwork == nil {
			return &api.NotFoundError{Resource: fmt.Sprintf("network %s", c.NetworkID)}
		}

		// Warn if this might be a main network
		if targetNetwork.Purpose == "corporate" || targetNetwork.VLAN == 1 {
			fmt.Printf("⚠ Warning: Disabling network '%s' (VLAN: %d) may disrupt connectivity.\n", targetNetwork.Name, targetNetwork.VLAN)
			fmt.Println("  This network appears to be a main/corporate network.")
		}
	}

	if err := ctx.Client.DisableNetwork(c.SiteID, c.NetworkID); err != nil {
		return err
	}

	fmt.Printf("✓ Network %s disabled successfully\n", c.NetworkID)
	return nil
}
