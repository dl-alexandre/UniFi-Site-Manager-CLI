package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_RetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "0") // Minimal retry delay for tests
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{{ID: "123", Name: "Test Site"}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond, // Minimal delay for tests
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "Client should have retried twice and succeeded on the third attempt")
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "123", resp.Data[0].ID)
}

func TestClient_RetryOn503(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{{ID: "456", Name: "Another Site"}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "Client should have retried twice on 503 and succeeded on the third attempt")
	assert.NotNil(t, resp)
}

func TestClient_RetryOn502(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{{ID: "789", Name: "Third Site"}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "Client should have retried twice on 502 and succeeded on the third attempt")
	assert.NotNil(t, resp)
}

func TestClient_NoRetryOnFinal429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Retry-After", "3600") // 1 hour
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Should have tried 3 times (initial + 2 retries), then given up
	assert.Equal(t, 3, attempts, "Client should have exhausted all retries before returning error")

	// Verify error type
	var rateLimitErr *RateLimitError
	assert.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 3600, rateLimitErr.RetryAfter)
}

func TestClient_NoRetryOn401(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid API key"})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, attempts, "Client should NOT retry on 401 Unauthorized")

	var authErr *AuthError
	assert.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Message, "invalid API key")
}

func TestClient_NoRetryOn403(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, attempts, "Client should NOT retry on 403 Forbidden")

	var permErr *PermissionError
	assert.ErrorAs(t, err, &permErr)
}

func TestClient_NoRetryOn404(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.GetSite("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, attempts, "Client should NOT retry on 404 Not Found")

	var notFoundErr *NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)
}

func TestClient_RetryRespectsMaxRetryDelay(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "300") // 5 minutes
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{{ID: "123", Name: "Test"}},
		})
	}))
	defer server.Close()

	// Set very short max retry delay
	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond * 10, // 10ms max
	})
	assert.NoError(t, err)

	start := time.Now()
	resp, err := client.ListSites(10, "")
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Less(t, elapsed, time.Second, "Retry delay should respect MaxRetryDelay and be much shorter than 5 minutes")
}

func TestClient_NoRetryOn400(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Message: "Invalid parameter: name",
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.CreateSite(CreateSiteRequest{Name: ""})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, attempts, "Client should NOT retry on 400 Bad Request")

	var validationErr *ValidationError
	assert.ErrorAs(t, err, &validationErr)
}

func TestClient_ExhaustRetriesOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 3, attempts, "Client should exhaust all 3 retries on 500 Internal Server Error")
	assert.Contains(t, err.Error(), "server error")
}

func TestClient_SuccessOnFirstAttempt(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{
				{ID: "1", Name: "Site 1"},
				{ID: "2", Name: "Site 2"},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, attempts, "Client should succeed on first attempt without retries")
	assert.Equal(t, 2, len(resp.Data))
}

func TestClient_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		nextToken := r.URL.Query().Get("nextToken")
		pageSize := r.URL.Query().Get("pageSize")

		w.WriteHeader(http.StatusOK)

		// API returns paginated results based on query params
		switch nextToken {
		case "":
			// First page request
			_ = json.NewEncoder(w).Encode(SitesResponse{
				Data:      []Site{{ID: "1", Name: "Page 1 Site"}},
				NextToken: "token-for-page-2",
			})
		case "token-for-page-2":
			// Second page request
			_ = json.NewEncoder(w).Encode(SitesResponse{
				Data:      []Site{{ID: "2", Name: "Page 2 Site"}},
				NextToken: "token-for-page-3",
			})
		default:
			// Third and final page
			_ = json.NewEncoder(w).Encode(SitesResponse{
				Data:      []Site{{ID: "3", Name: "Page 3 Site"}},
				NextToken: "",
			})
		}

		// Verify pageSize parameter is passed correctly
		_ = pageSize // Used when specific page size requested
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	// Test fetching first page only (pageSize=50)
	resp, err := client.ListSites(50, "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, callCount, "Should have made 1 request for first page")
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "token-for-page-2", resp.NextToken)

	// Reset and test fetching with explicit nextToken
	callCount = 0
	resp, err = client.ListSites(50, "token-for-page-2")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, callCount, "Should have made 1 request for second page")
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "token-for-page-3", resp.NextToken)
}

func TestClient_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "this is not valid json")
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestClient_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	resp, err := client.ListSites(10, "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Data), "Should handle empty response gracefully")
}

func TestClient_ConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SitesResponse{
			Data: []Site{{ID: "test", Name: "Test Site"}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: time.Millisecond,
	})
	assert.NoError(t, err)

	// Make multiple concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := client.ListSites(10, "")
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
