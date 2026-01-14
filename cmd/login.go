package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

func newLoginCmd(kata *app.App) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Accept session and token, attempt to get user info",
		RunE:  handleErrors(kata, loginFunc(kata, &force)),
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Always refresh browser cookies")

	return cmd
}

func loginFunc(kata *app.App, force *bool) CommandFunc {
	return func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter()

		if !*force {
			if err := kata.Session.CheckSession(cmd.Context()); err == nil {
				presenter.ShowAlreadyLoggedIn(kata.Config.Username)
				res, err := kata.Question.GetStats(cmd.Context())
				if err != nil {
					return err
				}
				return presenter.ShowLoginResult(kata.Config.Username, res)
			}
		}

		if err := kata.Session.RefreshFromBrowser(); err != nil {
			return fmt.Errorf("%w: %w", app.ErrCookiesNotFound, err)
		}

		username, err := kata.Session.ValidateSession(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to validate session: %w", err)
		}
		presenter.ShowAuthenticationSuccess()

		res, err := kata.Question.GetStats(cmd.Context())
		if err != nil {
			return err
		}

		return presenter.ShowLoginResult(username, res)
	}
}
