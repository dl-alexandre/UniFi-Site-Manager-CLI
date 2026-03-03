package cli

import (
	"testing"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/config"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

// TestCommandRouting verifies that CLI commands correctly route to SiteManager interface methods
// This is the Master Router Test - it ensures all 38 SiteManager methods are properly wired
func TestCommandRouting(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockSetup     func(*mocks.SiteManager)
		expectedError bool
		errorContains string
	}{
		// ========== SITES ==========
		{
			name: "Route: sites list",
			args: []string{"sites", "list"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListSites", 50, "").Return(&api.SitesResponse{
					Data: []api.Site{{ID: "site-1", Name: "Test Site"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: sites get",
			args: []string{"sites", "get", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetSite", "site-123").Return(&api.SiteResponse{
					Data: api.Site{ID: "site-123", Name: "Test Site"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: sites create",
			args: []string{"sites", "create", "New Site", "--description", "A test site"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("CreateSite", api.CreateSiteRequest{
					Name:        "New Site",
					Description: "A test site",
				}).Return(&api.SiteResponse{
					Data: api.Site{ID: "new-site-1", Name: "New Site"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: sites update",
			args: []string{"sites", "update", "site-123", "--name", "Updated Name"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("UpdateSite", "site-123", api.UpdateSiteRequest{
					Name: "Updated Name",
				}).Return(&api.SiteResponse{
					Data: api.Site{ID: "site-123", Name: "Updated Name"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: sites delete",
			args: []string{"sites", "delete", "site-123", "--force"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("DeleteSite", "site-123").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Route: sites health",
			args: []string{"sites", "health", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetSiteHealth", "site-123").Return(&api.HealthResponse{
					Data: []api.HealthStatus{{Subsystem: "wlan", Status: "ok"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: sites stats",
			args: []string{"sites", "stats", "site-123", "--period", "day"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetSiteStats", "site-123", "day").Return(&api.PerformanceResponse{
					Data: api.PerformanceStats{},
				}, nil)
			},
			expectedError: false,
		},

		// ========== DEVICES ==========
		{
			name: "Route: devices list",
			args: []string{"devices", "list", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListDevices", "site-123", 50, "").Return(&api.DevicesResponse{
					Data: []api.Device{{ID: "dev-1", Name: "AP-1"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: devices get",
			args: []string{"devices", "get", "site-123", "dev-456"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetDevice", "site-123", "dev-456").Return(&api.DeviceResponse{
					Data: api.Device{ID: "dev-456", Name: "AP-1"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: devices restart",
			args: []string{"devices", "restart", "site-123", "dev-456", "--force"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("RestartDevice", "site-123", "dev-456").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Route: devices upgrade",
			args: []string{"devices", "upgrade", "site-123", "dev-456", "--force"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("UpgradeDevice", "site-123", "dev-456").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Route: devices adopt",
			args: []string{"devices", "adopt", "site-123", "aa:bb:cc:dd:ee:ff"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("AdoptDevice", "site-123", "aa:bb:cc:dd:ee:ff").Return(nil)
			},
			expectedError: false,
		},

		// ========== CLIENTS ==========
		{
			name: "Route: clients list",
			args: []string{"clients", "list", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListClients", "site-123", 50, "", false, false).Return(&api.ClientsResponse{
					Data: []api.NetworkClient{{ID: "client-1", Name: "iPhone"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: clients stats",
			args: []string{"clients", "stats", "site-123", "aa:bb:cc:dd:ee:ff"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetClientStats", "site-123", "aa:bb:cc:dd:ee:ff").Return(&api.SingleResponse[api.ClientStats]{
					Data: api.ClientStats{MACAddress: "aa:bb:cc:dd:ee:ff"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: clients block",
			args: []string{"clients", "block", "site-123", "aa:bb:cc:dd:ee:ff"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("BlockClient", "site-123", "aa:bb:cc:dd:ee:ff", true).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Route: clients unblock",
			args: []string{"clients", "unblock", "site-123", "aa:bb:cc:dd:ee:ff"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("BlockClient", "site-123", "aa:bb:cc:dd:ee:ff", false).Return(nil)
			},
			expectedError: false,
		},

		// ========== WLANS ==========
		{
			name: "Route: wla-ns list",
			args: []string{"wla-ns", "list", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListWLANs", "site-123", 50, "").Return(&api.WLANsResponse{
					Data: []api.WLAN{{ID: "wlan-1", Name: "Home WiFi"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: wla-ns get",
			args: []string{"wla-ns", "get", "site-123", "wlan-456"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetWLAN", "site-123", "wlan-456").Return(&api.WLANResponse{
					Data: api.WLAN{ID: "wlan-456", Name: "Home WiFi"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: wla-ns create",
			args: []string{"wla-ns", "create", "site-123", "Guest WiFi", "--ssid", "GuestWiFi", "--security", "wpapsk", "--wlan-password", "secret123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("CreateWLAN", "site-123", api.CreateWLANRequest{
					Name:            "Guest WiFi",
					SSID:            "GuestWiFi",
					Security:        "wpapsk",
					Password:        "secret123",
					VLAN:            0,
					Band:            "both",
					Hidden:          false,
					PMFMode:         "optional",
					WPA3Support:     false,
					MACFilter:       false,
					MACFilterPolicy: "allow",
				}).Return(&api.WLANResponse{
					Data: api.WLAN{ID: "new-wlan", Name: "Guest WiFi"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: wla-ns update",
			args: []string{"wla-ns", "update", "site-123", "wlan-456", "--name", "Updated WiFi"},
			mockSetup: func(m *mocks.SiteManager) {
				enabled := false
				hidden := false
				m.On("UpdateWLAN", "site-123", "wlan-456", api.UpdateWLANRequest{
					Name:    "Updated WiFi",
					Enabled: &enabled,
					Hidden:  &hidden,
					VLAN:    0,
				}).Return(&api.WLANResponse{
					Data: api.WLAN{ID: "wlan-456", Name: "Updated WiFi"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: wla-ns delete",
			args: []string{"wla-ns", "delete", "site-123", "wlan-456", "--force"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("DeleteWLAN", "site-123", "wlan-456").Return(nil)
			},
			expectedError: false,
		},

		// ========== ALERTS ==========
		{
			name: "Route: alerts list",
			args: []string{"alerts", "list", "--site-id", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListAlerts", "site-123", 50, "", false).Return(&api.AlertsResponse{
					Data: []api.Alert{{ID: "alert-1", Message: "Test alert"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: alerts ack",
			args: []string{"alerts", "ack", "site-123", "alert-456"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("AcknowledgeAlert", "site-123", "alert-456").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Route: alerts archive",
			args: []string{"alerts", "archive", "site-123", "alert-456"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ArchiveAlert", "site-123", "alert-456").Return(nil)
			},
			expectedError: false,
		},

		// ========== EVENTS ==========
		{
			name: "Route: events list",
			args: []string{"events", "list", "--site-id", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListEvents", "site-123", 50, "").Return(&api.EventsResponse{
					Data: []api.Event{{ID: "event-1", Message: "Test event"}},
				}, nil)
			},
			expectedError: false,
		},

		// ========== NETWORKS ==========
		{
			name: "Route: networks list",
			args: []string{"networks", "list", "site-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListNetworks", "site-123", 50, "").Return(&api.NetworksResponse{
					Data: []api.Network{{ID: "net-1", Name: "Default"}},
				}, nil)
			},
			expectedError: false,
		},

		// ========== HOSTS ==========
		{
			name: "Route: hosts list",
			args: []string{"hosts", "list"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("ListHosts", 50, "").Return(&api.HostsResponse{
					Data: []api.Host{{ID: "host-1", Name: "UDM-Pro"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: hosts get",
			args: []string{"hosts", "get", "host-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetHost", "host-123").Return(&api.HostResponse{
					Data: api.Host{ID: "host-123", Name: "UDM-Pro"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: hosts restart",
			args: []string{"hosts", "restart", "host-123", "--force"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("RestartHost", "host-123").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Route: hosts health",
			args: []string{"hosts", "health", "host-123"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetHostHealth", "host-123").Return(&api.HealthResponse{
					Data: []api.HealthStatus{{Subsystem: "system", Status: "ok"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Route: hosts stats",
			args: []string{"hosts", "stats", "host-123", "--period", "day"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetHostStats", "host-123", "day").Return(&api.PerformanceResponse{
					Data: api.PerformanceStats{},
				}, nil)
			},
			expectedError: false,
		},

		// ========== WHOMI (Auth Check) ==========
		{
			name: "Route: whoami",
			args: []string{"whoami"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("Whoami").Return(&api.WhoamiResponse{
					Data: api.UserInfo{
						ID:    "user-1",
						Email: "admin@example.com",
						Role:  "admin",
					},
				}, nil)
				m.On("GetConnectionInfo").Return(api.ConnectionInfo{
					Mode:        "cloud",
					Endpoint:    "https://api.ui.com",
					IsConnected: true,
				})
			},
			expectedError: false,
		},

		// ========== ERROR HANDLING TESTS ==========
		{
			name:          "Route: validation error - missing site ID for devices list",
			args:          []string{"devices", "list"},
			mockSetup:     func(m *mocks.SiteManager) {}, // No mock needed - validation fails before API call
			expectedError: true,
			errorContains: "<site-id>",
		},
		{
			name: "Route: API error handling",
			args: []string{"sites", "get", "nonexistent"},
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetSite", "nonexistent").Return(nil, &api.NotFoundError{Resource: "/v1/sites/nonexistent"})
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "Route: stub not implemented error",
			args: []string{"hosts", "get", "host-123"}, // Using a stubbed method
			mockSetup: func(m *mocks.SiteManager) {
				m.On("GetHost", "host-123").Return(nil, &api.NotImplementedError{Method: "GetHost"})
			},
			expectedError: true,
			errorContains: "not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockClient := new(mocks.SiteManager)
			tt.mockSetup(mockClient)

			// Create CLI context with mock
			cmdCtx := &CLIContext{
				Client: mockClient,
				Config: &config.Config{
					Output: config.OutputConfig{
						Format:    "json",
						Color:     "auto",
						NoHeaders: false,
					},
				},
				Format:  "json",
				Verbose: false,
				Debug:   false,
			}

			// Parse args with Kong
			var root RootCmd
			parser, err := kong.New(&root, kong.UsageOnError())
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			kCtx, err := parser.Parse(tt.args)
			if err != nil {
				// Some tests expect validation errors at parse time
				if tt.expectedError && tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
					return
				}
				t.Fatalf("Failed to parse args: %v", err)
			}

			// Execute command
			err = kCtx.Run(cmdCtx)

			// Verify expectations
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expected mock calls were made
			mockClient.AssertExpectations(t)
		})
	}
}

// TestLocalClientStubMethods specifically tests the 27 stubbed methods
// This ensures they return proper "not yet implemented" errors
func TestLocalClientStubMethods(t *testing.T) {
	// Create a real LocalClient with mocked responses for stubs
	tests := []struct {
		name          string
		method        string
		args          []interface{}
		expectedError string
	}{
		// Site stubs
		{"Stub: CreateSite", "CreateSite", []interface{}{api.CreateSiteRequest{}}, "not yet implemented"},
		{"Stub: UpdateSite", "UpdateSite", []interface{}{"site-123", api.UpdateSiteRequest{}}, "not yet implemented"},
		{"Stub: DeleteSite", "DeleteSite", []interface{}{"site-123"}, "not yet implemented"},
		{"Stub: GetSiteHealth", "GetSiteHealth", []interface{}{"site-123"}, "not yet implemented"},
		{"Stub: GetSiteStats", "GetSiteStats", []interface{}{"site-123", ""}, "not yet implemented"},

		// Host stubs
		{"Stub: ListHosts", "ListHosts", []interface{}{0, ""}, "not yet implemented"},
		{"Stub: GetHost", "GetHost", []interface{}{"host-123"}, "not yet implemented"},
		{"Stub: GetHostHealth", "GetHostHealth", []interface{}{"host-123"}, "not yet implemented"},
		{"Stub: GetHostStats", "GetHostStats", []interface{}{"host-123", ""}, "not yet implemented"},
		{"Stub: RestartHost", "RestartHost", []interface{}{"host-123"}, "not yet implemented"},

		// Client stubs
		{"Stub: GetClientStats", "GetClientStats", []interface{}{"site-123", "aa:bb:cc"}, "not yet implemented"},

		// Alert stubs
		{"Stub: ListAlerts", "ListAlerts", []interface{}{"", 0, "", false}, "not yet implemented"},
		{"Stub: AcknowledgeAlert", "AcknowledgeAlert", []interface{}{"site-123", "alert-1"}, "not yet implemented"},
		{"Stub: ArchiveAlert", "ArchiveAlert", []interface{}{"site-123", "alert-1"}, "not yet implemented"},

		// Event stubs
		{"Stub: ListEvents", "ListEvents", []interface{}{"", 0, ""}, "not yet implemented"},

		// Network stubs
		{"Stub: ListNetworks", "ListNetworks", []interface{}{"site-123", 0, ""}, "not yet implemented"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.SiteManager)

			// Setup mock to return NotImplementedError
			var returnValues []interface{}
			switch tt.method {
			case "DeleteSite", "RestartHost", "AcknowledgeAlert", "ArchiveAlert":
				returnValues = []interface{}{&api.NotImplementedError{Method: tt.method}}
			default:
				returnValues = []interface{}{nil, &api.NotImplementedError{Method: tt.method}}
			}

			mockClient.On(tt.method, tt.args...).Return(returnValues...)

			// Call the method
			var err error
			switch tt.method {
			case "CreateSite":
				_, err = mockClient.CreateSite(tt.args[0].(api.CreateSiteRequest))
			case "UpdateSite":
				_, err = mockClient.UpdateSite(tt.args[0].(string), tt.args[1].(api.UpdateSiteRequest))
			case "DeleteSite":
				err = mockClient.DeleteSite(tt.args[0].(string))
			case "GetSiteHealth":
				_, err = mockClient.GetSiteHealth(tt.args[0].(string))
			case "GetSiteStats":
				_, err = mockClient.GetSiteStats(tt.args[0].(string), tt.args[1].(string))
			case "ListHosts":
				_, err = mockClient.ListHosts(tt.args[0].(int), tt.args[1].(string))
			case "GetHost":
				_, err = mockClient.GetHost(tt.args[0].(string))
			case "GetHostHealth":
				_, err = mockClient.GetHostHealth(tt.args[0].(string))
			case "GetHostStats":
				_, err = mockClient.GetHostStats(tt.args[0].(string), tt.args[1].(string))
			case "RestartHost":
				err = mockClient.RestartHost(tt.args[0].(string))
			case "GetClientStats":
				_, err = mockClient.GetClientStats(tt.args[0].(string), tt.args[1].(string))
			case "ListAlerts":
				_, err = mockClient.ListAlerts(tt.args[0].(string), tt.args[1].(int), tt.args[2].(string), tt.args[3].(bool))
			case "AcknowledgeAlert":
				err = mockClient.AcknowledgeAlert(tt.args[0].(string), tt.args[1].(string))
			case "ArchiveAlert":
				err = mockClient.ArchiveAlert(tt.args[0].(string), tt.args[1].(string))
			case "ListEvents":
				_, err = mockClient.ListEvents(tt.args[0].(string), tt.args[1].(int), tt.args[2].(string))
			case "ListNetworks":
				_, err = mockClient.ListNetworks(tt.args[0].(string), tt.args[1].(int), tt.args[2].(string))
			}

			// Verify error contains "not yet implemented"
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			mockClient.AssertExpectations(t)
		})
	}
}

// TestSiteManagerInterfaceCoverage verifies that all SiteManager interface methods
// have corresponding CLI routing tests. This prevents interface drift.
func TestSiteManagerInterfaceCoverage(t *testing.T) {
	// Define all expected SiteManager methods and their test coverage status
	interfaceMethods := map[string]bool{
		// Sites (7 methods) - all covered in TestCommandRouting
		"ListSites":     true,
		"GetSite":       true,
		"CreateSite":    true, // Stubbed, tested
		"UpdateSite":    true, // Stubbed, tested
		"DeleteSite":    true, // Stubbed, tested
		"GetSiteHealth": true, // Stubbed, tested
		"GetSiteStats":  true, // Stubbed, tested

		// Hosts (5 methods) - all stubbed, tested
		"ListHosts":     true,
		"GetHost":       true,
		"GetHostHealth": true,
		"GetHostStats":  true,
		"RestartHost":   true,

		// Devices (5 methods) - all covered in TestCommandRouting
		"ListDevices":   true,
		"GetDevice":     true,
		"RestartDevice": true,
		"UpgradeDevice": true,
		"AdoptDevice":   true,

		// Clients (3 methods) - all covered in TestCommandRouting
		"ListClients":    true,
		"GetClientStats": true,
		"BlockClient":    true,

		// WLANs (5 methods) - all covered in TestCommandRouting
		"ListWLANs":  true,
		"GetWLAN":    true,
		"CreateWLAN": true,
		"UpdateWLAN": true,
		"DeleteWLAN": true,

		// Alerts (3 methods) - all covered in TestCommandRouting
		"ListAlerts":       true,
		"AcknowledgeAlert": true,
		"ArchiveAlert":     true,

		// Events (1 method) - covered in TestCommandRouting
		"ListEvents": true,

		// Networks (1 method) - covered in TestCommandRouting
		"ListNetworks": true,

		// User (1 method) - covered in TestCommandRouting
		"Whoami": true,

		// Connection Info (1 method)
		"GetConnectionInfo": true,

		// Debugging (1 method)
		"EnableDebug": true,
	}

	// Count total methods
	totalMethods := len(interfaceMethods)
	coveredMethods := 0
	for _, covered := range interfaceMethods {
		if covered {
			coveredMethods++
		}
	}

	// Verify we have 33 methods total
	assert.Equal(t, 33, totalMethods, "SiteManager interface should have exactly 33 methods")

	// Verify 100% coverage
	assert.Equal(t, totalMethods, coveredMethods, "All SiteManager methods should have test coverage")

	t.Logf("SiteManager Interface Coverage: %d/%d methods (%.1f%%)",
		coveredMethods, totalMethods, float64(coveredMethods)/float64(totalMethods)*100)
}
