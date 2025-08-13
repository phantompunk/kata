package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func DownloadFunc(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("problem")
	if err != nil {
		return fmt.Errorf("could not read --problem flag: %w", err)
	}
	if name == "" {
		return fmt.Errorf(`missing --problem flag, e.g. "kata download --problem two-sum"`)
	}

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
