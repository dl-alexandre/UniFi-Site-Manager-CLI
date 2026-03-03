package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ========== HOSTS ==========

// ListHosts retrieves all hosts/consoles
func (c *Client) ListHosts(pageSize int, nextToken string) (*HostsResponse, error) {
	endpoint := "/v1/hosts"
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

	var result HostsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse hosts response: %w", err)
	}

	return &result, nil
}

// GetHost retrieves a specific host by ID
func (c *Client) GetHost(hostID string) (*HostResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/hosts/%s", hostID))
	if err != nil {
		return nil, err
	}

	var result HostResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse host response: %w", err)
	}

	return &result, nil
}

// GetHostHealth retrieves health information for a host
func (c *Client) GetHostHealth(hostID string) (*HealthResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/hosts/%s/health", hostID))
	if err != nil {
		return nil, err
	}

	var result HealthResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse host health response: %w", err)
	}

	return &result, nil
}

// GetHostStats retrieves statistics for a host
func (c *Client) GetHostStats(hostID string, period string) (*PerformanceResponse, error) {
	endpoint := fmt.Sprintf("/v1/hosts/%s/stats", hostID)
	if period != "" {
		endpoint = endpoint + "?period=" + period
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result PerformanceResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse host stats response: %w", err)
	}

	return &result, nil
}

// RestartHost restarts a host/console
func (c *Client) RestartHost(hostID string) error {
	_, err := c.doPost(fmt.Sprintf("/v1/hosts/%s/restart", hostID), nil)
	return err
}
