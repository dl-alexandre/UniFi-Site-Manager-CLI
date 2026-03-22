package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Alert Operation Tests ==========

func TestClient_ListAlerts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/alerts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AlertsResponse{
			Code: "OK",
			Data: []Alert{
				{ID: "alert-1", Type: "DISCONNECT", Severity: "warning", Message: "Device went offline", SiteID: "site-123", DeviceID: "dev-1", Acknowledged: false},
				{ID: "alert-2", Type: "ROGUE_AP", Severity: "info", Message: "Rogue AP detected", SiteID: "site-123", Acknowledged: true},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListAlerts("site-123", 0, "", false)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "alert-1", resp.Data[0].ID)
	assert.Equal(t, "DISCONNECT", resp.Data[0].Type)
	assert.Equal(t, "warning", resp.Data[0].Severity)
	assert.False(t, resp.Data[0].Acknowledged)
	assert.True(t, resp.Data[1].Acknowledged)
}

func TestClient_ListAlerts_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/alerts", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "10", query.Get("pageSize"))
		assert.Equal(t, "token123", query.Get("nextToken"))
		assert.Equal(t, "true", query.Get("archived"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AlertsResponse{
			Code:       "OK",
			Data:       []Alert{{ID: "alert-3", Type: "INFO", Severity: "info"}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListAlerts("site-123", 10, "token123", true)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
}

func TestClient_ListAlerts_Global(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/alerts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AlertsResponse{
			Code: "OK",
			Data: []Alert{
				{ID: "alert-1", Type: "GLOBAL", Severity: "critical", Message: "System maintenance"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListAlerts("", 0, "", false)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "GLOBAL", resp.Data[0].Type)
}

func TestClient_AcknowledgeAlert_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/alerts/alert-456/ack", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req AcknowledgeAlertRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "alert-456", req.AlertID)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.AcknowledgeAlert("site-123", "alert-456")
	require.NoError(t, err)
}

func TestClient_ArchiveAlert_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/alerts/alert-456/archive", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req ArchiveAlertRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "alert-456", req.AlertID)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.ArchiveAlert("site-123", "alert-456")
	require.NoError(t, err)
}

// ========== Event Operation Tests ==========

func TestClient_ListEvents_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/events", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(EventsResponse{
			Code: "OK",
			Data: []Event{
				{ID: "event-1", Type: "CONNECTED", Message: "Client connected", SiteID: "site-123", ClientID: "client-1", Timestamp: "2024-01-15T10:30:00Z"},
				{ID: "event-2", Type: "DISCONNECTED", Message: "Client disconnected", SiteID: "site-123", ClientID: "client-2", Timestamp: "2024-01-15T10:35:00Z"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListEvents("site-123", 0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "event-1", resp.Data[0].ID)
	assert.Equal(t, "CONNECTED", resp.Data[0].Type)
	assert.Equal(t, "Client connected", resp.Data[0].Message)
}

func TestClient_ListEvents_Global(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/events", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(EventsResponse{
			Code:       "OK",
			Data:       []Event{{ID: "event-global", Type: "SYSTEM", Message: "System event"}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListEvents("", 0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
}

func TestClient_ListEvents_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/events", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "20", query.Get("pageSize"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(EventsResponse{
			Code:       "OK",
			Data:       []Event{{ID: "event-3", Type: "ROAMING", Message: "Client roamed"}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListEvents("site-123", 20, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
}

// ========== User/Auth Tests ==========

func TestClient_Whoami_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/whoami", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(WhoamiResponse{
			Code: "OK",
			Data: UserInfo{
				ID:        "user-123",
				Email:     "admin@example.com",
				FirstName: "Admin",
				LastName:  "User",
				Role:      "superadmin",
				IsOwner:   true,
				Sites:     []string{"site-1", "site-2"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.Whoami()
	require.NoError(t, err)
	assert.Equal(t, "user-123", resp.Data.ID)
	assert.Equal(t, "admin@example.com", resp.Data.Email)
	assert.Equal(t, "Admin", resp.Data.FirstName)
	assert.Equal(t, "User", resp.Data.LastName)
	assert.Equal(t, "superadmin", resp.Data.Role)
	assert.True(t, resp.Data.IsOwner)
	assert.Len(t, resp.Data.Sites, 2)
}

func TestClient_Whoami_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "invalid-key"})
	require.NoError(t, err)

	resp, err := client.Whoami()
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.IsType(t, &AuthError{}, err)
}
