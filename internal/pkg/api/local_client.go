package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/go-resty/resty/v2"
)

// LocalClient provides HTTP client for UniFi OS local controllers (UDM, UDM-Pro, UDR, etc.)
// Unlike CloudClient which uses API keys, LocalClient uses session-based authentication
// with cookies and CSRF tokens
type LocalClient struct {
	httpClient *resty.Client
	baseURL    string
	csrfToken  string
	username   string
}

// Ensure LocalClient implements SiteManager interface
var _ SiteManager = (*LocalClient)(nil)

// LocalClientOptions contains configuration for connecting to a local UniFi controller
type LocalClientOptions struct {
	Host     string // IP address or hostname of the controller
	Username string
	Password string
	// AllowInsecure allows connections to controllers with self-signed certificates
	// This is the default for local controllers
	AllowInsecure bool
}

// NewLocalClient creates a new client for local UniFi OS controller
func NewLocalClient(opts LocalClientOptions) (*LocalClient, error) {
	if opts.Host == "" {
		return nil, &ValidationError{Message: "controller host is required"}
	}
	if opts.Username == "" || opts.Password == "" {
		return nil, &ValidationError{Message: "username and password are required"}
	}

	// Initialize cookie jar for session management
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Ensure host has protocol
	host := opts.Host
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	// Setup Resty with cookie jar and TLS configuration
	// Local controllers use self-signed certs by default
	httpClient := resty.New().
		SetBaseURL(host).
		SetCookieJar(jar).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	client := &LocalClient{
		httpClient: httpClient,
		baseURL:    host,
		username:   opts.Username,
	}

	// Setup CSRF token injection middleware
	// This automatically adds X-CSRF-Token to all mutating requests (POST, PUT, DELETE)
	client.httpClient.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		// Only inject CSRF token for mutating operations
		method := req.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete {
			if client.csrfToken != "" {
				req.SetHeader("X-CSRF-Token", client.csrfToken)
			}
		}
		return nil
	})

	// Perform authentication
	if err := client.login(opts.Username, opts.Password); err != nil {
		return nil, err
	}

	return client, nil
}

// login authenticates with the local controller and extracts the CSRF token
func (c *LocalClient) login(username, password string) error {
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post("/api/auth/login")

	if err != nil {
		return &NetworkError{Message: fmt.Sprintf("login failed: %v", err)}
	}

	if resp.StatusCode() != http.StatusOK {
		return &AuthError{Message: fmt.Sprintf("authentication failed: %d", resp.StatusCode())}
	}

	// Extract CSRF token from response headers
	// UniFi OS returns this in X-CSRF-Token header after successful login
	csrfToken := resp.Header().Get("X-CSRF-Token")
	if csrfToken == "" {
		// Try alternative header names that UniFi might use
		csrfToken = resp.Header().Get("X-Csrf-Token")
		if csrfToken == "" {
			csrfToken = resp.Header().Get("x-csrf-token")
		}
	}

	if csrfToken == "" {
		// Some versions might include CSRF token in response body
		var responseBody map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &responseBody); err == nil {
			if csrf, ok := responseBody["csrfToken"].(string); ok && csrf != "" {
				csrfToken = csrf
			}
		}
	}

	if csrfToken == "" {
		// Log the available headers for debugging
		return &AuthError{Message: "authentication succeeded but CSRF token was not found in response"}
	}

	c.csrfToken = csrfToken
	return nil
}

// buildURL constructs the full URL for local controller endpoints
// Local controllers use proxy paths for different applications
func (c *LocalClient) buildURL(endpoint string) string {
	// Network application endpoints are prefixed with /proxy/network
	// This routes through the UniFi OS proxy to the Network application
	if strings.HasPrefix(endpoint, "/api/") || strings.HasPrefix(endpoint, "/v1/") {
		// Map cloud API endpoints to local proxy paths
		endpoint = c.mapCloudEndpointToLocal(endpoint)
	}
	return endpoint
}

// mapCloudEndpointToLocal translates cloud API endpoints to local controller proxy paths
func (c *LocalClient) mapCloudEndpointToLocal(cloudEndpoint string) string {
	// Cloud API: /v1/sites/{siteId}/devices
	// Local API: /proxy/network/api/s/{siteId}/stat/device

	// Handle site-specific endpoints
	if strings.Contains(cloudEndpoint, "/sites/") {
		parts := strings.Split(cloudEndpoint, "/")
		if len(parts) >= 3 {
			siteID := parts[3]

			// Map based on remaining path
			remaining := strings.Join(parts[4:], "/")
			switch remaining {
			case "":
				// /v1/sites/{siteId} -> /proxy/network/api/s/{siteId}
				return fmt.Sprintf("/proxy/network/api/s/%s", siteID)
			case "devices":
				return fmt.Sprintf("/proxy/network/api/s/%s/stat/device", siteID)
			case "wlans":
				return fmt.Sprintf("/proxy/network/api/s/%s/rest/wlanconf", siteID)
			case "clients":
				return fmt.Sprintf("/proxy/network/api/s/%s/stat/sta", siteID)
			case "health":
				return fmt.Sprintf("/proxy/network/api/s/%s/stat/health", siteID)
			default:
				// Device-specific operations
				if strings.HasPrefix(remaining, "devices/") {
					deviceParts := strings.Split(remaining, "/")
					if len(deviceParts) >= 2 {
						deviceID := deviceParts[1]
						if len(deviceParts) >= 3 && deviceParts[2] == "restart" {
							// Device restart
							return fmt.Sprintf("/proxy/network/api/s/%s/cmd/devmgr", siteID)
						}
						return fmt.Sprintf("/proxy/network/api/s/%s/stat/device/%s", siteID, deviceID)
					}
				}
				// Fallback: append to proxy path
				return fmt.Sprintf("/proxy/network/api/s/%s/%s", siteID, remaining)
			}
		}
	}

	// Global endpoints
	switch cloudEndpoint {
	case "/v1/sites":
		return "/proxy/network/api/self/sites"
	case "/v1/whoami":
		return "/api/auth/user-info"
	}

	// Default: return original with proxy prefix if not already prefixed
	if !strings.HasPrefix(cloudEndpoint, "/proxy/") {
		return "/proxy/network" + cloudEndpoint
	}

	return cloudEndpoint
}

// ========== SiteManager Interface Implementation ==========

// resolveSite maps an empty CLI site ID to the local controller's default site
// Local controllers (UDM/UDR) use "default" as the primary site ID
func (c *LocalClient) resolveSite(siteID string) string {
	if siteID == "" {
		return "default"
	}
	return siteID
}

// EnableDebug enables verbose API logging with credential redaction
// This is critical for debugging Local API issues without exposing passwords
func (c *LocalClient) EnableDebug() {
	// Log outgoing requests
	c.httpClient.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
		fmt.Println("[DEBUG] === REQUEST ===")
		fmt.Printf("[DEBUG] Method: %s URL: %s\n", req.Method, req.URL)

		// Print headers, redacting sensitive ones
		for k, v := range req.Header {
			kLower := strings.ToLower(k)
			if kLower == "x-csrf-token" || kLower == "cookie" || kLower == "authorization" || kLower == "x-api-key" {
				fmt.Printf("[DEBUG] Header: %s: [REDACTED]\n", k)
			} else {
				fmt.Printf("[DEBUG] Header: %s: %v\n", k, v)
			}
		}

		// Print JSON body, but NEVER for login endpoint
		if req.Body != nil && !strings.Contains(req.URL, "/api/auth/login") {
			// Try to get body as string
			var bodyStr string
			switch body := req.Body.(type) {
			case []byte:
				bodyStr = string(body)
			case string:
				bodyStr = body
			default:
				bodyStr = fmt.Sprintf("%v", req.Body)
			}

			// Redact password fields from body if present
			if strings.Contains(strings.ToLower(bodyStr), "password") || strings.Contains(bodyStr, "x_passphrase") {
				fmt.Printf("[DEBUG] Body: [REDACTED - contains credentials]\n")
			} else {
				fmt.Printf("[DEBUG] Body: %s\n", bodyStr)
			}
		}
		fmt.Println("[DEBUG] =================")
		return nil
	})

	// Log incoming responses
	c.httpClient.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		fmt.Println("[DEBUG] === RESPONSE ===")
		fmt.Printf("[DEBUG] Status: %d %s\n", resp.StatusCode(), resp.Status())
		fmt.Printf("[DEBUG] Time: %v\n", resp.Time())

		// Print raw JSON response for debugging struct mapping issues
		if resp.String() != "" {
			// Truncate very long responses
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

// ListSites retrieves a list of all sites from the local controller
// Maps to: GET /proxy/network/api/self/sites
func (c *LocalClient) ListSites(pageSize int, nextToken string) (*SitesResponse, error) {
	endpoint := c.buildURL("/v1/sites")

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result SitesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse sites response: %w", err)
	}

	return &result, nil
}

// GetSite retrieves a specific site by ID
// Maps to: GET /proxy/network/api/s/{siteId}
func (c *LocalClient) GetSite(siteID string) (*SiteResponse, error) {
	endpoint := c.buildURL(fmt.Sprintf("/v1/sites/%s", siteID))

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result SiteResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse site response: %w", err)
	}

	return &result, nil
}

// RestartDevice restarts a device on the local controller
// Maps to: POST /proxy/network/api/s/{siteId}/cmd/devmgr with restart command
func (c *LocalClient) RestartDevice(siteID, deviceID string) error {
	// For local controllers, device restart is done via cmd/devmgr endpoint
	// We need to map this to the correct local endpoint
	cmdEndpoint := fmt.Sprintf("/proxy/network/api/s/%s/cmd/devmgr", siteID)

	payload := map[string]interface{}{
		"cmd": "restart",
		"mac": deviceID, // Local API often uses MAC instead of device ID
	}

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(cmdEndpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return c.handleErrorResponse(resp.StatusCode(), cmdEndpoint, resp.Body())
	}

	return nil
}

// handleErrorResponse converts HTTP error codes to appropriate error types
func (c *LocalClient) handleErrorResponse(statusCode int, endpoint string, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return &AuthError{Message: "authentication required or session expired"}
	case http.StatusForbidden:
		return &PermissionError{Message: "insufficient permissions"}
	case http.StatusNotFound:
		return &NotFoundError{Resource: endpoint}
	case http.StatusTooManyRequests:
		return &RateLimitError{RetryAfter: 60}
	default:
		return fmt.Errorf("API error: %d - %s", statusCode, string(body))
	}
}

// ========== Stub implementations for remaining SiteManager methods ==========
// These will be implemented as needed for full local controller support

func (c *LocalClient) CreateSite(req CreateSiteRequest) (*SiteResponse, error) {
	return nil, &NotImplementedError{Method: "CreateSite"}
}

func (c *LocalClient) UpdateSite(siteID string, req UpdateSiteRequest) (*SiteResponse, error) {
	return nil, &NotImplementedError{Method: "UpdateSite"}
}

func (c *LocalClient) DeleteSite(siteID string) error {
	return &NotImplementedError{Method: "DeleteSite"}
}

func (c *LocalClient) GetSiteHealth(siteID string) (*HealthResponse, error) {
	return nil, &NotImplementedError{Method: "GetSiteHealth"}
}

func (c *LocalClient) GetSiteStats(siteID string, period string) (*PerformanceResponse, error) {
	return nil, &NotImplementedError{Method: "GetSiteStats"}
}

func (c *LocalClient) ListHosts(pageSize int, nextToken string) (*HostsResponse, error) {
	return nil, &NotImplementedError{Method: "ListHosts"}
}

func (c *LocalClient) GetHost(hostID string) (*HostResponse, error) {
	return nil, &NotImplementedError{Method: "GetHost"}
}

func (c *LocalClient) GetHostHealth(hostID string) (*HealthResponse, error) {
	return nil, &NotImplementedError{Method: "GetHostHealth"}
}

func (c *LocalClient) GetHostStats(hostID string, period string) (*PerformanceResponse, error) {
	return nil, &NotImplementedError{Method: "GetHostStats"}
}

func (c *LocalClient) RestartHost(hostID string) error {
	return &NotImplementedError{Method: "RestartHost"}
}

// ListDevices retrieves all devices for a site from the local controller
// Maps to: GET /proxy/network/api/s/{site}/stat/device
// Note: Local API does not support pagination - returns all devices in one request
func (c *LocalClient) ListDevices(siteID string, pageSize int, nextToken string) (*DevicesResponse, error) {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/stat/device", site)

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result DevicesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse devices response: %w", err)
	}

	// Local API returns all devices - apply client-side pagination if requested
	if pageSize > 0 && len(result.Data) > pageSize {
		result.Data = result.Data[:pageSize]
	}

	return &result, nil
}

// GetDevice retrieves a specific device by MAC address from the local controller
// Maps to: GET /proxy/network/api/s/{site}/stat/device/{mac}
// Note: Local API uses MAC address as identifier, not UUID
// The local controller returns a LIST with one item, not a single object
func (c *LocalClient) GetDevice(siteID, deviceID string) (*DeviceResponse, error) {
	site := c.resolveSite(siteID)

	// deviceID for local controllers is the MAC address
	// Format: lowercase, with or without colons (API accepts both)
	macAddress := deviceID

	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/stat/device/%s", site, macAddress)

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	// Local API returns a list response even for single device query
	var listResp DevicesResponse
	if err := json.Unmarshal(resp.Body(), &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse device response: %w", err)
	}

	if len(listResp.Data) == 0 {
		return nil, &NotFoundError{Resource: endpoint}
	}

	// Convert list response to single response
	return &DeviceResponse{
		Data: listResp.Data[0],
	}, nil
}

func (c *LocalClient) UpgradeDevice(siteID, deviceID string) error {
	return &NotImplementedError{Method: "UpgradeDevice"}
}

func (c *LocalClient) AdoptDevice(siteID string, macAddress string) error {
	return &NotImplementedError{Method: "AdoptDevice"}
}

// ListClients retrieves all connected clients for a site from the local controller
// Maps to: GET /proxy/network/api/s/{site}/stat/sta
// Note: "sta" = station (UniFi terminology for connected client)
// Local API does not support pagination - returns all clients in one request
func (c *LocalClient) ListClients(siteID string, pageSize int, nextToken string, wiredOnly, wirelessOnly bool) (*ClientsResponse, error) {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/stat/sta", site)

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result ClientsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse clients response: %w", err)
	}

	// Apply wired/wireless filters client-side
	if wiredOnly || wirelessOnly {
		filtered := make([]NetworkClient, 0)
		for _, client := range result.Data {
			if wiredOnly && client.IsWired {
				filtered = append(filtered, client)
			} else if wirelessOnly && !client.IsWired {
				filtered = append(filtered, client)
			}
		}
		result.Data = filtered
	}

	// Apply pagination client-side if requested
	if pageSize > 0 && len(result.Data) > pageSize {
		result.Data = result.Data[:pageSize]
	}

	return &result, nil
}

func (c *LocalClient) GetClientStats(siteID, macAddress string) (*SingleResponse[ClientStats], error) {
	return nil, &NotImplementedError{Method: "GetClientStats"}
}

func (c *LocalClient) BlockClient(siteID, macAddress string, block bool) error {
	return &NotImplementedError{Method: "BlockClient"}
}

// ListWLANs retrieves all wireless networks for a site from the local controller
// Maps to: GET /proxy/network/api/s/{site}/rest/wlanconf
// Note: Uses /rest/ path for configuration, not /stat/ path
func (c *LocalClient) ListWLANs(siteID string, pageSize int, nextToken string) (*WLANsResponse, error) {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/rest/wlanconf", site)

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result WLANsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse wlans response: %w", err)
	}

	return &result, nil
}

// CreateWLAN creates a new wireless network on the local controller
// Maps to: POST /proxy/network/api/s/{site}/rest/wlanconf
// Note: This is a mutating operation and requires CSRF token (auto-injected)
func (c *LocalClient) CreateWLAN(siteID string, req CreateWLANRequest) (*WLANResponse, error) {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/rest/wlanconf", site)

	// Convert CreateWLANRequest to local API payload
	// Local API uses slightly different field names
	payload := map[string]interface{}{
		"name":     req.Name,
		"essid":    req.SSID,
		"security": req.Security,
	}

	if req.Password != "" {
		payload["x_passphrase"] = req.Password // Local API uses x_passphrase for WPA PSK
	}

	if req.VLAN > 0 {
		payload["vlan"] = req.VLAN
		payload["vlan_enabled"] = true
	}

	if req.Band != "" {
		payload["band"] = req.Band
	}

	payload["hide_ssid"] = req.Hidden
	payload["enabled"] = true // Default to enabled

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result WLANResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse create wlan response: %w", err)
	}

	return &result, nil
}

// GetWLAN retrieves a specific WLAN by ID from the local controller
// Maps to: GET /proxy/network/api/s/{site}/rest/wlanconf/{wlan_id}
func (c *LocalClient) GetWLAN(siteID, wlanID string) (*WLANResponse, error) {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/rest/wlanconf/%s", site, wlanID)

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result WLANResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse wlan response: %w", err)
	}

	return &result, nil
}

// UpdateWLAN updates an existing wireless network on the local controller
// Maps to: PUT /proxy/network/api/s/{site}/rest/wlanconf/{wlan_id}
// Note: This is a mutating operation and requires CSRF token (auto-injected)
func (c *LocalClient) UpdateWLAN(siteID, wlanID string, req UpdateWLANRequest) (*WLANResponse, error) {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/rest/wlanconf/%s", site, wlanID)

	// Build update payload with only provided fields
	payload := map[string]interface{}{}

	if req.Name != "" {
		payload["name"] = req.Name
	}

	if req.Security != "" {
		payload["security"] = req.Security
	}

	if req.Password != "" {
		payload["x_passphrase"] = req.Password
	}

	if req.Band != "" {
		payload["band"] = req.Band
	}

	if req.VLAN > 0 {
		payload["vlan"] = req.VLAN
	}

	if req.PMFMode != "" {
		payload["pmf_mode"] = req.PMFMode
	}

	// Boolean fields
	if req.Enabled != nil {
		payload["enabled"] = *req.Enabled
	}

	if req.Hidden != nil {
		payload["hide_ssid"] = *req.Hidden
	}

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Put(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result WLANResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse update wlan response: %w", err)
	}

	return &result, nil
}

// DeleteWLAN deletes a wireless network from the local controller
// Maps to: DELETE /proxy/network/api/s/{site}/rest/wlanconf/{wlan_id}
// Note: This is a mutating operation and requires CSRF token (auto-injected)
func (c *LocalClient) DeleteWLAN(siteID, wlanID string) error {
	site := c.resolveSite(siteID)
	endpoint := fmt.Sprintf("/proxy/network/api/s/%s/rest/wlanconf/%s", site, wlanID)

	resp, err := c.httpClient.R().Delete(endpoint)
	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	return nil
}

func (c *LocalClient) ListAlerts(siteID string, pageSize int, nextToken string, archived bool) (*AlertsResponse, error) {
	return nil, &NotImplementedError{Method: "ListAlerts"}
}

func (c *LocalClient) AcknowledgeAlert(siteID, alertID string) error {
	return &NotImplementedError{Method: "AcknowledgeAlert"}
}

func (c *LocalClient) ArchiveAlert(siteID, alertID string) error {
	return &NotImplementedError{Method: "ArchiveAlert"}
}

func (c *LocalClient) ListEvents(siteID string, pageSize int, nextToken string) (*EventsResponse, error) {
	return nil, &NotImplementedError{Method: "ListEvents"}
}

func (c *LocalClient) ListNetworks(siteID string, pageSize int, nextToken string) (*NetworksResponse, error) {
	return nil, &NotImplementedError{Method: "ListNetworks"}
}

func (c *LocalClient) EnableNetwork(siteID, networkID string) error {
	return &NotImplementedError{Method: "EnableNetwork"}
}

func (c *LocalClient) DisableNetwork(siteID, networkID string) error {
	return &NotImplementedError{Method: "DisableNetwork"}
}

func (c *LocalClient) Whoami() (*WhoamiResponse, error) {
	// Maps to: GET /api/auth/user-info
	endpoint := "/api/auth/user-info"

	resp, err := c.httpClient.R().Get(endpoint)
	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode(), endpoint, resp.Body())
	}

	var result WhoamiResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse whoami response: %w", err)
	}

	return &result, nil
}

// GetConnectionInfo returns metadata about the Local controller connection
func (c *LocalClient) GetConnectionInfo() ConnectionInfo {
	return ConnectionInfo{
		Mode:        "local",
		Endpoint:    c.baseURL,
		Version:     "UniFi OS",
		SiteID:      "default",
		IsConnected: c.httpClient != nil && c.csrfToken != "",
	}
}
