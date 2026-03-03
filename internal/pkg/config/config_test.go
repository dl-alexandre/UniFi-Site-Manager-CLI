package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		flags   GlobalFlags
		wantErr bool
	}{
		{
			name:    "defaults only",
			flags:   GlobalFlags{},
			wantErr: false,
		},
		{
			name: "custom base URL",
			flags: GlobalFlags{
				BaseURL: "https://custom.ui.com",
			},
			wantErr: false,
		},
		{
			name: "custom timeout",
			flags: GlobalFlags{
				Timeout: 60,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cfg == nil {
				t.Error("Load() returned nil config")
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	// Save and restore original env var
	originalKey := os.Getenv("USM_API_KEY")
	defer os.Setenv("USM_API_KEY", originalKey)

	tests := []struct {
		name    string
		flagKey string
		envKey  string
		wantKey string
		wantErr bool
	}{
		{
			name:    "from flag",
			flagKey: "flag-api-key",
			envKey:  "env-api-key",
			wantKey: "flag-api-key",
			wantErr: false,
		},
		{
			name:    "from env",
			flagKey: "",
			envKey:  "env-api-key",
			wantKey: "env-api-key",
			wantErr: false,
		},
		{
			name:    "missing",
			flagKey: "",
			envKey:  "",
			wantKey: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("USM_API_KEY", tt.envKey)
			got, err := GetAPIKey(tt.flagKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantKey {
				t.Errorf("GetAPIKey() = %v, want %v", got, tt.wantKey)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "tilde expansion",
			path:     "~/.config/usm",
			expected: filepath.Join(home, ".config", "usm"),
		},
		{
			name:     "no expansion needed",
			path:     "/etc/usm/config",
			expected: "/etc/usm/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.expected {
				t.Errorf("expandPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}
