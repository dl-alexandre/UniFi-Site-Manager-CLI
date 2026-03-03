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

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(req *resty.Request, endpoint string) (*resty.Response, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := req.Execute(req.Method, endpoint)

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

// ListSites retrieves a list of all sites
func (c *Client) ListSites(pageSize int, nextToken string) (*SitesResponse, error) {
	req := c.httpClient.R()

	if pageSize > 0 {
		req.SetQueryParam("pageSize", strconv.Itoa(pageSize))
	}
	if nextToken != "" {
		req.SetQueryParam("nextToken", nextToken)
	}

	resp, err := c.doRequest(req, "/v1/sites")
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
	req := c.httpClient.R()

	resp, err := c.doRequest(req, fmt.Sprintf("/v1/sites/%s", siteID))
	if err != nil {
		return nil, err
	}

	var result SiteResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse site response: %w", err)
	}

	return &result, nil
}

// Whoami retrieves information about the authenticated user
func (c *Client) Whoami() (*WhoamiResponse, error) {
	req := c.httpClient.R()

	resp, err := c.doRequest(req, "/v1/whoami")
	if err != nil {
		return nil, err
	}

	var result WhoamiResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse whoami response: %w", err)
	}

	return &result, nil
}
