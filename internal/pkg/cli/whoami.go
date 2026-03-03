package cli

import (
	"fmt"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// WhoamiCmd handles showing authenticated user information
type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(ctx *CLIContext) error {
	// Get connection info for diagnostics
	connInfo := ctx.Client.GetConnectionInfo()

	resp, err := ctx.Client.Whoami()
	if err != nil {
		// Show connection info even if Whoami fails (helps with debugging)
		fmt.Printf("Connection Mode: %s\n", connInfo.Mode)
		fmt.Printf("Endpoint: %s\n", connInfo.Endpoint)
		fmt.Printf("Connected: %v\n\n", connInfo.IsConnected)
		return fmt.Errorf("authentication failed: %w\n\nHint: Use --debug flag to see detailed error information", err)
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		// Include connection info in JSON output
		type WhoamiResult struct {
			Connection interface{} `json:"connection"`
			User       interface{} `json:"user"`
		}
		result := WhoamiResult{
			Connection: connInfo,
			User:       resp.Data,
		}
		return formatter.PrintJSON(result)
	}

	// Table output with connection info header
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("         CONNECTION INFO")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("Mode:       %s\n", connInfo.Mode)
	fmt.Printf("Endpoint:   %s\n", connInfo.Endpoint)
	fmt.Printf("Version:    %s\n", connInfo.Version)
	if connInfo.SiteID != "" {
		fmt.Printf("Site ID:    %s\n", connInfo.SiteID)
	}
	fmt.Printf("Status:     ✓ Connected\n")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

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

	// Show helpful next steps
	fmt.Println()
	fmt.Println("Next steps:")
	if connInfo.Mode == "local" {
		fmt.Println("  • Try: usm devices list")
		fmt.Println("  • Try: usm clients list")
		fmt.Println("  • Try: usm wlans list")
	} else {
		fmt.Println("  • Try: usm sites list")
		fmt.Println("  • Try: usm hosts list")
	}

	return nil
}
