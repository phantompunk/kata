package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func TestFunc(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("problem")
	if err != nil {
		return fmt.Errorf("could not read --problem flag: %w", err)
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil {
		return fmt.Errorf("could not read --language flag: %w", err)
	}

	opts := app.AppOptions{
		Language: language,
		Problem:  name,
	}
	return kata.Test(opts)
}
