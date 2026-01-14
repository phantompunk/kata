package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/phantompunk/kata/internal/vcs"
	"github.com/spf13/cobra"
)

var kata *app.App
var kataErr error

var (
	open     bool
	language string
	version  string
	commit   string
	retry    bool
)

type CommandFunc func(cmd *cobra.Command, args []string) error

var rootCmd = &cobra.Command{
	Use:           "kata",
	Short:         "CLI for practicing Leetcode",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		kata, kataErr = app.New()
		displayWarnings(kata.Setting.GetWarnings())
		return kataErr
	},
}

func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return rootCmd.ExecuteContext(ctx)
}

func buildVersion() {
	version, commit = vcs.Version()
	setVersion()
}

func setVersion() {
	vt := fmt.Sprintf("%s versions %s (%s)\n", "kata", version, commit)
	rootCmd.SetVersionTemplate(vt)
	rootCmd.Version = version
}

func init() {
	buildVersion()
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
			return fmt.Errorf("error %+v", err)
		}
		return fmt.Errorf("%s", ui.FormatError(err))
	}
}

func displayWarnings(warnings []string) {
	ui.ShowWarnings(warnings)
}
