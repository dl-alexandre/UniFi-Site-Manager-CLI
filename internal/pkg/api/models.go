// Package api provides data models for UniFi Site Manager API
package api

import "encoding/json"

// APIResponse is the base response structure
type APIResponse struct {
	Code       string          `json:"code"`
	Data       json.RawMessage `json:"data"`
	HTTPStatus int             `json:"httpStatusCode"`
	TraceID    string          `json:"traceId"`
	NextToken  string          `json:"nextToken,omitempty"`
	Message    string          `json:"message,omitempty"`
}

// ListResponse is a generic paginated response
type ListResponse[T any] struct {
	Code       string `json:"code"`
	Data       []T    `json:"data"`
	HTTPStatus int    `json:"httpStatusCode"`
	TraceID    string `json:"traceId"`
	NextToken  string `json:"nextToken,omitempty"`
}

// SingleResponse is a generic single item response
type SingleResponse[T any] struct {
	Code       string `json:"code"`
	Data       T      `json:"data"`
	HTTPStatus int    `json:"httpStatusCode"`
	TraceID    string `json:"traceId"`
}

// Site represents a UniFi site
type Site struct {
	ID          string          `json:"_id"`
	Name        string          `json:"name"`
	Description string          `json:"desc"`
	HostID      string          `json:"hostId,omitempty"`
	HostName    string          `json:"hostName,omitempty"`
	Meta        SiteMeta        `json:"meta,omitempty"`
	Statistics  json.RawMessage `json:"statistics,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
	Status      string          `json:"status,omitempty"`
}

// SiteMeta contains site metadata
type SiteMeta struct {
	Name       string `json:"name,omitempty"`
	GatewayMAC string `json:"gateway_mac,omitempty"`
	NetworkID  string `json:"network_id,omitempty"`
}

// SitesResponse wraps the list of sites
type SitesResponse = ListResponse[Site]

// SiteResponse wraps a single site
type SiteResponse = SingleResponse[Site]

// CreateSiteRequest represents a request to create a site
type CreateSiteRequest struct {
	Name        string `json:"name"`
	Description string `json:"desc,omitempty"`
}

// UpdateSiteRequest represents a request to update a site
type UpdateSiteRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"desc,omitempty"`
}

// Host represents a UniFi host/console
type Host struct {
	ID             string          `json:"_id"`
	Name           string          `json:"name"`
	Description    string          `json:"desc,omitempty"`
	Type           string          `json:"type"`
	Model          string          `json:"model"`
	Version        string          `json:"version"`
	IPAddress      string          `json:"ipAddress,omitempty"`
	MACAddress     string          `json:"macAddress,omitempty"`
	Status         string          `json:"status"`
	SiteID         string          `json:"siteId,omitempty"`
	SiteName       string          `json:"siteName,omitempty"`
	FirmwareStatus string          `json:"firmwareStatus,omitempty"`
	LastSeen       string          `json:"lastSeen,omitempty"`
	Uptime         int64           `json:"uptime,omitempty"`
	Stats          json.RawMessage `json:"statistics,omitempty"`
	Health         json.RawMessage `json:"health,omitempty"`
}

// HostsResponse wraps the list of hosts
type HostsResponse = ListResponse[Host]

// HostResponse wraps a single host
type HostResponse = SingleResponse[Host]

// Device represents a UniFi network device (AP, switch, gateway, etc.)
type Device struct {
	ID             string          `json:"_id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Model          string          `json:"model"`
	Version        string          `json:"version,omitempty"`
	MACAddress     string          `json:"mac"`
	IPAddress      string          `json:"ip,omitempty"`
	Status         string          `json:"status"`
	SiteID         string          `json:"siteId"`
	HostID         string          `json:"hostId,omitempty"`
	Adopted        bool            `json:"adopted"`
	AdoptedAt      string          `json:"adoptedAt,omitempty"`
	Uptime         int64           `json:"uptime,omitempty"`
	Clients        int             `json:"numSta,omitempty"`
	Satisfaction   float64         `json:"satisfaction,omitempty"`
	CPUUsage       float64         `json:"cpu,omitempty"`
	MemoryUsage    float64         `json:"memory,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	FirmwareStatus string          `json:"firmwareStatus,omitempty"`
	LastSeen       string          `json:"lastSeen,omitempty"`
	RxBytes        int64           `json:"rxBytes,omitempty"`
	TxBytes        int64           `json:"txBytes,omitempty"`
	Stats          json.RawMessage `json:"statistics,omitempty"`
}

// UnmarshalJSON handles the structural differences between Cloud and Local API responses
// Cloud API uses: id (string), status (string)
// Local API uses: _id (MongoDB ObjectID), state (int: 1=online, 0=offline)
func (d *Device) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion
	type Alias Device

	// Define auxiliary struct with local-specific fields
	aux := &struct {
		*Alias
		LocalID    string `json:"_id"`
		LocalState *int   `json:"state"`
	}{
		Alias: (*Alias)(d),
	}

	// Unmarshal into auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Reconcile ID: Cloud uses 'id', Local uses '_id'
	if d.ID == "" && aux.LocalID != "" {
		d.ID = aux.LocalID
	}

	// Reconcile Status: Cloud uses 'status' string, Local uses 'state' int
	if d.Status == "" && aux.LocalState != nil {
		if *aux.LocalState == 1 {
			d.Status = "ONLINE"
		} else {
			d.Status = "OFFLINE"
		}
	}

	return nil
}

// DevicesResponse wraps the list of devices
type DevicesResponse = ListResponse[Device]

// DeviceResponse wraps a single device
type DeviceResponse = SingleResponse[Device]

// AdoptDeviceRequest represents a request to adopt a device
type AdoptDeviceRequest struct {
	MACAddress string `json:"mac"`
	SiteID     string `json:"siteId,omitempty"`
}

// RestartDeviceRequest represents a request to restart a device
type RestartDeviceRequest struct {
	DeviceID string `json:"deviceId"`
}

// UpgradeDeviceRequest represents a request to upgrade device firmware
type UpgradeDeviceRequest struct {
	DeviceID string `json:"deviceId"`
}

// NetworkClient represents a connected client device
type NetworkClient struct {
	ID           string  `json:"_id"`
	MACAddress   string  `json:"mac"`
	IPAddress    string  `json:"ip,omitempty"`
	Hostname     string  `json:"hostname,omitempty"`
	Name         string  `json:"name,omitempty"`
	DeviceType   string  `json:"deviceType,omitempty"`
	OSName       string  `json:"os,omitempty"`
	SiteID       string  `json:"siteId"`
	NetworkID    string  `json:"networkId,omitempty"`
	SSID         string  `json:"essid,omitempty"`
	APMAC        string  `json:"apMac,omitempty"`
	Signal       int     `json:"signal,omitempty"`
	RSSI         int     `json:"rssi,omitempty"`
	Noise        int     `json:"noise,omitempty"`
	RxRate       int64   `json:"rxRate,omitempty"`
	TxRate       int64   `json:"txRate,omitempty"`
	RxBytes      int64   `json:"rxBytes,omitempty"`
	TxBytes      int64   `json:"txBytes,omitempty"`
	Satisfaction float64 `json:"satisfaction,omitempty"`
	Uptime       int64   `json:"uptime,omitempty"`
	FirstSeen    string  `json:"firstSeen,omitempty"`
	LastSeen     string  `json:"lastSeen,omitempty"`
	IsWired      bool    `json:"isWired"`
	IsGuest      bool    `json:"isGuest,omitempty"`
	IsBlocked    bool    `json:"blocked,omitempty"`
	FixedIP      string  `json:"fixedIp,omitempty"`
	UseFixedIP   bool    `json:"useFixedIp,omitempty"`
	Note         string  `json:"note,omitempty"`
}

// UnmarshalJSON handles the structural differences between Cloud and Local API responses
// Cloud API uses: id (string), connectionType (string)
// Local API uses: _id (MongoDB ObjectID), is_wired (bool)
func (nc *NetworkClient) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion
	type Alias NetworkClient

	// Define auxiliary struct with local-specific fields
	aux := &struct {
		*Alias
		LocalID       string `json:"_id"`
		LocalIsWired  *bool  `json:"is_wired"`
		LocalHostname string `json:"hostname"`
		LocalName     string `json:"name"`
		LocalMAC      string `json:"mac"`
	}{
		Alias: (*Alias)(nc),
	}

	// Unmarshal into auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Reconcile ID: Cloud uses 'id', Local uses '_id'
	if nc.ID == "" && aux.LocalID != "" {
		nc.ID = aux.LocalID
	}

	// Reconcile Name: Local API may have name in 'name' or 'hostname' or neither
	// Priority: name > hostname > mac (fallback)
	if nc.Name == "" {
		if aux.LocalName != "" {
			nc.Name = aux.LocalName
		} else if aux.LocalHostname != "" {
			nc.Name = aux.LocalHostname
		} else if aux.LocalMAC != "" {
			nc.Name = aux.LocalMAC // Last resort: show MAC address
		}
	}

	// Reconcile IsWired: Cloud uses connectionType string, Local uses is_wired bool
	if aux.LocalIsWired != nil {
		nc.IsWired = *aux.LocalIsWired
	}

	return nil
}

// ClientsResponse wraps the list of clients
type ClientsResponse = ListResponse[NetworkClient]

// ClientStats represents client statistics
type ClientStats struct {
	MACAddress   string  `json:"mac"`
	RxBytes      int64   `json:"rxBytes"`
	TxBytes      int64   `json:"txBytes"`
	RxPackets    int64   `json:"rxPackets,omitempty"`
	TxPackets    int64   `json:"txPackets,omitempty"`
	RxErrors     int64   `json:"rxErrors,omitempty"`
	TxErrors     int64   `json:"txErrors,omitempty"`
	SignalAvg    int     `json:"signalAvg,omitempty"`
	Satisfaction float64 `json:"satisfaction,omitempty"`
}

// BlockClientRequest represents a request to block a client
type BlockClientRequest struct {
	MACAddress string `json:"mac"`
	Block      bool   `json:"block"`
}

// WLAN represents a wireless network configuration
type WLAN struct {
	ID               string   `json:"_id"`
	Name             string   `json:"name"`
	SSID             string   `json:"essid"`
	Security         string   `json:"security"`
	Password         string   `json:"password,omitempty"`
	SiteID           string   `json:"siteId"`
	Enabled          bool     `json:"enabled"`
	Hidden           bool     `json:"hideSsid,omitempty"`
	VLANEnabled      bool     `json:"vlanEnabled,omitempty"`
	VLAN             int      `json:"vlan,omitempty"`
	Band             string   `json:"band,omitempty"` // 2g, 5g, both
	PMFMode          string   `json:"pmfMode,omitempty"`
	WPA3Support      bool     `json:"wpa3Support,omitempty"`
	WPA3Transition   bool     `json:"wpa3Transition,omitempty"`
	MACFilterEnabled bool     `json:"macFilterEnabled,omitempty"`
	MACFilterPolicy  string   `json:"macFilterPolicy,omitempty"` // allow, deny
	MACFilterList    []string `json:"macFilterList,omitempty"`
	GroupRekey       int      `json:"groupRekey,omitempty"`
	BSSTransition    bool     `json:"bssTransition,omitempty"`
	ProxyARP         bool     `json:"proxyArp,omitempty"`
	UapsdEnabled     bool     `json:"uapsdEnabled,omitempty"`
	MinimumDataRate  int      `json:"minrate,omitempty"`
	ScheduleEnabled  bool     `json:"scheduleEnabled,omitempty"`
	Schedule         []string `json:"schedule,omitempty"`
	RadiusEnabled    bool     `json:"radiusEnabled,omitempty"`
	RadiusProfileID  string   `json:"radiusProfileId,omitempty"`
}

// WLANsResponse wraps the list of WLANs
type WLANsResponse = ListResponse[WLAN]

// WLANResponse wraps a single WLAN
type WLANResponse = SingleResponse[WLAN]

// CreateWLANRequest represents a request to create a WLAN
type CreateWLANRequest struct {
	Name            string `json:"name"`
	SSID            string `json:"essid"`
	Security        string `json:"security"`
	Password        string `json:"password,omitempty"`
	VLAN            int    `json:"vlan,omitempty"`
	Band            string `json:"band,omitempty"`
	Hidden          bool   `json:"hideSsid,omitempty"`
	PMFMode         string `json:"pmfMode,omitempty"`
	WPA3Support     bool   `json:"wpa3Support,omitempty"`
	MACFilter       bool   `json:"macFilterEnabled,omitempty"`
	MACFilterPolicy string `json:"macFilterPolicy,omitempty"`
}

// UpdateWLANRequest represents a request to update a WLAN
type UpdateWLANRequest struct {
	Name     string `json:"name,omitempty"`
	Security string `json:"security,omitempty"`
	Password string `json:"password,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Hidden   *bool  `json:"hideSsid,omitempty"`
	VLAN     int    `json:"vlan,omitempty"`
	Band     string `json:"band,omitempty"`
	PMFMode  string `json:"pmfMode,omitempty"`
}

// Alert represents a UniFi system alert
type Alert struct {
	ID           string `json:"_id"`
	Type         string `json:"type"`
	Message      string `json:"msg"`
	SiteID       string `json:"siteId,omitempty"`
	DeviceID     string `json:"deviceId,omitempty"`
	DeviceMAC    string `json:"deviceMac,omitempty"`
	DeviceName   string `json:"deviceName,omitempty"`
	Severity     string `json:"severity"`
	Timestamp    string `json:"timestamp"`
	Acknowledged bool   `json:"acknowledged,omitempty"`
	Archived     bool   `json:"archived,omitempty"`
	Key          string `json:"key,omitempty"`
}

// AlertsResponse wraps the list of alerts
type AlertsResponse = ListResponse[Alert]

// AlertResponse wraps a single alert
type AlertResponse = SingleResponse[Alert]

// AcknowledgeAlertRequest represents a request to acknowledge an alert
type AcknowledgeAlertRequest struct {
	AlertID string `json:"alertId"`
}

// ArchiveAlertRequest represents a request to archive an alert
type ArchiveAlertRequest struct {
	AlertID string `json:"alertId"`
}

// HealthStatus represents site health information
type HealthStatus struct {
	Subsystem   string `json:"subsystem"`
	Status      string `json:"status"`
	NumAdopted  int    `json:"numAdopted,omitempty"`
	NumDisabled int    `json:"numDisabled,omitempty"`
	NumPending  int    `json:"numPending,omitempty"`
	NumGateway  int    `json:"numGateway,omitempty"`
	NumAP       int    `json:"numAp,omitempty"`
	NumSwitch   int    `json:"numSwitch,omitempty"`
	NumClient   int    `json:"numSta,omitempty"`
	NumGuest    int    `json:"numGuest,omitempty"`
	NumIOT      int    `json:"numIot,omitempty"`
	NumUser     int    `json:"numUser,omitempty"`
	RxBytes     int64  `json:"rxBytes,omitempty"`
	TxBytes     int64  `json:"txBytes,omitempty"`
	Latency     int    `json:"latency,omitempty"`
	PacketLoss  int    `json:"packetLoss,omitempty"`
	WANIP       string `json:"wanIp,omitempty"`
	LANIP       string `json:"lanIp,omitempty"`
	GatewayName string `json:"gwName,omitempty"`
	Version     string `json:"version,omitempty"`
}

// HealthResponse wraps the health status
type HealthResponse struct {
	Code       string         `json:"code"`
	Data       []HealthStatus `json:"data"`
	HTTPStatus int            `json:"httpStatusCode"`
	TraceID    string         `json:"traceId"`
}

// PerformanceStats represents site performance statistics
type PerformanceStats struct {
	SiteID      string  `json:"siteId"`
	Timestamp   string  `json:"timestamp"`
	Period      string  `json:"period"`
	RxBytes     int64   `json:"rxBytes"`
	TxBytes     int64   `json:"txBytes"`
	RxRate      float64 `json:"rxRate,omitempty"`
	TxRate      float64 `json:"txRate,omitempty"`
	Latency     float64 `json:"latency,omitempty"`
	PacketLoss  float64 `json:"packetLoss,omitempty"`
	NumClients  int     `json:"numClients"`
	NumDevices  int     `json:"numDevices"`
	CPUUsage    float64 `json:"cpuUsage,omitempty"`
	MemoryUsage float64 `json:"memoryUsage,omitempty"`
}

// PerformanceResponse wraps performance statistics
type PerformanceResponse = SingleResponse[PerformanceStats]

// WhoamiResponse contains authenticated user information
type WhoamiResponse = SingleResponse[UserInfo]

// UserInfo represents the authenticated user
type UserInfo struct {
	ID        string   `json:"_id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      string   `json:"role"`
	IsOwner   bool     `json:"isOwner"`
	Sites     []string `json:"sites,omitempty"`
	APIKeyID  string   `json:"apiKeyId,omitempty"`
}

// Settings represents application settings
type Settings struct {
	AutoAdopt          bool `json:"autoAdopt"`
	AutoUpgrade        bool `json:"autoUpgrade"`
	AlertsEnabled      bool `json:"alertsEnabled"`
	EmailNotifications bool `json:"emailNotifications,omitempty"`
}

// SettingsResponse wraps settings
type SettingsResponse = SingleResponse[Settings]

// UpdateSettingsRequest represents a request to update settings
type UpdateSettingsRequest struct {
	AutoAdopt          *bool `json:"autoAdopt,omitempty"`
	AutoUpgrade        *bool `json:"autoUpgrade,omitempty"`
	AlertsEnabled      *bool `json:"alertsEnabled,omitempty"`
	EmailNotifications *bool `json:"emailNotifications,omitempty"`
}

// Event represents a UniFi system event
type Event struct {
	ID        string `json:"_id"`
	Type      string `json:"type"`
	Message   string `json:"msg"`
	SiteID    string `json:"siteId,omitempty"`
	DeviceID  string `json:"deviceId,omitempty"`
	ClientID  string `json:"clientId,omitempty"`
	UserID    string `json:"userId,omitempty"`
	Timestamp string `json:"timestamp"`
	Key       string `json:"key,omitempty"`
}

// EventsResponse wraps the list of events
type EventsResponse = ListResponse[Event]

// Network represents a network configuration
type Network struct {
	ID           string `json:"_id"`
	Name         string `json:"name"`
	Purpose      string `json:"purpose,omitempty"`
	SiteID       string `json:"siteId"`
	VLANEnabled  bool   `json:"vlanEnabled,omitempty"`
	VLAN         int    `json:"vlan,omitempty"`
	Subnet       string `json:"subnet,omitempty"`
	GatewayIP    string `json:"gatewayIp,omitempty"`
	DHCPStart    string `json:"dhcpdStart,omitempty"`
	DHCPStop     string `json:"dhcpdStop,omitempty"`
	DHCPEnabled  bool   `json:"dhcpdEnabled,omitempty"`
	NetworkGroup string `json:"networkgroup,omitempty"`
	DomainName   string `json:"domainName,omitempty"`
}

// NetworksResponse wraps the list of networks
type NetworksResponse = ListResponse[Network]

// NetworkResponse wraps a single network
type NetworkResponse = SingleResponse[Network]

// APChannelInfo represents channel information for an access point
type APChannelInfo struct {
	ID         string `json:"_id"`
	Name       string `json:"name"`
	Model      string `json:"model"`
	Channel24G int    `json:"channel24G,omitempty"`
	Channel5G  int    `json:"channel5G,omitempty"`
}

// APChannelsResponse wraps the list of AP channel information
type APChannelsResponse struct {
	Code       string          `json:"code"`
	Data       []APChannelInfo `json:"data"`
	HTTPStatus int             `json:"httpStatusCode"`
	TraceID    string          `json:"traceId"`
}

// WLANSettings represents detailed WLAN configuration including optimization settings
type WLANSettings struct {
	ID                string `json:"_id"`
	Name              string `json:"name"`
	BandSteering      string `json:"bandsteering,omitempty"`
	AirtimeFairness   bool   `json:"atf_enabled,omitempty"`
	IOTOptimize       bool   `json:"iot_enabled,omitempty"`
	MinimumDataRate   int    `json:"minrate,omitempty"`
	MinimumDataRate2G int    `json:"minrate_2g,omitempty"`
	MinimumDataRate5G int    `json:"minrate_5g,omitempty"`
}

// WLANSettingsResponse wraps WLAN settings
type WLANSettingsResponse = SingleResponse[WLANSettings]
