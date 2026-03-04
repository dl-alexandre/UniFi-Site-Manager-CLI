package cli

import (
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/cli"
)

// CheckUpdateCmd wraps the update check command from internal/cli
type CheckUpdateCmd struct {
	Force bool `help:"Force check, bypassing cache" flag:"force"`
}

// Run executes the update check
func (c *CheckUpdateCmd) Run(ctx *CLIContext) error {
	// Create the internal/cli CheckUpdateCmd and run it
	internalCmd := &cli.CheckUpdateCmd{
		Force: c.Force,
	}
	formatter := ctx.getFormatter()
	return internalCmd.Run(formatter)
}
