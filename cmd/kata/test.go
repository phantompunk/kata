package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func TestFunc(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf(`missing problem, try "kata get two-sum" or "kata get 1"`)
	}
	name := args[0]

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
