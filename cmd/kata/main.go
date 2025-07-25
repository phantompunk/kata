package main

import (
	"fmt"
	"os"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/spf13/cobra"
)

var kata *app.App
var kataErr error

var rootCmd = &cobra.Command{
	Use:   "kata",
	Short: "CLI for practicing Leetcode",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		kata, kataErr = app.New()
		return kataErr
	},
}

var downloadCmd = &cobra.Command{
	Use:           "download",
	Short:         "Download and stub a Leetcode problem",
	RunE:          DownloadFunc,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Select a random problem to complete",
	RunE:  QuizFunc,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all completed Leetcode problems",
	RunE:  ListFunc,
}

var loginCmd = &cobra.Command{
	Use:           "login",
	Short:         "Accept session and token, attempt to get user info",
	RunE:          LoginFunc,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs problem solution against leetcode test cases",
	RunE:  TestFunc,
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit solutions against leetcode servers",
	RunE:  SubmitFunc,
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
	downloadCmd.Flags().StringP("problem", "p", "", "LeetCode problem name")
	downloadCmd.Flags().StringP("language", "l", "", "Programming language to use")
	downloadCmd.Flags().BoolP("open", "o", false, "Open problem with $EDITOR")

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
