package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoldenFileContracts verifies that all JSON files in testdata/ can be
// unmarshaled into their corresponding API structs. This is a "Contract Test"
// that ensures our unmarshaling logic handles real device responses.
//
// When users report issues with specific firmware versions, paste their --debug
// output into testdata/<device>-<version>.json and this test will verify if
// our structs handle that response correctly.
func TestGoldenFileContracts(t *testing.T) {
	testdataDir := "testdata"

	// Find all golden files
	entries, err := os.ReadDir(testdataDir)
	if os.IsNotExist(err) {
		t.Skip("testdata/ directory does not exist, skipping contract tests")
	}
	require.NoError(t, err, "failed to read testdata directory")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(testdataDir, entry.Name())
			data, err := os.ReadFile(path)
			require.NoError(t, err, "failed to read golden file")

			// Determine contract type from filename prefix
			// Format: <type>-<device>-<version>.json
			// Examples:
			//   devices-udm-pro-v8.1.113.json
			//   clients-udr-v3.2.12.json
			//   wlans-udm-se-v8.0.7.json
			parts := strings.SplitN(entry.Name(), "-", 2)
			contractType := parts[0]

			switch contractType {
			case "devices":
				testDeviceContract(t, data)
			case "device":
				testDeviceContract(t, data)
			case "clients":
				testClientContract(t, data)
			case "client":
				testClientContract(t, data)
			case "wlans":
				testWLANContract(t, data)
			case "wlan":
				testWLANContract(t, data)
			case "sites":
				testSiteContract(t, data)
			case "site":
				testSiteContract(t, data)
			case "hosts":
				testHostContract(t, data)
			case "host":
				testHostContract(t, data)
			case "networks":
				testNetworkContract(t, data)
			case "network":
				testNetworkContract(t, data)
			default:
				t.Logf("Unknown contract type '%s' in filename, skipping validation", contractType)
			}
		})
	}
}

func testDeviceContract(t *testing.T, data []byte) {
	var response struct {
		Data []Device `json:"data"`
	}
	err := json.Unmarshal(data, &response)
	assert.NoError(t, err, "failed to unmarshal device response")

	if err == nil {
		t.Logf("✓ Device contract valid: %d devices parsed", len(response.Data))
		if len(response.Data) > 0 {
			device := response.Data[0]
			t.Logf("  Sample: ID=%s, Name=%s, Type=%s, Status=%s",
				device.ID, device.Name, device.Type, device.Status)
		}
	}
}

func testClientContract(t *testing.T, data []byte) {
	var response struct {
		Data []Client `json:"data"`
	}
	err := json.Unmarshal(data, &response)
	assert.NoError(t, err, "failed to unmarshal client response")

	if err == nil {
		t.Logf("✓ Client contract valid: %d clients parsed", len(response.Data))
	}
}

func testWLANContract(t *testing.T, data []byte) {
	var response struct {
		Data []WLAN `json:"data"`
	}
	err := json.Unmarshal(data, &response)
	assert.NoError(t, err, "failed to unmarshal WLAN response")

	if err == nil {
		t.Logf("✓ WLAN contract valid: %d WLANs parsed", len(response.Data))
	}
}

func testSiteContract(t *testing.T, data []byte) {
	var response struct {
		Data []Site `json:"data"`
	}
	err := json.Unmarshal(data, &response)
	assert.NoError(t, err, "failed to unmarshal site response")

	if err == nil {
		t.Logf("✓ Site contract valid: %d sites parsed", len(response.Data))
	}
}

func testHostContract(t *testing.T, data []byte) {
	var response struct {
		Data []Host `json:"data"`
	}
	err := json.Unmarshal(data, &response)
	assert.NoError(t, err, "failed to unmarshal host response")

	if err == nil {
		t.Logf("✓ Host contract valid: %d hosts parsed", len(response.Data))
	}
}

func testNetworkContract(t *testing.T, data []byte) {
	var response struct {
		Data []Network `json:"data"`
	}
	err := json.Unmarshal(data, &response)
	assert.NoError(t, err, "failed to unmarshal network response")

	if err == nil {
		t.Logf("✓ Network contract valid: %d networks parsed", len(response.Data))
	}
}

// TestDeviceUnmarshalCustomFields specifically tests our custom unmarshaling
// for fields that differ between Cloud and Local API responses
func TestDeviceUnmarshalCustomFields(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Device
	}{
		{
			name: "Cloud API response",
			json: `{
				"_id": "dev-123",
				"name": "Access Point",
				"type": "uap",
				"adopted": true,
				"status": "online"
			}`,
			expected: Device{
				ID:      "dev-123",
				Name:    "Access Point",
				Type:    "uap",
				Adopted: true,
				Status:  "online",
			},
		},
		{
			name: "Local API response (state field mapped to status)",
			json: `{
				"_id": "dev-456",
				"name": "Switch",
				"type": "usw",
				"state": 1
			}`,
			expected: Device{
				ID:     "dev-456",
				Name:   "Switch",
				Type:   "usw",
				Status: "ONLINE",
			},
		},
		{
			name: "Local API response (state 0 = OFFLINE)",
			json: `{
				"_id": "dev-789",
				"name": "Offline Device",
				"type": "uap",
				"state": 0
			}`,
			expected: Device{
				ID:     "dev-789",
				Name:   "Offline Device",
				Type:   "uap",
				Status: "OFFLINE",
			},
		},
		{
			name: "Local API with unknown state value",
			json: `{
				"_id": "dev-abc",
				"name": "Gateway",
				"type": "ugw",
				"state": 2
			}`,
			expected: Device{
				ID:     "dev-abc",
				Name:   "Gateway",
				Type:   "ugw",
				Status: "OFFLINE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var device Device
			err := json.Unmarshal([]byte(tt.json), &device)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.ID, device.ID)
			assert.Equal(t, tt.expected.Name, device.Name)
			assert.Equal(t, tt.expected.Type, device.Type)
			assert.Equal(t, tt.expected.Adopted, device.Adopted)
			assert.Equal(t, tt.expected.Status, device.Status)
			// Status mapping is derived from state field during unmarshaling
			t.Logf("Unmarshaled: status=%s, adopted=%v", device.Status, device.Adopted)
		})
	}
}

// TestContractTestHelper provides a helper function for manually adding
// golden files from user bug reports
func TestContractTestHelper(t *testing.T) {
	// This test documents how to add new golden files
	t.Log(`
How to add a new golden file contract test:

1. When a user reports an issue with --debug output:
   usm devices list --local --debug > user-device-response.json

2. Copy the JSON response to testdata/:
   cp user-device-response.json internal/pkg/api/testdata/devices-<model>-<version>.json
   
   Examples:
   - devices-udm-pro-v8.1.113.json
   - devices-udr-v3.2.17.json
   - devices-udm-se-v8.0.24.json

3. Run contract tests:
   go test ./internal/pkg/api/... -v -run TestGoldenFileContracts

4. If it fails, fix the struct in types.go or add custom unmarshaling

5. The file now serves as a regression test for that specific firmware version
`)
}
