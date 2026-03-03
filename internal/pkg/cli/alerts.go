package cli

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// AlertsCmd is the parent command for alert operations
type AlertsCmd struct {
	List    ListAlertsCmd   `cmd:"" help:"List all alerts"`
	Ack     AckAlertCmd     `cmd:"" help:"Acknowledge an alert"`
	Archive ArchiveAlertCmd `cmd:"" help:"Archive an alert"`
}

// ListAlertsCmd handles listing alerts
type ListAlertsCmd struct {
	SiteID   string `help:"Filter alerts by site ID (optional)"`
	PageSize int    `help:"Number of alerts per page (0 = fetch all)" default:"50"`
	Archived bool   `help:"Show archived alerts"`
	Search   string `help:"Filter alerts by message content"`
}

func (c *ListAlertsCmd) Run(ctx *CLIContext) error {
	var allAlerts []api.Alert
	nextToken := ""

	for {
		resp, err := ctx.Client.ListAlerts(c.SiteID, c.PageSize, nextToken, c.Archived)
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

	if c.Search != "" {
		filtered := make([]api.Alert, 0)
		searchLower := strings.ToLower(c.Search)
		for _, alert := range allAlerts {
			if strings.Contains(strings.ToLower(alert.Message), searchLower) {
				filtered = append(filtered, alert)
			}
		}
		allAlerts = filtered
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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

// AckAlertCmd handles acknowledging an alert
type AckAlertCmd struct {
	SiteID  string `arg:"" help:"Site ID"`
	AlertID string `arg:"" help:"Alert ID to acknowledge"`
}

func (c *AckAlertCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.AlertID == "" {
		return &api.ValidationError{Message: "alert ID is required"}
	}

	if err := ctx.Client.AcknowledgeAlert(c.SiteID, c.AlertID); err != nil {
		return err
	}

	fmt.Printf("Alert %s acknowledged successfully\n", c.AlertID)
	return nil
}

// ArchiveAlertCmd handles archiving an alert
type ArchiveAlertCmd struct {
	SiteID  string `arg:"" help:"Site ID"`
	AlertID string `arg:"" help:"Alert ID to archive"`
}

func (c *ArchiveAlertCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.AlertID == "" {
		return &api.ValidationError{Message: "alert ID is required"}
	}

	if err := ctx.Client.ArchiveAlert(c.SiteID, c.AlertID); err != nil {
		return err
	}

	fmt.Printf("Alert %s archived successfully\n", c.AlertID)
	return nil
}
