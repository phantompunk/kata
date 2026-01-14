package cmd

import (
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

func newSettingsCmd(kata *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Configure the client",
		RunE:  settingsFunc(kata),
	}

	return cmd
}

func settingsFunc(kata *app.App) CommandFunc {
	return func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter()
		presenter.ShowOpeningConfigFile(kata.Setting.GetPath())
		return kata.Setting.EditConfig()
	}
}
