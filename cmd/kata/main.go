package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/spf13/cobra"
)

var kata *app.App
var kataErr error

type CommandFunc func(cmd *cobra.Command, args []string) error

var rootCmd = &cobra.Command{
	Use:           "kata",
	Short:         "CLI for practicing Leetcode",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		kata, kataErr = app.New()
		return kataErr
	},
}

var downloadCmd = &cobra.Command{
	Use:   "get",
	Short: "Download and stub a Leetcode problem",
	RunE:  HandleErrors(DownloadFunc),
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all completed Leetcode problems",
	RunE:  HandleErrors(ListFunc),
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs problem solution against leetcode test cases",
	RunE:  HandleErrors(TestFunc),
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit solutions against leetcode servers",
	RunE:  HandleErrors(SubmitFunc),
}

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configure the client",
	RunE:  config.ConfigFunc,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Define flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	downloadCmd.Flags().StringP("language", "l", "", "Programming language to use")
	downloadCmd.Flags().BoolP("open", "o", false, "Open problem with $EDITOR")
	downloadCmd.Flags().BoolP("force", "f", false, "Force download even if problem already exists")

	testCmd.Flags().StringP("language", "l", "", "Programming language to use")
	submitCmd.Flags().StringP("language", "l", "", "Programming language to use")

	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(quizCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(submitCmd)
	rootCmd.AddCommand(settingsCmd)
}

func HandleErrors(fn CommandFunc) CommandFunc {
	return func(cmd *cobra.Command, args []string) error {
		err := fn(cmd, args)
		if err == nil {
			return nil
		}

		if v, _ := cmd.Flags().GetBool("verbose"); v || kata.Config.Verbose {
			return fmt.Errorf("Error %+v", err)
		}
		return fmt.Errorf("%s", userMessage(err))
	}
}

func userMessage(err error) string {
	switch {
	case errors.Is(err, leetcode.ErrQuestionNotFound):
		return "No matching question found. Please check the problem slug"
	case errors.Is(err, leetcode.ErrNotAuthenticated):
		return "Not auth"
	case errors.Is(err, app.ErrDuplicateProblem):
		return "Problem already exists, use --force to overwrite"
	case errors.Is(err, app.ErrCookiesNotFound):
		return "Session not found. Please sign in to https://leetcode.com then run 'kata login' again"
	case errors.Is(err, app.ErrInvalidSession):
		return "Session expired. Please sign in to https://leetcode.com then run 'kata login' again"
	case errors.Is(err, app.ErrNoQuestions):
		return "No questions found in the database. Please run `kata get` to fetch questions"
	default:
		return "An unexpected error occurred. Please try again"
	}
}
