// Package output provides output formatting for CLI results
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/rodaine/table"
)

// Formatter handles output formatting
type Formatter struct {
	Format    string
	Color     bool
	NoHeaders bool
}

// NewFormatter creates a new output formatter
func NewFormatter(format, color string, noHeaders bool) *Formatter {
	useColor := false
	switch color {
	case "always":
		useColor = true
	case "never":
		useColor = false
	case "auto":
		useColor = isatty.IsTerminal(os.Stdout.Fd())
	}

	return &Formatter{
		Format:    format,
		Color:     useColor,
		NoHeaders: noHeaders,
	}
}

// PrintJSON outputs data as formatted JSON
func (f *Formatter) PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// SiteData holds site information for table output
type SiteData struct {
	ID          string
	Name        string
	Description string
	HostID      string
}

// PrintSitesTable outputs sites in table format
func (f *Formatter) PrintSitesTable(sites []SiteData) {
	if len(sites) == 0 {
		fmt.Println("No sites found.")
		return
	}

	tbl := table.New("ID", "Name", "Description", "Host ID").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, site := range sites {
		if site.Description == "" {
			site.Description = "-"
		}
		tbl.AddRow(site.ID, site.Name, site.Description, site.HostID)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		// When no headers, just print rows
		for _, site := range sites {
			fmt.Printf("%s\t%s\t%s\t%s\n", site.ID, site.Name, site.Description, site.HostID)
		}
	}
}

// UserData holds user information for table output
type UserData struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Role      string
	IsOwner   bool
}

// PrintUserTable outputs user info in table format
func (f *Formatter) PrintUserTable(user UserData) {
	ownerStr := "No"
	if user.IsOwner {
		ownerStr = "Yes"
	}

	tbl := table.New("Property", "Value").WithWriter(os.Stdout)

	if f.Color {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	tbl.AddRow("ID", user.ID)
	tbl.AddRow("Email", user.Email)
	tbl.AddRow("Name", user.FirstName+" "+user.LastName)
	tbl.AddRow("Role", user.Role)
	tbl.AddRow("Owner", ownerStr)

	tbl.Print()
}

// ValidateFormat checks if the format is supported
func ValidateFormat(format string) error {
	switch format {
	case "json", "table":
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, table)", format)
	}
}

// PrintVersion outputs version information
func PrintVersion(version, gitCommit, buildTime string, checkLatest bool) {
	fmt.Printf("usm version %s\n", version)

	if version != "dev" && gitCommit != "unknown" {
		fmt.Printf("  commit: %s\n", gitCommit)
	}

	if buildTime != "unknown" {
		fmt.Printf("  built:  %s\n", buildTime)
	}

	if checkLatest {
		// This will be implemented with actual GitHub API check
		fmt.Println("\nChecking for updates...")
		fmt.Println("  (update check not yet implemented in MVP)")
	}
}

// PrintInitSuccess outputs a success message after init
func PrintInitSuccess(configPath string) {
	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set your API key:")
	fmt.Println("     export USM_API_KEY=your-api-key")
	fmt.Println("\n  2. Verify your setup:")
	fmt.Println("     usm whoami")
	fmt.Println("\n  3. List your sites:")
	fmt.Println("     usm sites list")
}
