package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ========== WLANS ==========

// ListWLANs retrieves WLANs for a site
func (c *Client) ListWLANs(siteID string, pageSize int, nextToken string) (*WLANsResponse, error) {
	endpoint := fmt.Sprintf("/v1/sites/%s/wlans", siteID)
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

	var result WLANsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse wlans response: %w", err)
	}

	return &result, nil
}

// CreateWLAN creates a new WLAN
func (c *Client) CreateWLAN(siteID string, req CreateWLANRequest) (*WLANResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doPost(fmt.Sprintf("/v1/sites/%s/wlans", siteID), body)
	if err != nil {
		return nil, err
	}

	var result WLANResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse create wlan response: %w", err)
	}

	return &result, nil
}

// GetWLAN retrieves a specific WLAN
func (c *Client) GetWLAN(siteID, wlanID string) (*WLANResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/sites/%s/wlans/%s", siteID, wlanID))
	if err != nil {
		return nil, err
	}

	var result WLANResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse wlan response: %w", err)
	}

	return &result, nil
}

// UpdateWLAN updates a WLAN
func (c *Client) UpdateWLAN(siteID, wlanID string, req UpdateWLANRequest) (*WLANResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doPut(fmt.Sprintf("/v1/sites/%s/wlans/%s", siteID, wlanID), body)
	if err != nil {
		return nil, err
	}

	var result WLANResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse update wlan response: %w", err)
	}

	return &result, nil
}

// DeleteWLAN deletes a WLAN
func (c *Client) DeleteWLAN(siteID, wlanID string) error {
	_, err := c.doDelete(fmt.Sprintf("/v1/sites/%s/wlans/%s", siteID, wlanID))
	return err
}
