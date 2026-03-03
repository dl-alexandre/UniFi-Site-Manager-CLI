package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ========== EVENTS ==========

// ListEvents retrieves events for a site
func (c *Client) ListEvents(siteID string, pageSize int, nextToken string) (*EventsResponse, error) {
	var endpoint string
	if siteID != "" {
		endpoint = fmt.Sprintf("/v1/sites/%s/events", siteID)
	} else {
		endpoint = "/v1/events"
	}

	params := []string{}
	if pageSize > 0 {
		params = append(params, "pageSize="+strconv.Itoa(pageSize))
	}
	if nextToken != "" {
		params = append(params, "nextToken="+nextToken)
	}

	if len(params) > 0 {
		endpoint = endpoint + "?" + strings.Join(params, "&")
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result EventsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse events response: %w", err)
	}

	return &result, nil
}

// ========== NETWORKS ==========

// ListNetworks retrieves networks for a site
func (c *Client) ListNetworks(siteID string, pageSize int, nextToken string) (*NetworksResponse, error) {
	endpoint := fmt.Sprintf("/v1/sites/%s/networks", siteID)
	if pageSize > 0 {
		endpoint = endpoint + "?pageSize=" + strconv.Itoa(pageSize)
	}
	if nextToken != "" {
		sep := "?"
		if pageSize > 0 {
			sep = "&"
		}
		endpoint = endpoint + sep + "nextToken=" + nextToken
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result NetworksResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse networks response: %w", err)
	}

	return &result, nil
}

// EnableNetwork enables a network by ID
func (c *Client) EnableNetwork(siteID, networkID string) error {
	endpoint := fmt.Sprintf("/v1/sites/%s/networks/%s/enable", siteID, networkID)

	resp, err := c.doPost(endpoint, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to enable network: %d", resp.StatusCode())
	}

	return nil
}

// DisableNetwork disables a network by ID
func (c *Client) DisableNetwork(siteID, networkID string) error {
	endpoint := fmt.Sprintf("/v1/sites/%s/networks/%s/disable", siteID, networkID)

	resp, err := c.doPost(endpoint, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to disable network: %d", resp.StatusCode())
	}

	return nil
}

// ========== USER / AUTH ==========

// Whoami retrieves information about the authenticated user
func (c *Client) Whoami() (*WhoamiResponse, error) {
	resp, err := c.doGet("/v1/whoami")
	if err != nil {
		return nil, err
	}

	var result WhoamiResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse whoami response: %w", err)
	}

	return &result, nil
}
