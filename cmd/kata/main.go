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
	Use:   "download",
	Short: "Download and stub a Leetcode problem",
	RunE:  HandleErrors(DownloadFunc),
}

var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Select a random problem to complete",
	RunE:  QuizFunc,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all completed Leetcode problems",
	RunE:  HandleErrors(ListFunc),
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Accept session and token, attempt to get user info",
	RunE:  HandleErrors(LoginFunc),
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
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Define flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	downloadCmd.Flags().StringP("problem", "p", "", "LeetCode problem name")
	downloadCmd.Flags().StringP("language", "l", "", "Programming language to use")
	downloadCmd.Flags().BoolP("open", "o", false, "Open problem with $EDITOR")
	downloadCmd.Flags().BoolP("force", "f", false, "Force download even if problem already exists")
	downloadCmd.MarkFlagRequired("problem")

	testCmd.Flags().StringP("problem", "p", "", "LeetCode problem name")
	testCmd.Flags().StringP("language", "l", "", "Programming language to use")

	submitCmd.Flags().StringP("problem", "p", "", "LeetCode problem name")
	submitCmd.Flags().StringP("language", "l", "", "Programming language to use")

	loginCmd.Flags().BoolP("force", "f", false, "Always refresh browser cookies")

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
		return "No matching question found. Please check the problem slug."
	// case errors.Is(err, kata.ErrInvalidFlag):
	// 	return "Invalid flag provided. Use --help for usage instructions."
	// case errors.Is(err, kata.ErrPermissionDenied):
	// 	return "Permission denied. Please check your file permissions."
	default:
		// Fallback: show generic message without internal stack traces
		return "An unexpected error occurred. Please try again."
	}
}
