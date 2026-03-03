// Package config provides configuration management for usm CLI
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	API    APIConfig    `mapstructure:"api"`
	Output OutputConfig `mapstructure:"output"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout"`
	// APIKey is NOT stored here - use environment variable or flag
}

// OutputConfig holds output-related configuration
type OutputConfig struct {
	Format    string `mapstructure:"format"`
	Color     string `mapstructure:"color"`
	NoHeaders bool   `mapstructure:"no_headers"`
}

// GlobalFlags holds CLI flag values that override config
type GlobalFlags struct {
	APIKey     string
	BaseURL    string
	Timeout    int
	Format     string
	Color      string
	NoHeaders  bool
	Verbose    bool
	Debug      bool
	ConfigFile string
}

// Load loads configuration from file, environment, and flags
// Precedence: flags > env vars > config file > defaults
func Load(flags GlobalFlags) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file if provided
	if flags.ConfigFile != "" {
		v.SetConfigFile(expandPath(flags.ConfigFile))
	} else {
		// Default config location
		configDir := getDefaultConfigDir()
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read config file (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	bindEnvVars(v)

	// Override with CLI flags
	applyFlags(v, flags)

	// Unmarshal to struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("api.base_url", "https://api.ui.com")
	v.SetDefault("api.timeout", 30)
	v.SetDefault("output.format", "table")
	v.SetDefault("output.color", "auto")
	v.SetDefault("output.no_headers", false)
}

func bindEnvVars(v *viper.Viper) {
	v.SetEnvPrefix("USM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit bindings for clarity
	_ = v.BindEnv("api.base_url", "USM_BASE_URL")
	_ = v.BindEnv("api.timeout", "USM_TIMEOUT")
}

func applyFlags(v *viper.Viper, flags GlobalFlags) {
	if flags.BaseURL != "" {
		v.Set("api.base_url", flags.BaseURL)
	}
	if flags.Timeout > 0 {
		v.Set("api.timeout", flags.Timeout)
	}
	if flags.Format != "" {
		v.Set("output.format", flags.Format)
	}
	if flags.Color != "" {
		v.Set("output.color", flags.Color)
	}
	if flags.NoHeaders {
		v.Set("output.no_headers", true)
	}
}

func getDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "usm")
}

// GetConfigFilePath returns the path to the config file
func GetConfigFilePath() string {
	return filepath.Join(getDefaultConfigDir(), "config.yaml")
}

// ConfigExists checks if a config file exists
func ConfigExists() bool {
	_, err := os.Stat(GetConfigFilePath())
	return !os.IsNotExist(err)
}

// Save saves the configuration to the default location
func (c *Config) Save() error {
	configDir := getDefaultConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := GetConfigFilePath()

	v := viper.New()
	v.SetConfigFile(configPath)

	// Set values (DO NOT include API key)
	v.Set("api.base_url", c.API.BaseURL)
	v.Set("api.timeout", c.API.Timeout)
	v.Set("output.format", c.Output.Format)
	v.Set("output.color", c.Output.Color)
	v.Set("output.no_headers", c.Output.NoHeaders)

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func expandPath(path string) string {
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	path = os.ExpandEnv(path)
	return path
}

// GetAPIKey retrieves the API key from environment or flag
func GetAPIKey(flagsAPIKey string) (string, error) {
	// Flag takes precedence
	if flagsAPIKey != "" {
		return flagsAPIKey, nil
	}

	// Check environment variable
	apiKey := os.Getenv("USM_API_KEY")
	if apiKey != "" {
		return apiKey, nil
	}

	return "", errors.New("API key required. Set USM_API_KEY environment variable or use --api-key flag")
}
