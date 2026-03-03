package cli

import (
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// NetworksCmd is the parent command for network operations
type NetworksCmd struct {
	List ListNetworksCmd `cmd:"" help:"List all networks"`
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
