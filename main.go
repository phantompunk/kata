package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kata",
	Short: "CLI for practicing Leetcode",
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download and stub a Leetcode problem",
	RunE:  DownloadFunc,
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

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configure the client",
	RunE:  ConfigFunc,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	downloadCmd.Flags().StringP("problem", "p", "", "LeetCode problem name")
	downloadCmd.Flags().StringP("language", "l", "python", "Programming language to use")
	downloadCmd.Flags().BoolP("open", "o", false, "Open problem with editor")

	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(quizCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(settingsCmd)
}
