package cmd

import (
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configure the client",
	RunE:  ConfigFunc,
}

func ConfigFunc(cmd *cobra.Command, args []string) error {
	presenter := ui.NewPresenter()
	presenter.ShowOpeningConfigFile(kata.Setting.GetPath())
	return kata.Setting.EditConfig()
}
