package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
	open, err := cmd.Flags().GetBool("open")
	if err != nil {
		return fmt.Errorf("could not read --open flag: %w", err)
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil {
		return fmt.Errorf("could not read --language flag: %w", err)
	}

	opts := app.AppOptions{
		Open:     open,
		Language: language,
	}

	return kata.Quiz(opts)
}
