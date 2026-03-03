// Package api provides HTTP client for UniFi Site Manager API
package api

// ConnectionInfo contains metadata about the current API connection
type ConnectionInfo struct {
	Mode        string // "cloud" or "local"
	Endpoint    string // API endpoint URL (redacted for security)
	Version     string // Controller/API version (if available)
	SiteID      string // Default or current site ID
	IsConnected bool   // Whether connection is active
}

// SiteManager defines all interactions with a UniFi controller (Cloud or Local).
// This interface abstracts the underlying implementation, allowing CLI commands
// to work with either cloud or local controllers without modification.
type SiteManager interface {
	// ========== Sites ==========
	// ListSites returns all sites with pagination support
	ListSites(pageSize int, nextToken string) (*SitesResponse, error)
	// GetSite returns a single site by ID
	GetSite(siteID string) (*SiteResponse, error)
	// CreateSite creates a new site
	CreateSite(req CreateSiteRequest) (*SiteResponse, error)
	// UpdateSite updates an existing site
	UpdateSite(siteID string, req UpdateSiteRequest) (*SiteResponse, error)
	// DeleteSite deletes a site by ID
	DeleteSite(siteID string) error
	// GetSiteHealth returns health status for a site
	GetSiteHealth(siteID string) (*HealthResponse, error)
	// GetSiteStats returns performance statistics for a site
	GetSiteStats(siteID string, period string) (*PerformanceResponse, error)

	// ========== Hosts ==========
	// ListHosts returns all hosts/consoles with pagination support
	ListHosts(pageSize int, nextToken string) (*HostsResponse, error)
	// GetHost returns a single host by ID
	GetHost(hostID string) (*HostResponse, error)
	// GetHostHealth returns health status for a host
	GetHostHealth(hostID string) (*HealthResponse, error)
	// GetHostStats returns performance statistics for a host
	GetHostStats(hostID string, period string) (*PerformanceResponse, error)
	// RestartHost restarts a host/console
	RestartHost(hostID string) error

	// ========== Devices ==========
	// ListDevices returns all devices for a site with pagination support
	ListDevices(siteID string, pageSize int, nextToken string) (*DevicesResponse, error)
	// GetDevice returns a single device by ID within a site
	GetDevice(siteID, deviceID string) (*DeviceResponse, error)
	// RestartDevice restarts a device within a site
	RestartDevice(siteID, deviceID string) error
	// UpgradeDevice upgrades firmware for a device
	UpgradeDevice(siteID, deviceID string) error
	// AdoptDevice adopts a device using its MAC address
	AdoptDevice(siteID string, macAddress string) error

	// ========== Clients ==========
	// ListClients returns all clients for a site with pagination and filtering
	ListClients(siteID string, pageSize int, nextToken string, wiredOnly, wirelessOnly bool) (*ClientsResponse, error)
	// GetClientStats returns statistics for a specific client by MAC address
	GetClientStats(siteID, macAddress string) (*SingleResponse[ClientStats], error)
	// BlockClient blocks or unblocks a client by MAC address
	BlockClient(siteID, macAddress string, block bool) error

	// ========== WLANs ==========
	// ListWLANs returns all wireless networks for a site
	ListWLANs(siteID string, pageSize int, nextToken string) (*WLANsResponse, error)
	// GetWLAN returns a single WLAN by ID
	GetWLAN(siteID, wlanID string) (*WLANResponse, error)
	// CreateWLAN creates a new wireless network
	CreateWLAN(siteID string, req CreateWLANRequest) (*WLANResponse, error)
	// UpdateWLAN updates an existing wireless network
	UpdateWLAN(siteID, wlanID string, req UpdateWLANRequest) (*WLANResponse, error)
	// DeleteWLAN deletes a wireless network
	DeleteWLAN(siteID, wlanID string) error

	// ========== Alerts ==========
	// ListAlerts returns all alerts for a site with filtering
	ListAlerts(siteID string, pageSize int, nextToken string, archived bool) (*AlertsResponse, error)
	// AcknowledgeAlert marks an alert as acknowledged
	AcknowledgeAlert(siteID, alertID string) error
	// ArchiveAlert marks an alert as archived
	ArchiveAlert(siteID, alertID string) error

	// ========== Events ==========
	// ListEvents returns all events for a site
	ListEvents(siteID string, pageSize int, nextToken string) (*EventsResponse, error)

	// ========== Networks ==========
	// ListNetworks returns all networks for a site
	ListNetworks(siteID string, pageSize int, nextToken string) (*NetworksResponse, error)
	// EnableNetwork enables a network by ID
	EnableNetwork(siteID, networkID string) error
	// DisableNetwork disables a network by ID
	DisableNetwork(siteID, networkID string) error

	// ========== User ==========
	// Whoami returns information about the authenticated user
	Whoami() (*WhoamiResponse, error)

	// ========== Connection Info ==========
	// GetConnectionInfo returns metadata about the current connection
	// Useful for diagnostics and the whoami command
	GetConnectionInfo() ConnectionInfo

	// ========== Debugging ==========
	// EnableDebug enables verbose API logging with credential redaction
	EnableDebug()
}

// Compile-time check to ensure CloudClient fully implements SiteManager interface
var _ SiteManager = (*Client)(nil)
