package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ========== ALERTS ==========

// ListAlerts retrieves alerts for a site
func (c *Client) ListAlerts(siteID string, pageSize int, nextToken string, archived bool) (*AlertsResponse, error) {
	var endpoint string
	if siteID != "" {
		endpoint = fmt.Sprintf("/v1/sites/%s/alerts", siteID)
	} else {
		endpoint = "/v1/alerts"
	}

	params := []string{}
	if pageSize > 0 {
		params = append(params, "pageSize="+strconv.Itoa(pageSize))
	}
	if nextToken != "" {
		params = append(params, "nextToken="+nextToken)
	}
	if archived {
		params = append(params, "archived=true")
	}

	if len(params) > 0 {
		endpoint = endpoint + "?" + strings.Join(params, "&")
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result AlertsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse alerts response: %w", err)
	}

	return &result, nil
}

// AcknowledgeAlert acknowledges an alert
func (c *Client) AcknowledgeAlert(siteID, alertID string) error {
	reqBody := AcknowledgeAlertRequest{AlertID: alertID}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.doPost(fmt.Sprintf("/v1/sites/%s/alerts/%s/ack", siteID, alertID), body)
	return err
}

// ArchiveAlert archives an alert
func (c *Client) ArchiveAlert(siteID, alertID string) error {
	reqBody := ArchiveAlertRequest{AlertID: alertID}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.doPost(fmt.Sprintf("/v1/sites/%s/alerts/%s/archive", siteID, alertID), body)
	return err
}
