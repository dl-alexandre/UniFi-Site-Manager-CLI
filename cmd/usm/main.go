package main

import (
	"fmt"
	"os"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/cli"
	pkgcli "github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/cli"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version info in the cli package
	cli.Version = version
	cli.GitCommit = gitCommit
	cli.BuildTime = buildTime

	// Perform automatic update check in background (non-blocking)
	cli.AutoUpdateCheck(version)

	exitCode, err := pkgcli.Run(os.Args[1:], version, gitCommit, buildTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(exitCode)
}
