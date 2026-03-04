package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Site Operation Tests ==========

func TestClient_ListSites_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SitesResponse{
			Code: "OK",
			Data: []Site{
				{ID: "site-1", Name: "Office", Description: "Main office", HostID: "host-1", Status: "active"},
				{ID: "site-2", Name: "Home", Description: "Home network", HostID: "host-2", Status: "active"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListSites(0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "site-1", resp.Data[0].ID)
	assert.Equal(t, "Office", resp.Data[0].Name)
}

func TestClient_ListSites_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites", r.URL.Path)

		// Check query parameters
		query := r.URL.Query()
		if query.Get("pageSize") == "10" && query.Get("nextToken") == "token123" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SitesResponse{
				Code: "OK",
				Data: []Site{
					{ID: "site-3", Name: "Remote", Status: "active"},
				},
				HTTPStatus: 200,
				NextToken:  "",
			})
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SitesResponse{
				Code: "OK",
				Data: []Site{
					{ID: "site-1", Name: "Office", Status: "active"},
					{ID: "site-2", Name: "Home", Status: "active"},
				},
				HTTPStatus: 200,
				NextToken:  "token123",
			})
		}
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListSites(0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "token123", resp.NextToken)

	// Get next page
	resp2, err := client.ListSites(10, "token123")
	require.NoError(t, err)
	assert.Len(t, resp2.Data, 1)
	assert.Equal(t, "site-3", resp2.Data[0].ID)
}

func TestClient_ListSites_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid API key"})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "invalid-key"})
	require.NoError(t, err)

	resp, err := client.ListSites(0, "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.IsType(t, &AuthError{}, err)
}

func TestClient_ListSites_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListSites(0, "")
	assert.Error(t, err)
	assert.Nil(t, resp)

	var rateLimitErr *RateLimitError
	assert.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 60, rateLimitErr.RetryAfter)
}

func TestClient_GetSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SiteResponse{
			Code: "OK",
			Data: Site{
				ID:          "site-123",
				Name:        "Test Site",
				Description: "A test site",
				HostID:      "host-1",
				Status:      "active",
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetSite("site-123")
	require.NoError(t, err)
	assert.Equal(t, "site-123", resp.Data.ID)
	assert.Equal(t, "Test Site", resp.Data.Name)
}

func TestClient_GetSite_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetSite("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.IsType(t, &NotFoundError{}, err)
}

func TestClient_CreateSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req CreateSiteRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "New Site", req.Name)
		assert.Equal(t, "Site description", req.Description)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(SiteResponse{
			Code: "OK",
			Data: Site{
				ID:          "new-site-123",
				Name:        "New Site",
				Description: "Site description",
				Status:      "active",
			},
			HTTPStatus: 201,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	req := CreateSiteRequest{Name: "New Site", Description: "Site description"}
	resp, err := client.CreateSite(req)
	require.NoError(t, err)
	assert.Equal(t, "new-site-123", resp.Data.ID)
	assert.Equal(t, "New Site", resp.Data.Name)
}

func TestClient_CreateSite_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    "VALIDATION_ERROR",
			Message: "name is required",
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	req := CreateSiteRequest{Name: ""}
	resp, err := client.CreateSite(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.IsType(t, &ValidationError{}, err)
}

func TestClient_UpdateSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req UpdateSiteRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Updated Name", req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SiteResponse{
			Code: "OK",
			Data: Site{
				ID:          "site-123",
				Name:        "Updated Name",
				Description: "Updated description",
				Status:      "active",
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	req := UpdateSiteRequest{Name: "Updated Name", Description: "Updated description"}
	resp, err := client.UpdateSite("site-123", req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.Data.Name)
}

func TestClient_DeleteSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.DeleteSite("site-123")
	require.NoError(t, err)
}

func TestClient_DeleteSite_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.DeleteSite("nonexistent")
	assert.Error(t, err)
	assert.IsType(t, &NotFoundError{}, err)
}

// ========== Device Operation Tests ==========

func TestClient_ListDevices_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/devices", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DevicesResponse{
			Code: "OK",
			Data: []Device{
				{ID: "dev-1", Name: "AP1", Type: "uap", Model: "U6-Lite", Status: "ONLINE", SiteID: "site-123"},
				{ID: "dev-2", Name: "Switch1", Type: "usw", Model: "USW-Pro", Status: "ONLINE", SiteID: "site-123"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListDevices("site-123", 0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "dev-1", resp.Data[0].ID)
	assert.Equal(t, "AP1", resp.Data[0].Name)
}

func TestClient_ListDevices_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/devices", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "5", query.Get("pageSize"))
		assert.Equal(t, "next-page-token", query.Get("nextToken"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DevicesResponse{
			Code:       "OK",
			Data:       []Device{{ID: "dev-3", Name: "AP3", Type: "uap", Status: "ONLINE"}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListDevices("site-123", 5, "next-page-token")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "dev-3", resp.Data[0].ID)
}

func TestClient_GetDevice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/devices/dev-456", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DeviceResponse{
			Code: "OK",
			Data: Device{
				ID:         "dev-456",
				Name:       "Office AP",
				Type:       "uap",
				Model:      "U6-Lite",
				Status:     "ONLINE",
				SiteID:     "site-123",
				IPAddress:  "192.168.1.100",
				MACAddress: "aa:bb:cc:dd:ee:ff",
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetDevice("site-123", "dev-456")
	require.NoError(t, err)
	assert.Equal(t, "dev-456", resp.Data.ID)
	assert.Equal(t, "Office AP", resp.Data.Name)
	assert.Equal(t, "192.168.1.100", resp.Data.IPAddress)
}

func TestClient_RestartDevice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/devices/dev-456/restart", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "dev-456", body["deviceId"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.RestartDevice("site-123", "dev-456")
	require.NoError(t, err)
}

func TestClient_UpgradeDevice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/devices/dev-456/upgrade", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.UpgradeDevice("site-123", "dev-456")
	require.NoError(t, err)
}

func TestClient_AdoptDevice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/devices/adopt", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req AdoptDeviceRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "aa:bb:cc:dd:ee:ff", req.MACAddress)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.AdoptDevice("site-123", "aa:bb:cc:dd:ee:ff")
	require.NoError(t, err)
}

// ========== Client Operation Tests ==========

func TestClient_ListClients_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/clients", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ClientsResponse{
			Code: "OK",
			Data: []NetworkClient{
				{ID: "client-1", Name: "Laptop", MACAddress: "aa:bb:cc:dd:ee:01", IPAddress: "192.168.1.10", IsWired: true, SiteID: "site-123"},
				{ID: "client-2", Hostname: "Phone", MACAddress: "aa:bb:cc:dd:ee:02", IPAddress: "192.168.1.11", IsWired: false, SiteID: "site-123"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListClients("site-123", 0, "", false, false)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "client-1", resp.Data[0].ID)
	assert.True(t, resp.Data[0].IsWired)
	assert.False(t, resp.Data[1].IsWired)
}

func TestClient_ListClients_WiredOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/clients", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "true", query.Get("wired"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ClientsResponse{
			Code:       "OK",
			Data:       []NetworkClient{{ID: "client-1", Name: "Wired Device", IsWired: true}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListClients("site-123", 0, "", true, false)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
}

func TestClient_ListClients_WirelessOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/clients", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "true", query.Get("wireless"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ClientsResponse{
			Code:       "OK",
			Data:       []NetworkClient{{ID: "client-1", Name: "Wireless Device", IsWired: false}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListClients("site-123", 0, "", false, true)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
}

func TestClient_GetClientStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/clients/aa:bb:cc:dd:ee:ff/stats", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SingleResponse[ClientStats]{
			Code: "OK",
			Data: ClientStats{
				MACAddress:   "aa:bb:cc:dd:ee:ff",
				RxBytes:      1024000,
				TxBytes:      512000,
				RxPackets:    1000,
				TxPackets:    500,
				Satisfaction: 95.5,
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetClientStats("site-123", "aa:bb:cc:dd:ee:ff")
	require.NoError(t, err)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", resp.Data.MACAddress)
	assert.Equal(t, int64(1024000), resp.Data.RxBytes)
	assert.Equal(t, 95.5, resp.Data.Satisfaction)
}

func TestClient_BlockClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/clients/block", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req BlockClientRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "aa:bb:cc:dd:ee:ff", req.MACAddress)
		assert.True(t, req.Block)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.BlockClient("site-123", "aa:bb:cc:dd:ee:ff", true)
	require.NoError(t, err)
}

func TestClient_UnblockClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/clients/block", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req BlockClientRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "aa:bb:cc:dd:ee:ff", req.MACAddress)
		assert.False(t, req.Block)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.BlockClient("site-123", "aa:bb:cc:dd:ee:ff", false)
	require.NoError(t, err)
}
