package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ========== CLIENTS ==========

// ListClients retrieves clients for a site
func (c *Client) ListClients(siteID string, pageSize int, nextToken string, wiredOnly, wirelessOnly bool) (*ClientsResponse, error) {
	endpoint := fmt.Sprintf("/v1/sites/%s/clients", siteID)
	params := []string{}

	if pageSize > 0 {
		params = append(params, "pageSize="+strconv.Itoa(pageSize))
	}
	if nextToken != "" {
		params = append(params, "nextToken="+nextToken)
	}
	if wiredOnly {
		params = append(params, "wired=true")
	}
	if wirelessOnly {
		params = append(params, "wireless=true")
	}

	if len(params) > 0 {
		endpoint = endpoint + "?" + strings.Join(params, "&")
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result ClientsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse clients response: %w", err)
	}

	return &result, nil
}

// GetClientStats retrieves statistics for a specific client
func (c *Client) GetClientStats(siteID, macAddress string) (*SingleResponse[ClientStats], error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/sites/%s/clients/%s/stats", siteID, macAddress))
	if err != nil {
		return nil, err
	}

	var result SingleResponse[ClientStats]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse client stats response: %w", err)
	}

	return &result, nil
}

// BlockClient blocks or unblocks a client
func (c *Client) BlockClient(siteID, macAddress string, block bool) error {
	reqBody := BlockClientRequest{MACAddress: macAddress, Block: block}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.doPost(fmt.Sprintf("/v1/sites/%s/clients/block", siteID), body)
	return err
}
