package cli

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
)

// WifiCmd is the parent command for WiFi optimization
type WifiCmd struct {
	Optimize           WifiOptimizeCmd           `cmd:"" help:"Show current optimization status and recommendations"`
	SetBandSteering    WifiSetBandSteeringCmd    `cmd:"" help:"Enable/disable band steering"`
	SetAirtimeFairness WifiSetAirtimeFairnessCmd `cmd:"" help:"Enable/disable airtime fairness"`
	SetIOTOptimize     WifiSetIOTOptimizeCmd     `cmd:"" help:"Enable/disable IoT optimization"`
	Channels           WifiChannelsCmd           `cmd:"" help:"Show current channel assignment and recommend changes"`
}

// WifiOptimizeCmd analyzes WiFi settings and provides recommendations
type WifiOptimizeCmd struct {
	SiteID string `arg:"" help:"Site ID to analyze"`
}

func (c *WifiOptimizeCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	// Get all WLANs for the site
	wlansResp, err := ctx.Client.ListWLANs(c.SiteID, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list WLANs: %w", err)
	}

	// Get all devices to check AP channels
	devicesResp, err := ctx.Client.ListDevices(c.SiteID, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	// Get AP channel information
	apChannelsResp, err := ctx.Client.GetAPChannels(c.SiteID)
	if err != nil {
		// Non-fatal: continue without channel info
		apChannelsResp = &api.APChannelsResponse{Data: []api.APChannelInfo{}}
	}

	fmt.Println("WiFi Optimization Analysis")
	fmt.Println("==========================")
	fmt.Println()

	// Analyze each WLAN
	if len(wlansResp.Data) == 0 {
		fmt.Println("No WLANs found for this site.")
		return nil
	}

	optimizationNeeded := false
	recommendations := []string{}

	for _, wlan := range wlansResp.Data {
		fmt.Printf("WLAN: %s (SSID: %s)\n", wlan.Name, wlan.SSID)
		fmt.Println(strings.Repeat("-", 40))

		// Get detailed settings for this WLAN
		settingsResp, err := ctx.Client.GetWLANSettings(c.SiteID, wlan.ID)

		bandSteeringStatus := "disabled"
		airtimeFairnessStatus := "disabled"
		iotOptimizeStatus := "disabled"
		minRateStatus := "not configured"

		if err == nil && settingsResp != nil {
			settings := settingsResp.Data

			// Check Band Steering
			if settings.BandSteering != "" {
				bandSteeringStatus = settings.BandSteering
			}

			// Check Airtime Fairness
			if settings.AirtimeFairness {
				airtimeFairnessStatus = "enabled"
			}

			// Check IoT Optimization
			if settings.IOTOptimize {
				iotOptimizeStatus = "enabled"
			}

			// Check Minimum Data Rate
			if settings.MinimumDataRate > 0 {
				minRateStatus = fmt.Sprintf("%d kbps", settings.MinimumDataRate)
			} else if settings.MinimumDataRate2G > 0 || settings.MinimumDataRate5G > 0 {
				minRateStatus = fmt.Sprintf("2.4G: %d kbps, 5G: %d kbps", settings.MinimumDataRate2G, settings.MinimumDataRate5G)
			}
		}

		// Display current settings with status indicators
		fmt.Printf("  Band Steering:    %s %s\n", bandSteeringStatus, getStatusIcon(bandSteeringStatus != "disabled"))
		fmt.Printf("  Airtime Fairness: %s %s\n", airtimeFairnessStatus, getStatusIcon(airtimeFairnessStatus == "enabled"))
		fmt.Printf("  IoT Optimize:     %s %s\n", iotOptimizeStatus, getStatusIcon(iotOptimizeStatus == "enabled"))
		fmt.Printf("  Min Rate:         %s %s\n", minRateStatus, getStatusIcon(minRateStatus != "not configured"))
		fmt.Println()

		// Build recommendations
		if bandSteeringStatus == "disabled" {
			recommendations = append(recommendations, fmt.Sprintf("🔴 %s: Enable Band Steering (prefer_5g)", wlan.Name))
			optimizationNeeded = true
		}
		if airtimeFairnessStatus == "disabled" {
			recommendations = append(recommendations, fmt.Sprintf("🔴 %s: Enable Airtime Fairness", wlan.Name))
			optimizationNeeded = true
		}
		if iotOptimizeStatus == "disabled" {
			recommendations = append(recommendations, fmt.Sprintf("🟡 %s: Enable IoT Optimization (if you have smart devices)", wlan.Name))
		}
	}

	// Display Channel Status
	fmt.Println("Channel Status:")
	fmt.Println(strings.Repeat("-", 40))

	apsFound := false
	for _, ap := range apChannelsResp.Data {
		if ap.Name != "" {
			apsFound = true
			channel24G := "auto"
			if ap.Channel24G > 0 {
				channel24G = fmt.Sprintf("%d", ap.Channel24G)
			}

			channel5G := "auto"
			if ap.Channel5G > 0 {
				channel5G = fmt.Sprintf("%d", ap.Channel5G)
			}

			fmt.Printf("  %s (%s): Channel %s (2.4G), %s (5G)\n", ap.Name, ap.Model, channel24G, channel5G)
		}
	}

	if !apsFound {
		// Fallback: extract channel info from devices
		for _, device := range devicesResp.Data {
			if device.Type == "uap" || strings.Contains(strings.ToLower(device.Model), "ap") {
				fmt.Printf("  %s (%s): Unable to retrieve channel details\n", device.Name, device.Model)
			}
		}
	}

	// Channel recommendations
	channelRecs := analyzeChannels(apChannelsResp.Data)
	if len(channelRecs) > 0 {
		fmt.Println()
		fmt.Println("⚠️  Channel Recommendations:")
		for _, rec := range channelRecs {
			fmt.Printf("   - %s\n", rec)
		}
	}

	fmt.Println()

	// Display Recommendations
	if optimizationNeeded || len(recommendations) > 0 {
		fmt.Println("Recommendations:")
		fmt.Println(strings.Repeat("-", 40))

		if len(recommendations) == 0 && len(channelRecs) == 0 {
			fmt.Println("✓ No optimization recommendations. Your WiFi settings look good!")
		} else {
			for _, rec := range recommendations {
				fmt.Println(rec)

				// Add impact description
				if strings.Contains(rec, "Band Steering") {
					fmt.Println("   Impact: 40-50% speed increase for 5GHz-capable devices")
				} else if strings.Contains(rec, "Airtime Fairness") {
					fmt.Println("   Impact: Prevents bandwidth hogs from slowing the network")
				} else if strings.Contains(rec, "IoT Optimization") {
					fmt.Println("   Impact: Reduces smart device airtime by 30%")
				}
				fmt.Println()
			}
		}
	} else {
		fmt.Println("✓ No optimization recommendations. Your WiFi settings look good!")
	}

	fmt.Println()
	fmt.Println("To apply these optimizations:")
	fmt.Println("  usm wifi set-bandsteering <wlan-id> --mode prefer_5g")
	fmt.Println("  usm wifi set-airtime-fairness <wlan-id> --enable")
	fmt.Println("  usm wifi set-iot-optimize <wlan-id> --enable")

	return nil
}

// WifiSetBandSteeringCmd enables/disables band steering
type WifiSetBandSteeringCmd struct {
	WLANID string `arg:"" help:"WLAN ID to configure"`
	Mode   string `help:"Band steering mode: off, prefer_5g, force_5g" default:"prefer_5g" enum:"off,prefer_5g,force_5g"`
	SiteID string `help:"Site ID (required for cloud mode)"`
}

func (c *WifiSetBandSteeringCmd) Run(ctx *CLIContext) error {
	if c.WLANID == "" {
		return &api.ValidationError{Message: "WLAN ID is required"}
	}

	// For cloud mode, site ID is required
	siteID := c.SiteID
	if siteID == "" && ctx.Client.GetConnectionInfo().Mode == "cloud" {
		return &api.ValidationError{Message: "site ID is required (--site)"}
	}

	// For local mode, use "default" site if not specified
	if siteID == "" {
		siteID = "default"
	}

	settings := map[string]interface{}{
		"bandsteering": c.Mode,
	}

	if err := ctx.Client.UpdateWLANSettings(siteID, c.WLANID, settings); err != nil {
		return fmt.Errorf("failed to update band steering: %w", err)
	}

	fmt.Printf("✓ Band steering set to '%s' for WLAN %s\n", c.Mode, c.WLANID)

	if c.Mode == "prefer_5g" {
		fmt.Println("  This will encourage dual-band devices to use 5GHz for better performance.")
	} else if c.Mode == "force_5g" {
		fmt.Println("  Warning: force_5g may prevent 2.4GHz-only devices from connecting.")
	}

	return nil
}

// WifiSetAirtimeFairnessCmd enables/disables airtime fairness
type WifiSetAirtimeFairnessCmd struct {
	WLANID string `arg:"" help:"WLAN ID to configure"`
	Enable bool   `help:"Enable airtime fairness" default:"true"`
	SiteID string `help:"Site ID (required for cloud mode)"`
}

func (c *WifiSetAirtimeFairnessCmd) Run(ctx *CLIContext) error {
	if c.WLANID == "" {
		return &api.ValidationError{Message: "WLAN ID is required"}
	}

	siteID := c.SiteID
	if siteID == "" && ctx.Client.GetConnectionInfo().Mode == "cloud" {
		return &api.ValidationError{Message: "site ID is required (--site)"}
	}
	if siteID == "" {
		siteID = "default"
	}

	settings := map[string]interface{}{
		"atf_enabled": c.Enable,
	}

	if err := ctx.Client.UpdateWLANSettings(siteID, c.WLANID, settings); err != nil {
		return fmt.Errorf("failed to update airtime fairness: %w", err)
	}

	if c.Enable {
		fmt.Printf("✓ Airtime fairness enabled for WLAN %s\n", c.WLANID)
		fmt.Println("  This prevents any single device from monopolizing bandwidth.")
	} else {
		fmt.Printf("✓ Airtime fairness disabled for WLAN %s\n", c.WLANID)
	}

	return nil
}

// WifiSetIOTOptimizeCmd enables/disables IoT optimization
type WifiSetIOTOptimizeCmd struct {
	WLANID string `arg:"" help:"WLAN ID to configure"`
	Enable bool   `help:"Enable IoT optimization" default:"true"`
	SiteID string `help:"Site ID (required for cloud mode)"`
}

func (c *WifiSetIOTOptimizeCmd) Run(ctx *CLIContext) error {
	if c.WLANID == "" {
		return &api.ValidationError{Message: "WLAN ID is required"}
	}

	siteID := c.SiteID
	if siteID == "" && ctx.Client.GetConnectionInfo().Mode == "cloud" {
		return &api.ValidationError{Message: "site ID is required (--site)"}
	}
	if siteID == "" {
		siteID = "default"
	}

	settings := map[string]interface{}{
		"iot_enabled": c.Enable,
	}

	if err := ctx.Client.UpdateWLANSettings(siteID, c.WLANID, settings); err != nil {
		return fmt.Errorf("failed to update IoT optimization: %w", err)
	}

	if c.Enable {
		fmt.Printf("✓ IoT optimization enabled for WLAN %s\n", c.WLANID)
		fmt.Println("  This reduces airtime used by low-bandwidth smart devices.")
	} else {
		fmt.Printf("✓ IoT optimization disabled for WLAN %s\n", c.WLANID)
	}

	return nil
}

// WifiChannelsCmd shows current channel assignment and recommendations
type WifiChannelsCmd struct {
	SiteID string `arg:"" help:"Site ID to analyze channels for"`
}

func (c *WifiChannelsCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	// Get all devices
	devicesResp, err := ctx.Client.ListDevices(c.SiteID, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	// Get AP channel information
	apChannelsResp, err := ctx.Client.GetAPChannels(c.SiteID)
	if err != nil {
		apChannelsResp = &api.APChannelsResponse{Data: []api.APChannelInfo{}}
	}

	fmt.Println("WiFi Channel Analysis")
	fmt.Println("=====================")
	fmt.Println()

	// Find all APs
	aps := []api.Device{}
	for _, device := range devicesResp.Data {
		if device.Type == "uap" || strings.Contains(strings.ToLower(device.Model), "ap") {
			aps = append(aps, device)
		}
	}

	if len(aps) == 0 {
		fmt.Println("No access points found for this site.")
		return nil
	}

	fmt.Println("Current Channel Assignment:")
	fmt.Println(strings.Repeat("-", 60))

	// Build a map of AP name to channel info
	apChannelMap := make(map[string]api.APChannelInfo)
	for _, ap := range apChannelsResp.Data {
		apChannelMap[ap.Name] = ap
	}

	for _, ap := range aps {
		channelInfo, exists := apChannelMap[ap.Name]
		if !exists {
			// Try matching by ID if name doesn't match
			for _, ci := range apChannelsResp.Data {
				if ci.ID == ap.ID {
					channelInfo = ci
					exists = true
					break
				}
			}
		}

		if exists {
			channel24G := "auto"
			if channelInfo.Channel24G > 0 {
				channel24G = fmt.Sprintf("%d", channelInfo.Channel24G)
			}

			channel5G := "auto"
			if channelInfo.Channel5G > 0 {
				channel5G = fmt.Sprintf("%d", channelInfo.Channel5G)
			}

			fmt.Printf("  %s (%s):\n", ap.Name, ap.Model)
			fmt.Printf("    2.4GHz: %s\n", channel24G)
			fmt.Printf("    5GHz:   %s\n", channel5G)
		} else {
			fmt.Printf("  %s (%s): Channel info unavailable\n", ap.Name, ap.Model)
		}
	}

	// Analyze and recommend channels
	recommendations := analyzeChannels(apChannelsResp.Data)

	fmt.Println()
	if len(recommendations) > 0 {
		fmt.Println("Recommendations:")
		fmt.Println(strings.Repeat("-", 60))
		for _, rec := range recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	} else {
		fmt.Println("✓ Channel configuration looks good!")
	}

	// General channel advice
	fmt.Println()
	fmt.Println("General Channel Guidelines:")
	fmt.Println("  2.4GHz: Use channels 1, 6, or 11 only (non-overlapping)")
	fmt.Println("  5GHz:   Use channels 36, 40, 44, 48 (lower UNII-1 band)")
	fmt.Println("          or 149, 153, 157, 161, 165 (UNII-3 band)")
	fmt.Println("  DFS channels (52-144) may cause disconnections if radar is detected")

	return nil
}

// Helper functions

func getStatusIcon(enabled bool) string {
	if enabled {
		return "✓"
	}
	return "❌"
}

func analyzeChannels(apChannels []api.APChannelInfo) []string {
	recommendations := []string{}

	if len(apChannels) < 2 {
		return recommendations
	}

	// Check for overlapping 2.4GHz channels
	channels24G := make(map[int][]string)
	for _, ap := range apChannels {
		if ap.Channel24G > 0 {
			channels24G[ap.Channel24G] = append(channels24G[ap.Channel24G], ap.Name)
		}
	}

	// Non-overlapping 2.4GHz channels are 1, 6, 11
	nonOverlapping := map[int]bool{1: true, 6: true, 11: true}

	for channel, aps := range channels24G {
		if len(aps) > 1 {
			recommendations = append(recommendations,
				fmt.Sprintf("Multiple APs on 2.4GHz channel %d: %s", channel, strings.Join(aps, ", ")))
		}

		if !nonOverlapping[channel] && channel > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("AP on overlapping 2.4GHz channel %d - consider moving to 1, 6, or 11", channel))
		}
	}

	// Check for auto channel usage on 5GHz
	auto5GCount := 0
	for _, ap := range apChannels {
		if ap.Channel5G == 0 {
			auto5GCount++
		}
	}

	if auto5GCount > 0 {
		recommendations = append(recommendations,
			"Some APs are using auto channel selection for 5GHz - consider using fixed channels 36, 40, 44, 48")
	}

	return recommendations
}
