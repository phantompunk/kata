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

type CommandFunc func(cmd *cobra.Command, args []string) error

func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	kata, err := app.New()
	if err != nil {
		return err
	}

	displayWarnings(kata.Setting.GetWarnings())

	rootCmd := newRootCmd(kata)
	return rootCmd.ExecuteContext(ctx)
}

func newRootCmd(kata *app.App) *cobra.Command {
	version, commit := vcs.Version()
	vt := fmt.Sprintf("%s versions %s (%s)\n", "kata", version, commit)

	rootCmd := &cobra.Command{
		Use:           "kata",
		Short:         "CLI for practicing Leetcode",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	rootCmd.SetVersionTemplate(vt)
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(newDownloadCmd(kata))
	rootCmd.AddCommand(newQuizCmd(kata))
	rootCmd.AddCommand(newListCmd(kata))
	rootCmd.AddCommand(newLoginCmd(kata))
	rootCmd.AddCommand(newTestCmd(kata))
	rootCmd.AddCommand(newSubmitCmd(kata))
	rootCmd.AddCommand(newSettingsCmd(kata))

	return rootCmd
}

func handleErrors(kata *app.App, fn CommandFunc) CommandFunc {
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
