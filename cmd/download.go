package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func DownloadFunc(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf(`missing problem, try "kata get two-sum" or "kata get 1"`)
	}
	name := args[0]

	language, err := cmd.Flags().GetString("language")
	if err != nil {
		return fmt.Errorf("could not read --language flag: %w", err)
	}

	open, err := cmd.Flags().GetBool("open")
	if err != nil {
		return fmt.Errorf("could not read --open flag: %w", err)
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("could not read --force flag: %w", err)
	}

	opts := app.AppOptions{
		Problem:  name,
		Language: language,
		Open:     open,
		Force:    force,
	}

	return kata.DownloadQuestion(opts)
}
