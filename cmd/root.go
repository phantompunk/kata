package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var kata *app.App
var kataErr error

var (
	open     bool
	language string
)

type CommandFunc func(cmd *cobra.Command, args []string) error

var rootCmd = &cobra.Command{
	Use:           "kata",
	Short:         "CLI for practicing Leetcode",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		kata, kataErr = app.New()
		displayWarnings(kata.Settings.GetWarnings())
		return kataErr
	},
}

func Execute() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Println("Received signal", sig)

		cancel()
	}()

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// Define flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

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
	case errors.Is(err, config.ErrUnsupportedLanguage):
		return "Supported languages: cpp, golang, java, python3, javascript"
	default:
		return "An unexpected error occurred. Please try again"
	}
}

func displayWarnings(warnings []string) {
	for warning := range warnings {
		ui.PrintError(warnings[warning])
	}
}
