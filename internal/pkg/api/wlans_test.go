package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== WLAN Operation Tests ==========

func TestClient_ListWLANs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/wlans", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WLANsResponse{
			Code: "OK",
			Data: []WLAN{
				{ID: "wlan-1", Name: "Corporate", SSID: "CorpWiFi", Security: "WPA2", Enabled: true, SiteID: "site-123"},
				{ID: "wlan-2", Name: "Guest", SSID: "GuestWiFi", Security: "WPA2", Enabled: false, SiteID: "site-123"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListWLANs("site-123", 0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "wlan-1", resp.Data[0].ID)
	assert.Equal(t, "CorpWiFi", resp.Data[0].SSID)
	assert.True(t, resp.Data[0].Enabled)
	assert.False(t, resp.Data[1].Enabled)
}

func TestClient_ListWLANs_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/wlans", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "10", query.Get("pageSize"))
		assert.Equal(t, "token123", query.Get("nextToken"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WLANsResponse{
			Code:       "OK",
			Data:       []WLAN{{ID: "wlan-3", Name: "IoT", SSID: "IoTWiFi", Enabled: true}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListWLANs("site-123", 10, "token123")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "wlan-3", resp.Data[0].ID)
}

func TestClient_CreateWLAN_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/wlans", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req CreateWLANRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "New WLAN", req.Name)
		assert.Equal(t, "NewWiFi", req.SSID)
		assert.Equal(t, "WPA2", req.Security)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(WLANResponse{
			Code: "OK",
			Data: WLAN{
				ID:       "wlan-new",
				Name:     "New WLAN",
				SSID:     "NewWiFi",
				Security: "WPA2",
				Enabled:  true,
				SiteID:   "site-123",
			},
			HTTPStatus: 201,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	req := CreateWLANRequest{Name: "New WLAN", SSID: "NewWiFi", Security: "WPA2"}
	resp, err := client.CreateWLAN("site-123", req)
	require.NoError(t, err)
	assert.Equal(t, "wlan-new", resp.Data.ID)
	assert.Equal(t, "New WLAN", resp.Data.Name)
}

func TestClient_GetWLAN_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/wlans/wlan-456", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WLANResponse{
			Code: "OK",
			Data: WLAN{
				ID:       "wlan-456",
				Name:     "Test WLAN",
				SSID:     "TestSSID",
				Security: "WPA2",
				Enabled:  true,
				SiteID:   "site-123",
				VLAN:     10,
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetWLAN("site-123", "wlan-456")
	require.NoError(t, err)
	assert.Equal(t, "wlan-456", resp.Data.ID)
	assert.Equal(t, "Test WLAN", resp.Data.Name)
	assert.Equal(t, 10, resp.Data.VLAN)
}

func TestClient_UpdateWLAN_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/wlans/wlan-456", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req UpdateWLANRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Updated Name", req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WLANResponse{
			Code: "OK",
			Data: WLAN{
				ID:       "wlan-456",
				Name:     "Updated Name",
				SSID:     "TestSSID",
				Security: "WPA3",
				Enabled:  true,
				SiteID:   "site-123",
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	req := UpdateWLANRequest{Name: "Updated Name", Security: "WPA3"}
	resp, err := client.UpdateWLAN("site-123", "wlan-456", req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.Data.Name)
	assert.Equal(t, "WPA3", resp.Data.Security)
}

func TestClient_DeleteWLAN_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/wlans/wlan-456", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.DeleteWLAN("site-123", "wlan-456")
	require.NoError(t, err)
}

// ========== Network Operation Tests ==========

func TestClient_ListNetworks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/networks", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(NetworksResponse{
			Code: "OK",
			Data: []Network{
				{ID: "net-1", Name: "LAN", SiteID: "site-123", VLAN: 1, Subnet: "192.168.1.0/24", DHCPEnabled: true},
				{ID: "net-2", Name: "Guest", SiteID: "site-123", VLAN: 10, Subnet: "192.168.10.0/24", DHCPEnabled: true},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListNetworks("site-123", 0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "net-1", resp.Data[0].ID)
	assert.Equal(t, "LAN", resp.Data[0].Name)
	assert.Equal(t, 1, resp.Data[0].VLAN)
}

func TestClient_ListNetworks_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/networks", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "5", query.Get("pageSize"))
		assert.Equal(t, "token123", query.Get("nextToken"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(NetworksResponse{
			Code:       "OK",
			Data:       []Network{{ID: "net-3", Name: "IoT", VLAN: 20}},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListNetworks("site-123", 5, "token123")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "net-3", resp.Data[0].ID)
}

func TestClient_EnableNetwork_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/networks/net-456/enable", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.EnableNetwork("site-123", "net-456")
	require.NoError(t, err)
}

func TestClient_DisableNetwork_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/networks/net-456/disable", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.DisableNetwork("site-123", "net-456")
	require.NoError(t, err)
}

// ========== Health & Stats Tests ==========

func TestClient_GetSiteHealth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/health", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{
			Code: "OK",
			Data: []HealthStatus{
				{Subsystem: "wifi", Status: "ok", NumAdopted: 5, NumPending: 0, NumClient: 25},
				{Subsystem: "wan", Status: "ok", Latency: 15, PacketLoss: 0},
				{Subsystem: "lan", Status: "ok", NumAdopted: 3, NumPending: 0},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetSiteHealth("site-123")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 3)
	assert.Equal(t, "wifi", resp.Data[0].Subsystem)
	assert.Equal(t, "ok", resp.Data[0].Status)
	assert.Equal(t, 5, resp.Data[0].NumAdopted)
	assert.Equal(t, 25, resp.Data[0].NumClient)
}

func TestClient_GetSiteStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/stats", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "day", query.Get("period"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PerformanceResponse{
			Code: "OK",
			Data: PerformanceStats{
				SiteID:      "site-123",
				Period:      "day",
				RxBytes:     1073741824, // 1 GB
				TxBytes:     2147483648, // 2 GB
				RxRate:      100.5,
				TxRate:      50.2,
				Latency:     15.0,
				PacketLoss:  0.1,
				NumClients:  50,
				NumDevices:  10,
				CPUUsage:    25.5,
				MemoryUsage: 40.0,
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetSiteStats("site-123", "day")
	require.NoError(t, err)
	assert.Equal(t, "site-123", resp.Data.SiteID)
	assert.Equal(t, "day", resp.Data.Period)
	assert.Equal(t, int64(1073741824), resp.Data.RxBytes)
	assert.Equal(t, 100.5, resp.Data.RxRate)
	assert.Equal(t, 50, resp.Data.NumClients)
}

func TestClient_GetSiteStats_DefaultPeriod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sites/site-123/stats", r.URL.Path)

		query := r.URL.Query()
		assert.Empty(t, query.Get("period"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PerformanceResponse{
			Code:       "OK",
			Data:       PerformanceStats{SiteID: "site-123"},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetSiteStats("site-123", "")
	require.NoError(t, err)
	assert.Equal(t, "site-123", resp.Data.SiteID)
}

// ========== Host Operation Tests ==========

func TestClient_ListHosts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/hosts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HostsResponse{
			Code: "OK",
			Data: []Host{
				{ID: "host-1", Name: "UDM-Pro", Type: "udm", Model: "UDM-Pro", Version: "2.4.0", Status: "ONLINE"},
				{ID: "host-2", Name: "CloudKey", Type: "uck", Model: "UCK-G2-Plus", Version: "2.5.0", Status: "ONLINE"},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.ListHosts(0, "")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "host-1", resp.Data[0].ID)
	assert.Equal(t, "UDM-Pro", resp.Data[0].Name)
}

func TestClient_GetHost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/hosts/host-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HostResponse{
			Code: "OK",
			Data: Host{
				ID:         "host-123",
				Name:       "Main Console",
				Type:       "udm",
				Model:      "UDM-Pro",
				Version:    "2.4.0",
				Status:     "ONLINE",
				IPAddress:  "192.168.1.1",
				MACAddress: "aa:bb:cc:dd:ee:ff",
				Uptime:     86400,
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetHost("host-123")
	require.NoError(t, err)
	assert.Equal(t, "host-123", resp.Data.ID)
	assert.Equal(t, "Main Console", resp.Data.Name)
	assert.Equal(t, "192.168.1.1", resp.Data.IPAddress)
}

func TestClient_GetHostHealth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/hosts/host-123/health", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{
			Code: "OK",
			Data: []HealthStatus{
				{Subsystem: "system", Status: "ok", NumAdopted: 10},
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetHostHealth("host-123")
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "system", resp.Data[0].Subsystem)
}

func TestClient_GetHostStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/hosts/host-123/stats", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "week", query.Get("period"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PerformanceResponse{
			Code: "OK",
			Data: PerformanceStats{
				SiteID:  "host-123",
				Period:  "week",
				RxBytes: 1000000000,
				TxBytes: 2000000000,
			},
			HTTPStatus: 200,
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	resp, err := client.GetHostStats("host-123", "week")
	require.NoError(t, err)
	assert.Equal(t, "host-123", resp.Data.SiteID)
	assert.Equal(t, "week", resp.Data.Period)
}

func TestClient_RestartHost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/hosts/host-123/restart", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{BaseURL: server.URL, APIKey: "test-api-key"})
	require.NoError(t, err)

	err = client.RestartHost("host-123")
	require.NoError(t, err)
}
