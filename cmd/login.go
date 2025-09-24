package cmd

import (
	"context"
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var force bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Accept session and token, attempt to get user info",
	RunE:  HandleErrors(LoginFunc),
}

func init() {
	loginCmd.Flags().BoolVarP(&force, "force", "f", false, "Always refresh browser cookies")
}

func LoginFunc(cmd *cobra.Command, args []string) error {
	if !kata.Config.IsSessionValid() {
		if err := kata.RefreshCookies(); err != nil {
			return fmt.Errorf("%w: %w", app.ErrCookiesNotFound, err)
		}
	}

	if err := kata.ValidateSession(); err != nil {
		return fmt.Errorf("failed to validate session: %w", err)
	}

	ui.PrintSuccess("Authentication successful")
	res, err := kata.Repo.GetStats(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Print(ui.RenderLoginResult(kata.Config.Username, res))
	return nil
}
