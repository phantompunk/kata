package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func SubmitFunc(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("problem")
	if err != nil {
		return fmt.Errorf("could not read --problem flag: %w", err)
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil {
		return fmt.Errorf("could not read --language flag: %w", err)
	}

	opts := app.AppOptions{
		Problem:  name,
		Language: language,
	}

	return kata.Submit(opts)
}
