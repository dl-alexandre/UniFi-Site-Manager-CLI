package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

// ========== SITES ==========

// ListSites retrieves a list of all sites
func (c *Client) ListSites(pageSize int, nextToken string) (*SitesResponse, error) {
	endpoint := "/v1/sites"
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

	var result SitesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse sites response: %w", err)
	}

	return &result, nil
}

// GetSite retrieves a specific site by ID
func (c *Client) GetSite(siteID string) (*SiteResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/sites/%s", siteID))
	if err != nil {
		return nil, err
	}

	var result SiteResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse site response: %w", err)
	}

	return &result, nil
}

// CreateSite creates a new site
func (c *Client) CreateSite(req CreateSiteRequest) (*SiteResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doPost("/v1/sites", body)
	if err != nil {
		return nil, err
	}

	var result SiteResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse create site response: %w", err)
	}

	return &result, nil
}

// UpdateSite updates an existing site
func (c *Client) UpdateSite(siteID string, req UpdateSiteRequest) (*SiteResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doPut(fmt.Sprintf("/v1/sites/%s", siteID), body)
	if err != nil {
		return nil, err
	}

	var result SiteResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse update site response: %w", err)
	}

	return &result, nil
}

// DeleteSite deletes a site by ID
func (c *Client) DeleteSite(siteID string) error {
	_, err := c.doDelete(fmt.Sprintf("/v1/sites/%s", siteID))
	return err
}

// GetSiteHealth retrieves health information for a site
func (c *Client) GetSiteHealth(siteID string) (*HealthResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/sites/%s/health", siteID))
	if err != nil {
		return nil, err
	}

	var result HealthResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse site health response: %w", err)
	}

	return &result, nil
}

// GetSiteStats retrieves statistics for a site
func (c *Client) GetSiteStats(siteID string, period string) (*PerformanceResponse, error) {
	endpoint := fmt.Sprintf("/v1/sites/%s/stats", siteID)
	if period != "" {
		endpoint = endpoint + "?period=" + period
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result PerformanceResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse site stats response: %w", err)
	}

	return &result, nil
}

// EnableDebug enables verbose API logging with credential redaction
// This provides the same interface as LocalClient for debugging Cloud API issues
func (c *Client) EnableDebug() {
	// Cloud client uses different HTTP client (resty in Client vs raw in LocalClient)
	// For Cloud client, we enable resty's built-in debug but with redaction
	c.httpClient.SetDebug(true)

	// Add middleware to redact sensitive info
	c.httpClient.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
		fmt.Println("[DEBUG] === CLOUD API REQUEST ===")
		fmt.Printf("[DEBUG] Method: %s URL: %s\n", req.Method, req.URL)

		// Print headers, redacting API key
		for k, v := range req.Header {
			kLower := strings.ToLower(k)
			if kLower == "x-api-key" || kLower == "authorization" {
				fmt.Printf("[DEBUG] Header: %s: [REDACTED]\n", k)
			} else {
				fmt.Printf("[DEBUG] Header: %s: %v\n", k, v)
			}
		}
		fmt.Println("[DEBUG] =================")
		return nil
	})

	c.httpClient.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		fmt.Println("[DEBUG] === CLOUD API RESPONSE ===")
		fmt.Printf("[DEBUG] Status: %d %s\n", resp.StatusCode(), resp.Status())

		if resp.String() != "" {
			body := resp.String()
			if len(body) > 2000 {
				body = body[:2000] + "... [truncated]"
			}
			fmt.Printf("[DEBUG] Raw Payload: %s\n", body)
		}
		fmt.Println("[DEBUG] ==================")
		return nil
	})
}
