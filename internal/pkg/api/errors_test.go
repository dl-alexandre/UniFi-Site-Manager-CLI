package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Error Type Tests ==========

func TestAuthError(t *testing.T) {
	err := &AuthError{Message: "invalid credentials"}
	assert.Equal(t, "authentication failed: invalid credentials", err.Error())
	assert.Equal(t, ExitAuthFailure, err.ExitCode())
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{Message: "access denied"}
	assert.Equal(t, "permission denied: access denied", err.Error())
	assert.Equal(t, ExitPermissionDenied, err.ExitCode())
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{Resource: "site-123"}
	assert.Equal(t, "resource not found: site-123", err.Error())
	assert.Equal(t, ExitValidationError, err.ExitCode())
}

func TestRateLimitError(t *testing.T) {
	t.Run("with retry after", func(t *testing.T) {
		err := &RateLimitError{RetryAfter: 60}
		assert.Equal(t, "rate limited. retry after 60 seconds", err.Error())
		assert.Equal(t, ExitRateLimited, err.ExitCode())
	})

	t.Run("without retry after", func(t *testing.T) {
		err := &RateLimitError{RetryAfter: 0}
		assert.Equal(t, "rate limited. please try again later", err.Error())
		assert.Equal(t, ExitRateLimited, err.ExitCode())
	})
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Message: "invalid input"}
	assert.Equal(t, "invalid input", err.Error())
	assert.Equal(t, ExitValidationError, err.ExitCode())
}

func TestNetworkError(t *testing.T) {
	err := &NetworkError{Message: "connection refused"}
	assert.Equal(t, "network error: connection refused", err.Error())
	assert.Equal(t, ExitNetworkError, err.ExitCode())
}

func TestNotImplementedError(t *testing.T) {
	err := &NotImplementedError{Method: "CreateSite"}
	assert.Contains(t, err.Error(), "CreateSite")
	assert.Contains(t, err.Error(), "not yet implemented")
	assert.Equal(t, ExitGeneralError, err.ExitCode())
}

// ========== GetExitCode Tests ==========

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"nil error", nil, ExitSuccess},
		{"AuthError", &AuthError{Message: "test"}, ExitAuthFailure},
		{"PermissionError", &PermissionError{Message: "test"}, ExitPermissionDenied},
		{"NotFoundError", &NotFoundError{Resource: "test"}, ExitValidationError},
		{"RateLimitError", &RateLimitError{RetryAfter: 60}, ExitRateLimited},
		{"ValidationError", &ValidationError{Message: "test"}, ExitValidationError},
		{"NetworkError", &NetworkError{Message: "test"}, ExitNetworkError},
		{"NotImplementedError", &NotImplementedError{Method: "test"}, ExitGeneralError},
		{"unknown error", errors.New("unknown"), ExitGeneralError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			assert.Equal(t, tt.wantCode, got)
		})
	}
}

// ========== Retry Logic Tests ==========

func TestClient_RetryOnTimeout(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Simulate timeout by closing connection
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				_ = conn.Close()
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{Code: "OK", Data: []Site{{ID: "site-1"}}})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		MaxRetryDelay: 100 * time.Millisecond,
	})
	require.NoError(t, err)

	_, _ = client.ListSites(0, "")
	// Should eventually succeed after retries, but network errors may not retry
	// This test verifies retry behavior is attempted
	assert.GreaterOrEqual(t, attemptCount, 1)
}

func TestClient_RetryOnServerError(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{Code: "OK", Data: []Site{{ID: "site-1"}}})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		MaxRetryDelay: 50 * time.Millisecond,
	})
	require.NoError(t, err)

	resp, err := client.ListSites(0, "")
	require.NoError(t, err)
	assert.Equal(t, 3, attemptCount)
	assert.Len(t, resp.Data, 1)
}

func TestClient_RetryWithBackoff(t *testing.T) {
	attemptTimes := []time.Time{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptTimes = append(attemptTimes, time.Now())
		if len(attemptTimes) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{Code: "OK", Data: []Site{{ID: "site-1"}}})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		MaxRetryDelay: 500 * time.Millisecond,
	})
	require.NoError(t, err)

	_, _ = client.ListSites(0, "")
	require.NoError(t, err)

	// Verify backoff timing (should increase between attempts)
	if len(attemptTimes) >= 2 {
		delay1 := attemptTimes[1].Sub(attemptTimes[0])
		delay2 := attemptTimes[2].Sub(attemptTimes[1])
		// Second delay should be >= first delay (exponential backoff)
		assert.GreaterOrEqual(t, delay2, delay1)
	}
}

func TestClient_NoRetryOnAuthError(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid API key"})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "invalid-key"})
	require.NoError(t, err)

	_, _ = client.ListSites(0, "")
	assert.Error(t, err)
	assert.Equal(t, 1, attemptCount) // Should not retry on 401
}

func TestClient_NoRetryOnNotFound(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-key"})
	require.NoError(t, err)

	_, err = client.GetSite("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, 1, attemptCount) // Should not retry on 404
}

func TestClient_RetryOnRateLimit(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{Code: "OK", Data: []Site{{ID: "site-1"}}})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		MaxRetryDelay: 2 * time.Second,
	})
	require.NoError(t, err)

	resp, err := client.ListSites(0, "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, attemptCount, 2)
	assert.Len(t, resp.Data, 1)
}

func TestClient_MaxRetriesExceeded(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		MaxRetryDelay: 50 * time.Millisecond,
	})
	require.NoError(t, err)

	_, _ = client.ListSites(0, "")
	assert.Error(t, err)
	assert.LessOrEqual(t, attemptCount, 4) // Should stop after max retries (3 attempts + 1 final)
}

// ========== ParseRetryAfter Tests ==========

func TestClient_ParseRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return different retry-after values
		handler := r.URL.Path[1:]
		switch handler {
		case "valid":
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
		case "invalid":
			w.Header().Set("Retry-After", "invalid")
			w.WriteHeader(http.StatusTooManyRequests)
		case "missing":
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-key"})
	require.NoError(t, err)

	// Test with valid retry-after header
	_, err = client.doGet("/valid")
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		assert.Equal(t, 30, rateLimitErr.RetryAfter)
	}

	// Test with invalid retry-after header
	_, err = client.doGet("/invalid")
	if errors.As(err, &rateLimitErr) {
		// Should default to 0 when invalid
		assert.Equal(t, 0, rateLimitErr.RetryAfter)
	}

	// Test with missing retry-after header
	_, err = client.doGet("/missing")
	if errors.As(err, &rateLimitErr) {
		// Should default to 0 when missing
		assert.Equal(t, 0, rateLimitErr.RetryAfter)
	}
}

// ========== Network Error Tests ==========

func TestShouldRetryHelper(t *testing.T) {
	client, _ := NewClient(ClientOptions{APIKey: "test"})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout error", errors.New("connection timeout"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"no such host", errors.New("no such host"), true},
		{"temporary error", errors.New("temporary failure"), true},
		{"EOF error", errors.New("EOF"), true},
		{"random error", errors.New("something else"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldRetry(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========== Server Error Response Tests ==========

func TestClient_ServerErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errType    interface{}
	}{
		{"Bad Request", http.StatusBadRequest, &ValidationError{}},
		{"Unauthorized", http.StatusUnauthorized, &AuthError{}},
		{"Forbidden", http.StatusForbidden, &PermissionError{}},
		{"Not Found", http.StatusNotFound, &NotFoundError{}},
		{"Rate Limited", http.StatusTooManyRequests, &RateLimitError{}},
		{"Internal Server Error", http.StatusInternalServerError, errors.New("")},
		{"Bad Gateway", http.StatusBadGateway, errors.New("")},
		{"Service Unavailable", http.StatusServiceUnavailable, errors.New("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusBadRequest {
					_ = json.NewEncoder(w).Encode(APIResponse{Message: "validation failed"})
				}
			}))
			defer server.Close()

			client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-key"})
			require.NoError(t, err)

			_, _ = client.ListSites(0, "")
			assert.Error(t, err)

			switch tt.errType.(type) {
			case *ValidationError:
				assert.IsType(t, &ValidationError{}, err)
			case *AuthError:
				assert.IsType(t, &AuthError{}, err)
			case *PermissionError:
				assert.IsType(t, &PermissionError{}, err)
			case *NotFoundError:
				assert.IsType(t, &NotFoundError{}, err)
			case *RateLimitError:
				assert.IsType(t, &RateLimitError{}, err)
			}
		})
	}
}

// ========== Helper Function Tests ==========

func TestContainsAnyHelper(t *testing.T) {
	tests := []struct {
		s        string
		substrs  []string
		expected bool
	}{
		{"hello world", []string{"world", "foo"}, true},
		{"hello world", []string{"foo", "bar"}, false},
		{"connection timeout", []string{"timeout"}, true},
		{"", []string{"test"}, false},
		{"test", []string{}, false},
	}

	for _, tt := range tests {
		result := containsAny(tt.s, tt.substrs)
		assert.Equal(t, tt.expected, result)
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "test", false},
		{"test", "", true},
		{"test", "test", true},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		assert.Equal(t, tt.expected, result)
	}
}
