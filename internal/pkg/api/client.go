// Package api provides HTTP client for UniFi Site Manager API
package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultBaseURL = "https://api.ui.com"
	apiVersion     = "v1"
)

// Client wraps the HTTP client for UniFi Site Manager API
// This is the Cloud API implementation of the SiteManager interface
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

// SetBaseURL updates the client's base URL (primarily for testing)
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
	c.httpClient.SetBaseURL(url)
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

// shouldRetry determines if a request should be retried based on the error
func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsAny(errStr, []string{"timeout", "connection refused", "no such host", "temporary", "EOF"})
}

// calculateBackoff calculates the sleep duration for retry attempts using exponential backoff with jitter
func (c *Client) calculateBackoff(attempt int) time.Duration {
	base := time.Duration(1<<attempt) * time.Second
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return base + jitter
}

// parseRetryAfter parses the Retry-After header from a response
func (c *Client) parseRetryAfter(resp *resty.Response) int {
	header := resp.Header().Get("Retry-After")
	if header == "" {
		return 0
	}
	seconds, err := strconv.Atoi(header)
	if err != nil {
		return 0
	}
	return seconds
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrs []string) bool {
	lower := ""
	for _, substr := range substrs {
		if lower == "" {
			lower = s
		}
		if contains(lower, substr) {
			return true
		}
	}
	return false
}

// contains is a simple substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && containsAt(s, substr, 0)))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
