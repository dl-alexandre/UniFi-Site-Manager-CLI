package cli

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// DevicesCmd is the parent command for device operations
type DevicesCmd struct {
	List    ListDevicesCmd   `cmd:"" help:"List all devices"`
	Get     GetDeviceCmd     `cmd:"" help:"Get a specific device"`
	Restart RestartDeviceCmd `cmd:"" help:"Restart a device"`
	Upgrade UpgradeDeviceCmd `cmd:"" help:"Upgrade device firmware"`
	Adopt   AdoptDeviceCmd   `cmd:"" help:"Adopt a new device"`
}

// ListDevicesCmd handles listing devices
type ListDevicesCmd struct {
	SiteID   string `arg:"" help:"Site ID to list devices for"`
	PageSize int    `help:"Number of devices per page (0 = fetch all)" default:"50"`
	Search   string `help:"Filter devices by name or MAC"`
}

func (c *ListDevicesCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}

	var allDevices []api.Device
	nextToken := ""

	for {
		resp, err := ctx.Client.ListDevices(c.SiteID, c.PageSize, nextToken)
		if err != nil {
			return err
		}

		allDevices = append(allDevices, resp.Data...)

		if c.PageSize == 0 {
			nextToken = resp.NextToken
			if nextToken == "" {
				break
			}
		} else {
			break
		}
	}

	if c.Search != "" {
		filtered := make([]api.Device, 0)
		searchLower := strings.ToLower(c.Search)
		for _, device := range allDevices {
			if strings.Contains(strings.ToLower(device.Name), searchLower) ||
				strings.Contains(strings.ToLower(device.MACAddress), searchLower) {
				filtered = append(filtered, device)
			}
		}
		allDevices = filtered
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(allDevices)
	}

	deviceData := make([]output.DeviceData, len(allDevices))
	for i, device := range allDevices {
		deviceData[i] = output.DeviceData{
			ID:           device.ID,
			Name:         device.Name,
			Type:         device.Type,
			Model:        device.Model,
			Version:      device.Version,
			MACAddress:   device.MACAddress,
			IPAddress:    device.IPAddress,
			Status:       device.Status,
			Adopted:      device.Adopted,
			Uptime:       device.Uptime,
			Clients:      device.Clients,
			Satisfaction: device.Satisfaction,
			CPUUsage:     device.CPUUsage,
			MemoryUsage:  device.MemoryUsage,
		}
	}

	formatter.PrintDevicesTable(deviceData)
	return nil
}

// GetDeviceCmd handles getting a specific device
type GetDeviceCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	DeviceID string `arg:"" help:"Device ID to retrieve"`
}

func (c *GetDeviceCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.DeviceID == "" {
		return &api.ValidationError{Message: "device ID is required"}
	}

	resp, err := ctx.Client.GetDevice(c.SiteID, c.DeviceID)
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	device := resp.Data
	deviceData := []output.DeviceData{{
		ID:           device.ID,
		Name:         device.Name,
		Type:         device.Type,
		Model:        device.Model,
		Version:      device.Version,
		MACAddress:   device.MACAddress,
		IPAddress:    device.IPAddress,
		Status:       device.Status,
		Adopted:      device.Adopted,
		Uptime:       device.Uptime,
		Clients:      device.Clients,
		Satisfaction: device.Satisfaction,
		CPUUsage:     device.CPUUsage,
		MemoryUsage:  device.MemoryUsage,
	}}

	formatter.PrintDevicesTable(deviceData)
	return nil
}

// RestartDeviceCmd handles restarting a device
type RestartDeviceCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	DeviceID string `arg:"" help:"Device ID to restart"`
	Force    bool   `help:"Skip confirmation"`
}

func (c *RestartDeviceCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.DeviceID == "" {
		return &api.ValidationError{Message: "device ID is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to restart device %s? (y/N): ", c.DeviceID)
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restart cancelled")
			return nil
		}
	}

	if err := ctx.Client.RestartDevice(c.SiteID, c.DeviceID); err != nil {
		return err
	}

	fmt.Printf("Device %s restart initiated\n", c.DeviceID)
	return nil
}

// UpgradeDeviceCmd handles upgrading device firmware
type UpgradeDeviceCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	DeviceID string `arg:"" help:"Device ID to upgrade"`
	Force    bool   `help:"Skip confirmation"`
}

func (c *UpgradeDeviceCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.DeviceID == "" {
		return &api.ValidationError{Message: "device ID is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to upgrade firmware on device %s? (y/N): ", c.DeviceID)
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Upgrade cancelled")
			return nil
		}
	}

	if err := ctx.Client.UpgradeDevice(c.SiteID, c.DeviceID); err != nil {
		return err
	}

	fmt.Printf("Device %s firmware upgrade initiated\n", c.DeviceID)
	return nil
}

// AdoptDeviceCmd handles adopting a new device
type AdoptDeviceCmd struct {
	SiteID     string `arg:"" help:"Site ID to adopt device into"`
	MACAddress string `arg:"" help:"MAC address of device to adopt"`
}

func (c *AdoptDeviceCmd) Run(ctx *CLIContext) error {
	if c.SiteID == "" {
		return &api.ValidationError{Message: "site ID is required"}
	}
	if c.MACAddress == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	if err := ctx.Client.AdoptDevice(c.SiteID, c.MACAddress); err != nil {
		return err
	}

	fmt.Printf("Device %s adoption initiated in site %s\n", c.MACAddress, c.SiteID)
	return nil
}
