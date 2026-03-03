package cli

import (
	"os"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// VersionCmd handles version information display
type VersionCmd struct {
	Check bool `help:"Check for updates"`
}

func (c *VersionCmd) Run(ctx *CLIContext) error {
	version := os.Getenv("USM_VERSION")
	gitCommit := os.Getenv("USM_COMMIT")
	buildTime := os.Getenv("USM_BUILD_TIME")

	if version == "" {
		version = "dev"
	}
	if gitCommit == "" {
		gitCommit = "unknown"
	}
	if buildTime == "" {
		buildTime = "unknown"
	}

	output.PrintVersion(version, gitCommit, buildTime, c.Check)
	return nil
}
