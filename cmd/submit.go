package cmd

import (
	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit solutions against leetcode servers",
	RunE:  HandleErrors(SubmitFunc),
	Args:  cobra.ExactArgs(1),
}

func init() {
	submitCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func SubmitFunc(cmd *cobra.Command, args []string) error {
	name := args[0]

	if language == "" {
		language = kata.Config.LanguageName()
	}

	opts := app.AppOptions{
		Problem:  name,
		Language: language,
	}

	return kata.Submit(opts)
}
