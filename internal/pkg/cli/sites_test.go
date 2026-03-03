package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/config"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test helper to create CLI context with mock client
func createTestContext(mockClient *mocks.SiteManager) *CLIContext {
	return &CLIContext{
		Client: mockClient,
		Config: &config.Config{
			Output: config.OutputConfig{
				Format: "table",
				Color:  "never",
			},
		},
		Format:  "table",
		Color:   "never",
		Verbose: false,
		Debug:   false,
	}
}

// Test helper to capture stdout
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// ========== SitesExecCmd Tests ==========

func TestSitesExecCmd_AllSites_Flag(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", HostID: "host-1"},
			{ID: "site-2", Name: "Home", HostID: "host-2"},
		},
	}

	devicesResp := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "AP1", Type: "uap", Status: "ONLINE"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devicesResp, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devicesResp, nil)

	cmd := &SitesExecCmd{
		Command:  "devices list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Site: Office")
	assert.Contains(t, output, "Site: Home")
	assert.Contains(t, output, "2/2 sites succeeded")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Sites_List_Parsing(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office", HostID: "host-1"},
	}

	site2 := api.SiteResponse{
		Data: api.Site{ID: "site-2", Name: "Home", HostID: "host-2"},
	}

	devicesResp := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "AP1", Type: "uap", Status: "ONLINE"},
		},
	}

	mockClient.On("GetSite", "site-1").Return(&site1, nil)
	mockClient.On("GetSite", "site-2").Return(&site2, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devicesResp, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devicesResp, nil)

	cmd := &SitesExecCmd{
		Command: "devices list",
		Sites:   []string{"site-1", "site-2"},
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Site: Office")
	assert.Contains(t, output, "Site: Home")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Error_One_Site_Continues(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", HostID: "host-1"},
			{ID: "site-2", Name: "Home", HostID: "host-2"},
		},
	}

	devicesResp := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "AP1", Type: "uap", Status: "ONLINE"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devicesResp, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(nil, fmt.Errorf("connection timeout"))

	cmd := &SitesExecCmd{
		Command:  "devices list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Site: Office")
	assert.Contains(t, output, "Site: Home")
	assert.Contains(t, output, "ERROR: connection timeout")
	assert.Contains(t, output, "1/2 sites succeeded")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_No_Sites_Flag_Error(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &SitesExecCmd{
		Command: "devices list",
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either --sites or --all-sites must be specified")
}

func TestSitesExecCmd_Empty_Site_List(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)

	cmd := &SitesExecCmd{
		Command:  "devices list",
		AllSites: true,
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid sites")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Clients_Command(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", HostID: "host-1"},
		},
	}

	clientsResp := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "client-1", Name: "Laptop", IPAddress: "192.168.1.100", IsWired: true},
			{ID: "client-2", Hostname: "Phone", IPAddress: "192.168.1.101", IsWired: false},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clientsResp, nil)

	cmd := &SitesExecCmd{
		Command:  "clients list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Laptop")
	assert.Contains(t, output, "Phone")
	assert.Contains(t, output, "wired")
	assert.Contains(t, output, "wireless")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Alerts_Command(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", HostID: "host-1"},
		},
	}

	alertsResp := api.AlertsResponse{
		Data: []api.Alert{
			{ID: "alert-1", Type: "DISCONNECT", Severity: "warning", Message: "Device offline"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&alertsResp, nil)

	cmd := &SitesExecCmd{
		Command:  "alerts list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "warning")
	assert.Contains(t, output, "Device offline")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Unsupported_Command(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", HostID: "host-1"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)

	cmd := &SitesExecCmd{
		Command:  "invalid command",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "unsupported command type")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Invalid_Site_ID(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("GetSite", "invalid-id").Return(nil, &api.NotFoundError{Resource: "site"})

	cmd := &SitesExecCmd{
		Command: "devices list",
		Sites:   []string{"invalid-id"},
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid sites")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Permission_Error(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("GetSite", "restricted-site").Return(nil, &api.PermissionError{Message: "access denied"})

	cmd := &SitesExecCmd{
		Command: "devices list",
		Sites:   []string{"restricted-site"},
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid sites")
	mockClient.AssertExpectations(t)
}

// ========== SitesCompareCmd Tests ==========

func TestSitesCompareCmd_Two_Sites_Comparison(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office"},
	}

	site2 := api.SiteResponse{
		Data: api.Site{ID: "site-2", Name: "Home"},
	}

	devices1 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "AP1", Type: "uap", Status: "ONLINE", Model: "U6-Lite", Version: "6.5.0"},
			{ID: "dev-2", Name: "Switch1", Type: "usw", Status: "ONLINE", Model: "USW-Pro", Version: "6.4.0"},
		},
	}

	devices2 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-3", Name: "AP2", Type: "uap", Status: "ONLINE", Model: "U6-Lite", Version: "6.5.0"},
		},
	}

	clients1 := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "c1", Name: "Laptop1"},
			{ID: "c2", Name: "Phone1"},
		},
	}

	clients2 := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "c3", Name: "Laptop2"},
		},
	}

	mockClient.On("GetSite", "site-1").Return(&site1, nil)
	mockClient.On("GetSite", "site-2").Return(&site2, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices1, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devices2, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients1, nil)
	mockClient.On("ListClients", "site-2", 0, "", false, false).Return(&clients2, nil)

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1", "site-2"},
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Site Comparison")
	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "Home")
	assert.Contains(t, output, "2 devices")
	assert.Contains(t, output, "1 devices")
	assert.Contains(t, output, "2 clients")
	assert.Contains(t, output, "1 clients")
	mockClient.AssertExpectations(t)
}

func TestSitesCompareCmd_Firmware_Mismatch(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office"},
	}

	site2 := api.SiteResponse{
		Data: api.Site{ID: "site-2", Name: "Home"},
	}

	devices1 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "AP1", Type: "uap", Status: "ONLINE", Model: "U6-Lite", Version: "6.5.0"},
		},
	}

	devices2 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-2", Name: "AP2", Type: "uap", Status: "ONLINE", Model: "U6-Lite", Version: "6.4.0"},
		},
	}

	clients1 := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	clients2 := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	mockClient.On("GetSite", "site-1").Return(&site1, nil)
	mockClient.On("GetSite", "site-2").Return(&site2, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices1, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devices2, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients1, nil)
	mockClient.On("ListClients", "site-2", 0, "", false, false).Return(&clients2, nil)

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1", "site-2"},
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "MISMATCH")
	assert.Contains(t, output, "U6-Lite")
	mockClient.AssertExpectations(t)
}

func TestSitesCompareCmd_Identical_Configs(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office"},
	}

	site2 := api.SiteResponse{
		Data: api.Site{ID: "site-2", Name: "Home"},
	}

	// Devices with same model but identical firmware versions
	devices1 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "AP1", Type: "uap", Status: "ONLINE", Model: "U6-Lite", Version: "6.5.0"},
		},
	}

	devices2 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-2", Name: "AP2", Type: "uap", Status: "ONLINE", Model: "U6-Lite", Version: "6.5.0"},
		},
	}

	clients1 := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	clients2 := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	mockClient.On("GetSite", "site-1").Return(&site1, nil)
	mockClient.On("GetSite", "site-2").Return(&site2, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices1, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devices2, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients1, nil)
	mockClient.On("ListClients", "site-2", 0, "", false, false).Return(&clients2, nil)

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1", "site-2"},
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	// When all firmware versions match, should show Firmware Versions section
	assert.Contains(t, output, "Firmware Versions")
	mockClient.AssertExpectations(t)
}

func TestSitesCompareCmd_Three_Plus_Sites(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office"},
	}

	site2 := api.SiteResponse{
		Data: api.Site{ID: "site-2", Name: "Home"},
	}

	site3 := api.SiteResponse{
		Data: api.Site{ID: "site-3", Name: "Remote"},
	}

	emptyDevices := api.DevicesResponse{
		Data: []api.Device{},
	}

	emptyClients := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	mockClient.On("GetSite", "site-1").Return(&site1, nil)
	mockClient.On("GetSite", "site-2").Return(&site2, nil)
	mockClient.On("GetSite", "site-3").Return(&site3, nil)
	mockClient.On("ListDevices", mock.Anything, 0, "").Return(&emptyDevices, nil).Times(3)
	mockClient.On("ListClients", mock.Anything, 0, "", false, false).Return(&emptyClients, nil).Times(3)

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1", "site-2", "site-3"},
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "Home")
	assert.Contains(t, output, "Remote")
	mockClient.AssertExpectations(t)
}

func TestSitesCompareCmd_Requires_Two_Sites(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1"},
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 sites required")
}

func TestSitesCompareCmd_Site_Fetch_Error(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("GetSite", "site-1").Return(nil, fmt.Errorf("network error"))

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1", "site-2"},
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get site")
}

func TestSitesCompareCmd_JSON_Output(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office"},
	}

	site2 := api.SiteResponse{
		Data: api.Site{ID: "site-2", Name: "Home"},
	}

	emptyDevices := api.DevicesResponse{
		Data: []api.Device{},
	}

	emptyClients := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	mockClient.On("GetSite", "site-1").Return(&site1, nil)
	mockClient.On("GetSite", "site-2").Return(&site2, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&emptyDevices, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&emptyDevices, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&emptyClients, nil)
	mockClient.On("ListClients", "site-2", 0, "", false, false).Return(&emptyClients, nil)

	cmd := &SitesCompareCmd{
		Sites:  []string{"site-1", "site-2"},
		Output: "json",
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "site_id")
	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "Home")
	mockClient.AssertExpectations(t)
}

// ========== SitesReportCmd Tests ==========

func TestSitesReportCmd_Aggregation_Totals(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", HostID: "host-1"},
			{ID: "site-2", Name: "Home", HostID: "host-2"},
		},
	}

	devices1 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Status: "ONLINE"},
			{ID: "dev-2", Status: "ONLINE"},
		},
	}

	devices2 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-3", Status: "ONLINE"},
		},
	}

	clients1 := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "c1"},
			{ID: "c2"},
		},
	}

	clients2 := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "c3"},
			{ID: "c4"},
			{ID: "c5"},
		},
	}

	alerts1 := api.AlertsResponse{Data: []api.Alert{}}
	alerts2 := api.AlertsResponse{Data: []api.Alert{}}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices1, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devices2, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients1, nil)
	mockClient.On("ListClients", "site-2", 0, "", false, false).Return(&clients2, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&alerts1, nil)
	mockClient.On("ListAlerts", "site-2", 0, "", false).Return(&alerts2, nil)

	cmd := &SitesReportCmd{}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Sites monitored: 2")
	assert.Contains(t, output, "Total devices:")
	assert.Contains(t, output, "Total clients:")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_Per_Site_Breakdown(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	devices := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Status: "ONLINE"},
			{ID: "dev-2", Status: "OFFLINE"},
		},
	}

	clients := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "c1"},
		},
	}

	alerts := api.AlertsResponse{Data: []api.Alert{}}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&alerts, nil)

	cmd := &SitesReportCmd{}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "2 devices")
	assert.Contains(t, output, "1 clients")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_Health_Issues_Detection(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	devices := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Status: "ONLINE"},
			{ID: "dev-2", Status: "OFFLINE"},
			{ID: "dev-3", Status: "OFFLINE"},
		},
	}

	clients := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	alerts := api.AlertsResponse{
		Data: []api.Alert{
			{ID: "alert-1", Acknowledged: false},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&alerts, nil)

	cmd := &SitesReportCmd{}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "1 site(s) have issues")
	assert.Contains(t, output, "Offline devices:")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_JSON_Output(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	devices := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Status: "ONLINE"},
		},
	}

	clients := api.ClientsResponse{
		Data: []api.NetworkClient{},
	}

	alerts := api.AlertsResponse{Data: []api.Alert{}}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&alerts, nil)

	cmd := &SitesReportCmd{Output: "json"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "total_sites")
	assert.Contains(t, output, "total_devices")
	assert.Contains(t, output, "Office")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_Zero_Devices_Clients(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Empty"},
		},
	}

	emptyDevices := api.DevicesResponse{Data: []api.Device{}}
	emptyClients := api.ClientsResponse{Data: []api.NetworkClient{}}
	emptyAlerts := api.AlertsResponse{Data: []api.Alert{}}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&emptyDevices, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&emptyClients, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&emptyAlerts, nil)

	cmd := &SitesReportCmd{}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Empty")
	assert.Contains(t, output, "Total devices:")
	assert.Contains(t, output, "Total clients:")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_API_Unreachable_For_Device(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
			{ID: "site-2", Name: "Home"},
		},
	}

	devices1 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Status: "ONLINE"},
		},
	}

	emptyClients := api.ClientsResponse{Data: []api.NetworkClient{}}
	emptyAlerts := api.AlertsResponse{Data: []api.Alert{}}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices1, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(nil, fmt.Errorf("connection timeout"))
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&emptyClients, nil)
	mockClient.On("ListClients", "site-2", 0, "", false, false).Return(&emptyClients, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&emptyAlerts, nil)
	mockClient.On("ListAlerts", "site-2", 0, "", false).Return(nil, fmt.Errorf("unreachable"))

	cmd := &SitesReportCmd{}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Sites monitored: 2")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_No_Sites_Found(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)

	cmd := &SitesReportCmd{}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no sites found")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_Table_Formatting(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	devices := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Status: "ONLINE"},
		},
	}

	clients := api.ClientsResponse{
		Data: []api.NetworkClient{
			{ID: "c1"},
			{ID: "c2"},
		},
	}

	alerts := api.AlertsResponse{Data: []api.Alert{}}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices, nil)
	mockClient.On("ListClients", "site-1", 0, "", false, false).Return(&clients, nil)
	mockClient.On("ListAlerts", "site-1", 0, "", false).Return(&alerts, nil)

	cmd := &SitesReportCmd{Output: "table"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Multi-Site Report")
	assert.Contains(t, output, "Summary:")
	assert.Contains(t, output, "Per-Site Breakdown:")
	mockClient.AssertExpectations(t)
}

// ========== Additional Exec Tests ==========

func TestSitesExecCmd_Networks_Command(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	networksResp := api.NetworksResponse{
		Data: []api.Network{
			{ID: "net-1", Name: "LAN", VLAN: 1, Subnet: "192.168.1.0/24"},
			{ID: "net-2", Name: "Guest", VLAN: 10, Subnet: "192.168.10.0/24"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListNetworks", "site-1", 0, "").Return(&networksResp, nil)

	cmd := &SitesExecCmd{
		Command:  "networks list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "LAN")
	assert.Contains(t, output, "Guest")
	assert.Contains(t, output, "VLAN")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_WLANs_Command(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	wlansResp := api.WLANsResponse{
		Data: []api.WLAN{
			{ID: "wlan-1", Name: "Corporate", SSID: "CorpWiFi", Enabled: true},
			{ID: "wlan-2", Name: "Guest", SSID: "GuestWiFi", Enabled: false},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListWLANs", "site-1", 0, "").Return(&wlansResp, nil)

	cmd := &SitesExecCmd{
		Command:  "wlans list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "CorpWiFi")
	assert.Contains(t, output, "GuestWiFi")
	assert.Contains(t, output, "enabled")
	assert.Contains(t, output, "disabled")
	mockClient.AssertExpectations(t)
}

// ========== Edge Cases and Error Scenarios ==========

func TestSitesExecCmd_Empty_Command(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)

	cmd := &SitesExecCmd{
		Command:  "",
		AllSites: true,
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command cannot be empty")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_All_Sites_Fetch_Error(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("ListSites", 0, "").Return(nil, fmt.Errorf("API unavailable"))

	cmd := &SitesExecCmd{
		Command:  "devices list",
		AllSites: true,
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list sites")
	mockClient.AssertExpectations(t)
}

func TestSitesReportCmd_ListSites_Error(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("ListSites", 0, "").Return(nil, &api.NetworkError{Message: "connection refused"})

	cmd := &SitesReportCmd{}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list sites")
	mockClient.AssertExpectations(t)
}

func TestSitesCompareCmd_Device_List_Error(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site1 := api.SiteResponse{
		Data: api.Site{ID: "site-1", Name: "Office"},
	}

	// The code fetches site-1, then devices for site-1 (fails), so it never gets to site-2
	mockClient.On("GetSite", "site-1").Return(&site1, nil).Once()
	mockClient.On("ListDevices", "site-1", 0, "").Return(nil, fmt.Errorf("device API error")).Once()

	cmd := &SitesCompareCmd{
		Sites: []string{"site-1", "site-2"},
	}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list devices")
	mockClient.AssertExpectations(t)
}

func TestSitesExecCmd_Results_Grouped_By_Site(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office"},
			{ID: "site-2", Name: "Home"},
		},
	}

	devices1 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-1", Name: "Office-AP", Type: "uap", Status: "ONLINE"},
		},
	}

	devices2 := api.DevicesResponse{
		Data: []api.Device{
			{ID: "dev-2", Name: "Home-AP", Type: "uap", Status: "ONLINE"},
		},
	}

	mockClient.On("ListSites", 0, "").Return(&sites, nil)
	mockClient.On("ListDevices", "site-1", 0, "").Return(&devices1, nil)
	mockClient.On("ListDevices", "site-2", 0, "").Return(&devices2, nil)

	cmd := &SitesExecCmd{
		Command:  "devices list",
		AllSites: true,
	}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	lines := strings.Split(output, "\n")

	var officeIndex, homeIndex int = -1, -1
	for i, line := range lines {
		if strings.Contains(line, "Office") {
			officeIndex = i
		}
		if strings.Contains(line, "Home") {
			homeIndex = i
		}
	}

	assert.Greater(t, officeIndex, -1, "Office section not found")
	assert.Greater(t, homeIndex, -1, "Home section not found")

	// Check that site names appear in output
	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "Home")

	mockClient.AssertExpectations(t)
}
