// Package output provides output formatting for CLI results
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/mattn/go-isatty"
	"github.com/rodaine/table"
)

// Formatter handles output formatting
type Formatter struct {
	Format    string
	Color     bool
	NoHeaders bool
}

// NewFormatter creates a new output formatter
func NewFormatter(format, color string, noHeaders bool) *Formatter {
	useColor := false
	switch color {
	case "always":
		useColor = true
	case "never":
		useColor = false
	case "auto":
		useColor = isatty.IsTerminal(os.Stdout.Fd())
	}

	return &Formatter{
		Format:    format,
		Color:     useColor,
		NoHeaders: noHeaders,
	}
}

// PrintJSON outputs data as formatted JSON
func (f *Formatter) PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// ========== SITE OUTPUT ==========

// SiteData holds site information for table output
type SiteData struct {
	ID          string
	Name        string
	Description string
	HostID      string
	Status      string
}

// PrintSitesTable outputs sites in table format
func (f *Formatter) PrintSitesTable(sites []SiteData) {
	if len(sites) == 0 {
		fmt.Println("No sites found.")
		return
	}

	tbl := table.New("ID", "Name", "Description", "Host ID", "Status").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, site := range sites {
		if site.Description == "" {
			site.Description = "-"
		}
		if site.Status == "" {
			site.Status = "-"
		}
		tbl.AddRow(site.ID, site.Name, site.Description, site.HostID, site.Status)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, site := range sites {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", site.ID, site.Name, site.Description, site.HostID, site.Status)
		}
	}
}

// ========== HOST OUTPUT ==========

// HostData holds host information for table output
type HostData struct {
	ID         string
	Name       string
	Type       string
	Model      string
	Version    string
	IPAddress  string
	MACAddress string
	Status     string
	SiteID     string
	SiteName   string
	Uptime     int64
}

// PrintHostsTable outputs hosts in table format
func (f *Formatter) PrintHostsTable(hosts []HostData) {
	if len(hosts) == 0 {
		fmt.Println("No hosts found.")
		return
	}

	tbl := table.New("ID", "Name", "Type", "Model", "IP Address", "Status", "Uptime").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, host := range hosts {
		uptime := formatDuration(host.Uptime)
		if host.IPAddress == "" {
			host.IPAddress = "-"
		}
		tbl.AddRow(host.ID, host.Name, host.Type, host.Model, host.IPAddress, host.Status, uptime)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, host := range hosts {
			uptime := formatDuration(host.Uptime)
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", host.ID, host.Name, host.Type, host.Model, host.IPAddress, host.Status, uptime)
		}
	}
}

// ========== DEVICE OUTPUT ==========

// DeviceData holds device information for table output
type DeviceData struct {
	ID           string
	Name         string
	Type         string
	Model        string
	Version      string
	MACAddress   string
	IPAddress    string
	Status       string
	Adopted      bool
	Uptime       int64
	Clients      int
	Satisfaction float64
	CPUUsage     float64
	MemoryUsage  float64
}

// PrintDevicesTable outputs devices in table format
func (f *Formatter) PrintDevicesTable(devices []DeviceData) {
	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return
	}

	tbl := table.New("ID", "Name", "Type", "Model", "Status", "IP Address", "Clients", "Uptime").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, device := range devices {
		uptime := formatDuration(device.Uptime)
		if device.IPAddress == "" {
			device.IPAddress = "-"
		}
		clients := "-"
		if device.Clients > 0 {
			clients = fmt.Sprintf("%d", device.Clients)
		}
		tbl.AddRow(device.ID, device.Name, device.Type, device.Model, device.Status, device.IPAddress, clients, uptime)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, device := range devices {
			uptime := formatDuration(device.Uptime)
			if device.IPAddress == "" {
				device.IPAddress = "-"
			}
			clients := "-"
			if device.Clients > 0 {
				clients = fmt.Sprintf("%d", device.Clients)
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", device.ID, device.Name, device.Type, device.Model, device.Status, device.IPAddress, clients, uptime)
		}
	}
}

// ========== CLIENT OUTPUT ==========

// ClientData holds client information for table output
type ClientData struct {
	ID             string
	MACAddress     string
	IPAddress      string
	Hostname       string
	Name           string
	ConnectionType string
	SSID           string
	Signal         int
	Satisfaction   float64
	Uptime         int64
	IsBlocked      bool
	IsGuest        bool
}

// PrintClientsTable outputs clients in table format
func (f *Formatter) PrintClientsTable(clients []ClientData) {
	if len(clients) == 0 {
		fmt.Println("No clients found.")
		return
	}

	tbl := table.New("MAC Address", "Name/Hostname", "Connection", "IP Address", "Signal", "Satisfaction", "Uptime").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, client := range clients {
		name := client.Hostname
		if name == "" {
			name = client.Name
		}
		if name == "" {
			name = "-"
		}

		signal := "-"
		if client.Signal != 0 {
			signal = fmt.Sprintf("%d dBm", client.Signal)
		}

		satisfaction := "-"
		if client.Satisfaction > 0 {
			satisfaction = fmt.Sprintf("%.0f%%", client.Satisfaction)
		}

		uptime := formatDuration(client.Uptime)
		if client.IPAddress == "" {
			client.IPAddress = "-"
		}

		tbl.AddRow(client.MACAddress, name, client.ConnectionType, client.IPAddress, signal, satisfaction, uptime)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, client := range clients {
			name := client.Hostname
			if name == "" {
				name = client.Name
			}
			if name == "" {
				name = "-"
			}
			signal := "-"
			if client.Signal != 0 {
				signal = fmt.Sprintf("%d dBm", client.Signal)
			}
			satisfaction := "-"
			if client.Satisfaction > 0 {
				satisfaction = fmt.Sprintf("%.0f%%", client.Satisfaction)
			}
			uptime := formatDuration(client.Uptime)
			if client.IPAddress == "" {
				client.IPAddress = "-"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", client.MACAddress, name, client.ConnectionType, client.IPAddress, signal, satisfaction, uptime)
		}
	}
}

// ========== WLAN OUTPUT ==========

// WLANData holds WLAN information for table output
type WLANData struct {
	ID       string
	Name     string
	SSID     string
	Security string
	Enabled  bool
	Hidden   bool
	VLAN     int
	Band     string
}

// PrintWLANsTable outputs WLANs in table format
func (f *Formatter) PrintWLANsTable(wlans []WLANData) {
	if len(wlans) == 0 {
		fmt.Println("No WLANs found.")
		return
	}

	tbl := table.New("ID", "Name", "SSID", "Security", "Enabled", "Hidden", "VLAN", "Band").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, wlan := range wlans {
		enabled := "No"
		if wlan.Enabled {
			enabled = "Yes"
		}
		hidden := "No"
		if wlan.Hidden {
			hidden = "Yes"
		}
		vlan := "-"
		if wlan.VLAN > 0 {
			vlan = fmt.Sprintf("%d", wlan.VLAN)
		}
		if wlan.Band == "" {
			wlan.Band = "both"
		}

		tbl.AddRow(wlan.ID, wlan.Name, wlan.SSID, wlan.Security, enabled, hidden, vlan, wlan.Band)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, wlan := range wlans {
			enabled := "No"
			if wlan.Enabled {
				enabled = "Yes"
			}
			hidden := "No"
			if wlan.Hidden {
				hidden = "Yes"
			}
			vlan := "-"
			if wlan.VLAN > 0 {
				vlan = fmt.Sprintf("%d", wlan.VLAN)
			}
			if wlan.Band == "" {
				wlan.Band = "both"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", wlan.ID, wlan.Name, wlan.SSID, wlan.Security, enabled, hidden, vlan, wlan.Band)
		}
	}
}

// ========== ALERT OUTPUT ==========

// AlertData holds alert information for table output
type AlertData struct {
	ID           string
	Type         string
	Message      string
	SiteID       string
	DeviceID     string
	DeviceName   string
	Severity     string
	Timestamp    string
	Acknowledged bool
	Archived     bool
}

// PrintAlertsTable outputs alerts in table format
func (f *Formatter) PrintAlertsTable(alerts []AlertData) {
	if len(alerts) == 0 {
		fmt.Println("No alerts found.")
		return
	}

	tbl := table.New("Timestamp", "Severity", "Type", "Message", "Device").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, alert := range alerts {
		device := alert.DeviceName
		if device == "" {
			device = alert.DeviceID
		}
		if device == "" {
			device = "-"
		}

		msg := alert.Message
		if len(msg) > 50 {
			msg = msg[:47] + "..."
		}

		timestamp := formatTimestamp(alert.Timestamp)
		tbl.AddRow(timestamp, alert.Severity, alert.Type, msg, device)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, alert := range alerts {
			device := alert.DeviceName
			if device == "" {
				device = alert.DeviceID
			}
			if device == "" {
				device = "-"
			}
			msg := alert.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}
			timestamp := formatTimestamp(alert.Timestamp)
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", timestamp, alert.Severity, alert.Type, msg, device)
		}
	}
}

// ========== EVENT OUTPUT ==========

// EventData holds event information for table output
type EventData struct {
	ID        string
	Type      string
	Message   string
	SiteID    string
	DeviceID  string
	ClientID  string
	UserID    string
	Timestamp string
}

// PrintEventsTable outputs events in table format
func (f *Formatter) PrintEventsTable(events []EventData) {
	if len(events) == 0 {
		fmt.Println("No events found.")
		return
	}

	tbl := table.New("Timestamp", "Type", "Message").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, event := range events {
		msg := event.Message
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}

		timestamp := formatTimestamp(event.Timestamp)
		tbl.AddRow(timestamp, event.Type, msg)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, event := range events {
			msg := event.Message
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			timestamp := formatTimestamp(event.Timestamp)
			fmt.Printf("%s\t%s\t%s\n", timestamp, event.Type, msg)
		}
	}
}

// ========== NETWORK OUTPUT ==========

// NetworkData holds network information for table output
type NetworkData struct {
	ID           string
	Name         string
	Purpose      string
	VLAN         int
	Subnet       string
	GatewayIP    string
	DHCPEnabled  bool
	DHCPStart    string
	DHCPStop     string
	NetworkGroup string
	DomainName   string
}

// PrintNetworksTable outputs networks in table format
func (f *Formatter) PrintNetworksTable(networks []NetworkData) {
	if len(networks) == 0 {
		fmt.Println("No networks found.")
		return
	}

	tbl := table.New("ID", "Name", "Purpose", "VLAN", "Subnet", "Gateway", "DHCP").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, network := range networks {
		vlan := "-"
		if network.VLAN > 0 {
			vlan = fmt.Sprintf("%d", network.VLAN)
		}
		dhcp := "No"
		if network.DHCPEnabled {
			dhcp = "Yes"
		}
		if network.Purpose == "" {
			network.Purpose = "-"
		}
		if network.Subnet == "" {
			network.Subnet = "-"
		}
		if network.GatewayIP == "" {
			network.GatewayIP = "-"
		}

		tbl.AddRow(network.ID, network.Name, network.Purpose, vlan, network.Subnet, network.GatewayIP, dhcp)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, network := range networks {
			vlan := "-"
			if network.VLAN > 0 {
				vlan = fmt.Sprintf("%d", network.VLAN)
			}
			dhcp := "No"
			if network.DHCPEnabled {
				dhcp = "Yes"
			}
			if network.Purpose == "" {
				network.Purpose = "-"
			}
			if network.Subnet == "" {
				network.Subnet = "-"
			}
			if network.GatewayIP == "" {
				network.GatewayIP = "-"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", network.ID, network.Name, network.Purpose, vlan, network.Subnet, network.GatewayIP, dhcp)
		}
	}
}

// ========== HEALTH OUTPUT ==========

// PrintHealthTable outputs health status in table format
func (f *Formatter) PrintHealthTable(health []api.HealthStatus) {
	if len(health) == 0 {
		fmt.Println("No health data available.")
		return
	}

	tbl := table.New("Subsystem", "Status", "Devices", "Clients", "Latency").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, h := range health {
		devices := "-"
		if h.NumAdopted > 0 || h.NumPending > 0 {
			devices = fmt.Sprintf("%d/%d", h.NumAdopted, h.NumAdopted+h.NumPending)
		}

		clients := "-"
		if h.NumClient > 0 {
			clients = fmt.Sprintf("%d", h.NumClient)
		}

		latency := "-"
		if h.Latency > 0 {
			latency = fmt.Sprintf("%d ms", h.Latency)
		}

		tbl.AddRow(h.Subsystem, h.Status, devices, clients, latency)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, h := range health {
			devices := "-"
			if h.NumAdopted > 0 || h.NumPending > 0 {
				devices = fmt.Sprintf("%d/%d", h.NumAdopted, h.NumAdopted+h.NumPending)
			}
			clients := "-"
			if h.NumClient > 0 {
				clients = fmt.Sprintf("%d", h.NumClient)
			}
			latency := "-"
			if h.Latency > 0 {
				latency = fmt.Sprintf("%d ms", h.Latency)
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", h.Subsystem, h.Status, devices, clients, latency)
		}
	}
}

// ========== PERFORMANCE OUTPUT ==========

// PrintPerformanceTable outputs performance stats in table format
func (f *Formatter) PrintPerformanceTable(stats *api.PerformanceStats) {
	if stats == nil {
		fmt.Println("No performance data available.")
		return
	}

	tbl := table.New("Metric", "Value").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	period := stats.Period
	if period == "" {
		period = "current"
	}

	tbl.AddRow("Site ID", stats.SiteID)
	tbl.AddRow("Period", period)
	tbl.AddRow("Rx Bytes", formatBytes(stats.RxBytes))
	tbl.AddRow("Tx Bytes", formatBytes(stats.TxBytes))

	if stats.RxRate > 0 {
		tbl.AddRow("Rx Rate", fmt.Sprintf("%.2f Mbps", stats.RxRate))
	}
	if stats.TxRate > 0 {
		tbl.AddRow("Tx Rate", fmt.Sprintf("%.2f Mbps", stats.TxRate))
	}
	if stats.Latency > 0 {
		tbl.AddRow("Latency", fmt.Sprintf("%.2f ms", stats.Latency))
	}
	if stats.PacketLoss > 0 {
		tbl.AddRow("Packet Loss", fmt.Sprintf("%.2f%%", stats.PacketLoss))
	}
	if stats.NumClients > 0 {
		tbl.AddRow("Clients", fmt.Sprintf("%d", stats.NumClients))
	}
	if stats.NumDevices > 0 {
		tbl.AddRow("Devices", fmt.Sprintf("%d", stats.NumDevices))
	}
	if stats.CPUUsage > 0 {
		tbl.AddRow("CPU Usage", fmt.Sprintf("%.1f%%", stats.CPUUsage))
	}
	if stats.MemoryUsage > 0 {
		tbl.AddRow("Memory Usage", fmt.Sprintf("%.1f%%", stats.MemoryUsage))
	}

	tbl.Print()
}

// ========== USER OUTPUT ==========

// UserData holds user information for table output
type UserData struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Role      string
	IsOwner   bool
}

// PrintUserTable outputs user info in table format
func (f *Formatter) PrintUserTable(user UserData) {
	ownerStr := "No"
	if user.IsOwner {
		ownerStr = "Yes"
	}

	tbl := table.New("Property", "Value").WithWriter(os.Stdout)

	if f.Color {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	tbl.AddRow("ID", user.ID)
	tbl.AddRow("Email", user.Email)
	tbl.AddRow("Name", user.FirstName+" "+user.LastName)
	tbl.AddRow("Role", user.Role)
	tbl.AddRow("Owner", ownerStr)

	tbl.Print()
}

// ========== UTILITY FUNCTIONS ==========

// ValidateFormat checks if the format is supported
func ValidateFormat(format string) error {
	switch format {
	case "json", "table":
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, table)", format)
	}
}

// PrintVersion outputs version information
func PrintVersion(version, gitCommit, buildTime string, checkLatest bool) {
	fmt.Printf("usm version %s\n", version)

	if version != "dev" && gitCommit != "unknown" {
		fmt.Printf("  commit: %s\n", gitCommit)
	}

	if buildTime != "unknown" {
		fmt.Printf("  built:  %s\n", buildTime)
	}

	if checkLatest {
		fmt.Println("\nChecking for updates...")
		fmt.Println("  (update check not yet implemented)")
	}
}

// PrintInitSuccess outputs a success message after init
func PrintInitSuccess(configPath string) {
	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set your API key:")
	fmt.Println("     export USM_API_KEY=your-api-key")
	fmt.Println("\n  2. Verify your setup:")
	fmt.Println("     usm whoami")
	fmt.Println("\n  3. List your sites:")
	fmt.Println("     usm sites list")
	fmt.Println("\n  4. List your hosts:")
	fmt.Println("     usm hosts list")
}

// formatDuration formats seconds into a human-readable duration
func formatDuration(seconds int64) string {
	if seconds == 0 {
		return "-"
	}

	duration := time.Duration(seconds) * time.Second

	if duration < time.Minute {
		return fmt.Sprintf("%ds", seconds)
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}

	days := int(duration.Hours()) / 24
	return fmt.Sprintf("%dd", days)
}

// formatTimestamp formats an ISO timestamp to a more readable format
func formatTimestamp(ts string) string {
	if ts == "" {
		return "-"
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}

	return t.Format("2006-01-02 15:04")
}

// formatBytes formats bytes into a human-readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
