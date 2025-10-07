package cmd

import (
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
	if !force {
		if err := kata.Session.CheckSession(cmd.Context()); err == nil {
			ui.PrintInfo("You are already logged in as " + kata.Config.Username)
			res, err := kata.Question.GetStats(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Print(ui.RenderLoginResult(kata.Config.Username, res))
			return nil
		}
	}

	if err := kata.Session.RefreshFromBrowser(); err != nil {
		return fmt.Errorf("%w: %w", app.ErrCookiesNotFound, err)
	}

	username, err := kata.Session.ValidateSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to validate session: %w", err)
	}
	ui.PrintSuccess("Authentication successful")

	res, err := kata.Question.GetStats(cmd.Context())
	if err != nil {
		return err
	}

	fmt.Print(ui.RenderLoginResult(username, res))
	return nil
}
