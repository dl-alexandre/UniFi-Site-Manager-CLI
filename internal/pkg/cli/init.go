package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/config"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// InitCmd handles interactive configuration setup
type InitCmd struct {
	Force bool `help:"Overwrite existing config"`
}

func (c *InitCmd) Run(ctx *CLIContext) error {
	if config.ConfigExists() && !c.Force {
		return fmt.Errorf("config already exists. Use --force to overwrite")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("UniFi Site Manager CLI - Configuration Setup")
	fmt.Println("==========================================")

	fmt.Print("Base URL [https://api.ui.com]: ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://api.ui.com"
	}

	fmt.Print("Default output format [table]: ")
	format, _ := reader.ReadString('\n')
	format = strings.TrimSpace(format)
	if format == "" {
		format = "table"
	}
	if err := output.ValidateFormat(format); err != nil {
		return err
	}

	fmt.Print("Color mode [auto]: ")
	color, _ := reader.ReadString('\n')
	color = strings.TrimSpace(color)
	if color == "" {
		color = "auto"
	}

	cfg := &config.Config{
		API: config.APIConfig{
			BaseURL: baseURL,
			Timeout: 30,
		},
		Output: config.OutputConfig{
			Format:    format,
			Color:     color,
			NoHeaders: false,
		},
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	output.PrintInitSuccess(config.GetConfigFilePath())
	return nil
}
