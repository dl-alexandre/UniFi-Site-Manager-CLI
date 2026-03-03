// Package api provides HTTP client for UniFi Site Manager API
package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultBaseURL = "https://api.ui.com"
	apiVersion     = "v1"
)

// Client wraps the HTTP client for UniFi Site Manager API
type Client struct {
	httpClient    *resty.Client
	baseURL       string
	apiKey        string
	timeout       time.Duration
	verbose       bool
	debug         bool
	maxRetryDelay time.Duration
}

// ClientOptions contains configuration options for the client
type ClientOptions struct {
	BaseURL       string
	APIKey        string
	Timeout       int // seconds
	Verbose       bool
	Debug         bool
	MaxRetryDelay time.Duration
}

// NewClient creates a new API client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.APIKey == "" {
		return nil, &AuthError{Message: "API key is required"}
	}

	client := resty.New()

	timeout := time.Duration(opts.Timeout) * time.Second
	if opts.Timeout <= 0 {
		timeout = 30 * time.Second
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	client.SetTimeout(timeout)
	client.SetBaseURL(baseURL)
	client.SetHeader("X-API-Key", opts.APIKey)
	client.SetHeader("Accept", "application/json")
	client.SetHeader("Content-Type", "application/json")

	if opts.Debug {
		client.SetDebug(true)
	}

	return &Client{
		httpClient:    client,
		baseURL:       baseURL,
		apiKey:        opts.APIKey,
		timeout:       timeout,
		verbose:       opts.Verbose,
		debug:         opts.Debug,
		maxRetryDelay: opts.MaxRetryDelay,
	}, nil
}

// doGet performs a GET request with retry logic
func (c *Client) doGet(endpoint string) (*resty.Response, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.httpClient.R().Get(endpoint)

		if err != nil {
			lastErr = &NetworkError{Message: err.Error()}

			if attempt < maxRetries-1 && c.shouldRetry(err) {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}

			return nil, lastErr
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return resp, nil
		case http.StatusUnauthorized:
			return nil, &AuthError{Message: "invalid API key"}
		case http.StatusForbidden:
			return nil, &PermissionError{Message: "permission denied"}
		case http.StatusNotFound:
			return nil, &NotFoundError{Resource: endpoint}
		case http.StatusTooManyRequests:
			retryAfter := c.parseRetryAfter(resp)
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(retryAfter) * time.Second
				if sleepDuration <= 0 {
					sleepDuration = c.calculateBackoff(attempt)
				}
				if c.maxRetryDelay > 0 && sleepDuration > c.maxRetryDelay {
					sleepDuration = c.maxRetryDelay
				}
				time.Sleep(sleepDuration)
				continue
			}
			return nil, &RateLimitError{RetryAfter: retryAfter}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			if attempt < maxRetries-1 {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("server error: %d", resp.StatusCode())
		default:
			if resp.StatusCode() >= 400 {
				return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), string(resp.Body()))
			}
			return resp, nil
		}
	}

	return nil, lastErr
}

// doPost performs a POST request with retry logic
func (c *Client) doPost(endpoint string, body []byte) (*resty.Response, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		var resp *resty.Response
		var err error

		if body != nil {
			resp, err = c.httpClient.R().SetBody(body).Post(endpoint)
		} else {
			resp, err = c.httpClient.R().Post(endpoint)
		}

		if err != nil {
			lastErr = &NetworkError{Message: err.Error()}

			if attempt < maxRetries-1 && c.shouldRetry(err) {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}

			return nil, lastErr
		}

		switch resp.StatusCode() {
		case http.StatusOK, http.StatusCreated:
			return resp, nil
		case http.StatusNoContent:
			return resp, nil
		case http.StatusBadRequest:
			var apiErr APIResponse
			if err := json.Unmarshal(resp.Body(), &apiErr); err == nil && apiErr.Message != "" {
				return nil, &ValidationError{Message: apiErr.Message}
			}
			return nil, &ValidationError{Message: "invalid request"}
		case http.StatusUnauthorized:
			return nil, &AuthError{Message: "invalid API key"}
		case http.StatusForbidden:
			return nil, &PermissionError{Message: "permission denied"}
		case http.StatusNotFound:
			return nil, &NotFoundError{Resource: endpoint}
		case http.StatusTooManyRequests:
			retryAfter := c.parseRetryAfter(resp)
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(retryAfter) * time.Second
				if sleepDuration <= 0 {
					sleepDuration = c.calculateBackoff(attempt)
				}
				if c.maxRetryDelay > 0 && sleepDuration > c.maxRetryDelay {
					sleepDuration = c.maxRetryDelay
				}
				time.Sleep(sleepDuration)
				continue
			}
			return nil, &RateLimitError{RetryAfter: retryAfter}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			if attempt < maxRetries-1 {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("server error: %d", resp.StatusCode())
		default:
			if resp.StatusCode() >= 400 {
				return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), string(resp.Body()))
			}
			return resp, nil
		}
	}

	return nil, lastErr
}

// doPut performs a PUT request with retry logic
func (c *Client) doPut(endpoint string, body []byte) (*resty.Response, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		var resp *resty.Response
		var err error

		if body != nil {
			resp, err = c.httpClient.R().SetBody(body).Put(endpoint)
		} else {
			resp, err = c.httpClient.R().Put(endpoint)
		}

		if err != nil {
			lastErr = &NetworkError{Message: err.Error()}

			if attempt < maxRetries-1 && c.shouldRetry(err) {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}

			return nil, lastErr
		}

		switch resp.StatusCode() {
		case http.StatusOK, http.StatusCreated:
			return resp, nil
		case http.StatusBadRequest:
			var apiErr APIResponse
			if err := json.Unmarshal(resp.Body(), &apiErr); err == nil && apiErr.Message != "" {
				return nil, &ValidationError{Message: apiErr.Message}
			}
			return nil, &ValidationError{Message: "invalid request"}
		case http.StatusUnauthorized:
			return nil, &AuthError{Message: "invalid API key"}
		case http.StatusForbidden:
			return nil, &PermissionError{Message: "permission denied"}
		case http.StatusNotFound:
			return nil, &NotFoundError{Resource: endpoint}
		case http.StatusTooManyRequests:
			retryAfter := c.parseRetryAfter(resp)
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(retryAfter) * time.Second
				if sleepDuration <= 0 {
					sleepDuration = c.calculateBackoff(attempt)
				}
				if c.maxRetryDelay > 0 && sleepDuration > c.maxRetryDelay {
					sleepDuration = c.maxRetryDelay
				}
				time.Sleep(sleepDuration)
				continue
			}
			return nil, &RateLimitError{RetryAfter: retryAfter}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			if attempt < maxRetries-1 {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("server error: %d", resp.StatusCode())
		default:
			if resp.StatusCode() >= 400 {
				return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), string(resp.Body()))
			}
			return resp, nil
		}
	}

	return nil, lastErr
}

// doDelete performs a DELETE request with retry logic
func (c *Client) doDelete(endpoint string) (*resty.Response, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.httpClient.R().Delete(endpoint)

		if err != nil {
			lastErr = &NetworkError{Message: err.Error()}

			if attempt < maxRetries-1 && c.shouldRetry(err) {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}

			return nil, lastErr
		}

		switch resp.StatusCode() {
		case http.StatusOK, http.StatusNoContent:
			return resp, nil
		case http.StatusUnauthorized:
			return nil, &AuthError{Message: "invalid API key"}
		case http.StatusForbidden:
			return nil, &PermissionError{Message: "permission denied"}
		case http.StatusNotFound:
			return nil, &NotFoundError{Resource: endpoint}
		case http.StatusTooManyRequests:
			retryAfter := c.parseRetryAfter(resp)
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(retryAfter) * time.Second
				if sleepDuration <= 0 {
					sleepDuration = c.calculateBackoff(attempt)
				}
				if c.maxRetryDelay > 0 && sleepDuration > c.maxRetryDelay {
					sleepDuration = c.maxRetryDelay
				}
				time.Sleep(sleepDuration)
				continue
			}
			return nil, &RateLimitError{RetryAfter: retryAfter}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			if attempt < maxRetries-1 {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("server error: %d", resp.StatusCode())
		default:
			if resp.StatusCode() >= 400 {
				return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), string(resp.Body()))
			}
			return resp, nil
		}
	}

	return nil, lastErr
}

func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "temporary")
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}
	baseDelay := time.Duration(1<<(attempt-1)) * time.Second
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return baseDelay + jitter
}

func (c *Client) parseRetryAfter(resp *resty.Response) int {
	retryAfter := resp.Header().Get("Retry-After")
	if retryAfter == "" {
		return 0
	}
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		return 0
	}
	return seconds
}

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

// ========== DEVICES ==========

// ListDevices retrieves devices for a site
func (c *Client) ListDevices(siteID string, pageSize int, nextToken string) (*DevicesResponse, error) {
	endpoint := fmt.Sprintf("/v1/sites/%s/devices", siteID)
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

	var result DevicesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse devices response: %w", err)
	}

	return &result, nil
}

// GetDevice retrieves a specific device by ID
func (c *Client) GetDevice(siteID, deviceID string) (*DeviceResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/sites/%s/devices/%s", siteID, deviceID))
	if err != nil {
		return nil, err
	}

	var result DeviceResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse device response: %w", err)
	}

	return &result, nil
}

// RestartDevice restarts a device
func (c *Client) RestartDevice(siteID, deviceID string) error {
	body := []byte(fmt.Sprintf(`{"deviceId":"%s"}`, deviceID))
	_, err := c.doPost(fmt.Sprintf("/v1/sites/%s/devices/%s/restart", siteID, deviceID), body)
	return err
}

// UpgradeDevice upgrades device firmware
func (c *Client) UpgradeDevice(siteID, deviceID string) error {
	body := []byte(fmt.Sprintf(`{"deviceId":"%s"}`, deviceID))
	_, err := c.doPost(fmt.Sprintf("/v1/sites/%s/devices/%s/upgrade", siteID, deviceID), body)
	return err
}

// AdoptDevice adopts a new device
func (c *Client) AdoptDevice(siteID string, macAddress string) error {
	reqBody := AdoptDeviceRequest{MACAddress: macAddress}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.doPost(fmt.Sprintf("/v1/sites/%s/devices/adopt", siteID), body)
	return err
}

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

// ========== SITE HEALTH & STATS ==========

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
