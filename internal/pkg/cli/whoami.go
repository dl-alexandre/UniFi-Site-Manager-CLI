package cli

import (
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/output"
)

// WhoamiCmd handles showing authenticated user information
type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(ctx *CLIContext) error {
	resp, err := ctx.Client.Whoami()
	if err != nil {
		return err
	}

	formatter := ctx.getFormatter()

	if ctx.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	user := resp.Data
	userData := output.UserData{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		IsOwner:   user.IsOwner,
	}

	formatter.PrintUserTable(userData)
	return nil
}
