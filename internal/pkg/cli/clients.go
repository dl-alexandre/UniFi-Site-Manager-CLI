package cli

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// ClientsCmd is the parent command for client operations
type ClientsCmd struct {
	List    ListClientsCmd   `cmd:"" help:"List all clients"`
	Stats   ClientStatsCmd   `cmd:"" help:"Get client statistics"`
	Block   BlockClientCmd   `cmd:"" help:"Block a client"`
	Unblock UnblockClientCmd `cmd:"" help:"Unblock a client"`
}

// ListClientsCmd handles listing clients
type ListClientsCmd struct {
	SiteID       string `arg:"" help:"Site ID to list clients for"`
	PageSize     int    `help:"Number of clients per page (0 = fetch all)" default:"50"`
	WiredOnly    bool   `help:"Show only wired clients"`
	WirelessOnly bool   `help:"Show only wireless clients"`
	Search       string `help:"Filter clients by name, hostname, or MAC"`
}

func (c *ListClientsCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allClients []api.NetworkClient
	nextToken := ""

	for {
		resp, err := ctx.Client.ListClients(c.SiteID, c.PageSize, nextToken, c.WiredOnly, c.WirelessOnly)
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
			if strings.Contains(strings.ToLower(client.Name), searchLower) ||
				strings.Contains(strings.ToLower(client.Hostname), searchLower) ||
				strings.Contains(strings.ToLower(client.MACAddress), searchLower) {
				filtered = append(filtered, client)
			}
		}
		allClients = filtered
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(allClients)
	}

	clientData := make([]output.ClientData, len(allClients))
	for i, client := range allClients {
		connType := "Wireless"
		if client.IsWired {
			connType = "Wired"
		}
		clientData[i] = output.ClientData{
			ID:             client.ID,
			MACAddress:     client.MACAddress,
			IPAddress:      client.IPAddress,
			Hostname:       client.Hostname,
			Name:           client.Name,
			ConnectionType: connType,
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

// ClientStatsCmd handles getting client statistics
type ClientStatsCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Client MAC address"`
}

func (c *ClientStatsCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.MACAddress == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	resp, err := ctx.Client.GetClientStats(c.SiteID, c.MACAddress)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	stats := resp.Data
	fmt.Printf("Client Statistics for %s:\n", c.MACAddress)
	fmt.Printf("  RX Bytes: %d\n", stats.RxBytes)
	fmt.Printf("  TX Bytes: %d\n", stats.TxBytes)
	fmt.Printf("  RX Packets: %d\n", stats.RxPackets)
	fmt.Printf("  TX Packets: %d\n", stats.TxPackets)
	fmt.Printf("  Signal Avg: %d\n", stats.SignalAvg)
	fmt.Printf("  Satisfaction: %.2f\n", stats.Satisfaction)
	return nil
}

// BlockClientCmd handles blocking a client
type BlockClientCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Client MAC address to block"`
}

func (c *BlockClientCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.MACAddress == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	if err := ctx.Client.BlockClient(c.SiteID, c.MACAddress, true); err != nil {
		return err
	}

	fmt.Printf("Client %s blocked successfully\n", c.MACAddress)
	return nil
}

// UnblockClientCmd handles unblocking a client
type UnblockClientCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	MACAddress string `arg:"" help:"Client MAC address to unblock"`
}

func (c *UnblockClientCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.MACAddress == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	if err := ctx.Client.BlockClient(c.SiteID, c.MACAddress, false); err != nil {
		return err
	}

	fmt.Printf("Client %s unblocked successfully\n", c.MACAddress)
	return nil
}
