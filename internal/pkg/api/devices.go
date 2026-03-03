package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ========== DEVICES ==========

// ListDevices retrieves devices for a site
func (c *Client) ListDevices(siteID string, pageSize int, nextToken string) (*DevicesResponse, error) {
	endpoint := fmt.Sprintf("/v1/sites/%s/devices", siteID)
	if pageSize > 0 {
		endpoint = endpoint + "?pageSize=" + strconv.Itoa(pageSize)
	}
	if nextToken != "" {
		sep := "?"
		if pageSize > 0 {
			sep = "&"
		}
		endpoint = endpoint + sep + "nextToken=" + nextToken
	}

	resp, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var result DevicesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse devices response: %w", err)
	}

	return &result, nil
}

// GetDevice retrieves a specific device by ID
func (c *Client) GetDevice(siteID, deviceID string) (*DeviceResponse, error) {
	resp, err := c.doGet(fmt.Sprintf("/v1/sites/%s/devices/%s", siteID, deviceID))
	if err != nil {
		return nil, err
	}

	var result DeviceResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse device response: %w", err)
	}

	return &result, nil
}

// RestartDevice restarts a device
func (c *Client) RestartDevice(siteID, deviceID string) error {
	body := []byte(fmt.Sprintf(`{"deviceId":"%s"}`, deviceID))
	_, err := c.doPost(fmt.Sprintf("/v1/sites/%s/devices/%s/restart", siteID, deviceID), body)
	return err
}

// UpgradeDevice upgrades device firmware
func (c *Client) UpgradeDevice(siteID, deviceID string) error {
	body := []byte(fmt.Sprintf(`{"deviceId":"%s"}`, deviceID))
	_, err := c.doPost(fmt.Sprintf("/v1/sites/%s/devices/%s/upgrade", siteID, deviceID), body)
	return err
}

// AdoptDevice adopts a new device
func (c *Client) AdoptDevice(siteID string, macAddress string) error {
	reqBody := AdoptDeviceRequest{MACAddress: macAddress}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.doPost(fmt.Sprintf("/v1/sites/%s/devices/adopt", siteID), body)
	return err
}
