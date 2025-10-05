package cmd

import (
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs problem solution against leetcode test cases",
	RunE:  HandleErrors(TestFunc),
	Args:  cobra.ExactArgs(1),
}

func init() {
	testCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func TestFunc(cmd *cobra.Command, args []string) error {
	problem := args[0]

	if err := validateLanguage(); err != nil {
		ui.PrintError("language %q not supported", language)
		return nil
	}

	opts := app.AppOptions{
		Language: language,
		Problem:  problem,
	}

	return kata.Test(opts)
}
