package cli

import (
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// EventsCmd is the parent command for event operations
type EventsCmd struct {
	List ListEventsCmd `cmd:"" help:"List all events"`
}

// ListEventsCmd handles listing events
type ListEventsCmd struct {
	SiteID   string `help:"Filter events by site ID"`
	PageSize int    `help:"Number of events per page (0 = fetch all)" default:"50"`
}

func (c *ListEventsCmd) Run(ctx *CLIContext) error {
	var allEvents []api.Event
	nextToken := ""

	for {
		resp, err := ctx.Client.ListEvents(c.SiteID, c.PageSize, nextToken)
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

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
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
