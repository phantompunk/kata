package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "get",
	Short: "Download and stub a Leetcode problem",
	RunE:  HandleErrors(DownloadFunc),
	Args:  cobra.ExactArgs(1),
}

func init() {
	downloadCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
	downloadCmd.Flags().BoolVarP(&open, "open", "o", false, "Open problem with $EDITOR")
	downloadCmd.Flags().BoolVarP(&force, "force", "f", false, "Force download even if problem already exists")
}

func DownloadFunc(cmd *cobra.Command, args []string) error {
	problem := args[0]

	if language == "" {
		language = kata.Config.Language
	}

	opts := app.AppOptions{
		Problem:  app.ConvertToSlug(problem),
		Language: language,
		Open:     open,
		Force:    force,
	}

	question, err := kata.Download.GetQuestion(cmd.Context(), opts)
	if err != nil {
		return fmt.Errorf("fetching question %q: %w", opts.Problem, err)
	}
	ui.PrintSuccess(fmt.Sprintf("Fetched problem: %s", question.Title))

	result, err := kata.Download.Stub(cmd.Context(), question, opts, kata.Config.Workspace)
	printResults(result)
	ui.PrintNextSteps(question.TitleSlug)
	return nil
}

func printResults(result *render.RenderResult) {
	if result.DirectoryCreated != "" {
		ui.PrintSuccess(fmt.Sprintf("Created directory: %s", result.DirectoryCreated))
	}

	if len(result.FilesCreated) > 0 {
		ui.PrintSuccess("Generated files:")
		for _, file := range result.FilesCreated {
			ui.PrintInfo(fmt.Sprintf("  • %s", file))
		}
	}

	if len(result.FilesUpdated) > 0 {
		ui.PrintWarning("Updated files:")
		for _, file := range result.FilesUpdated {
			ui.PrintInfo(fmt.Sprintf("  • %s", file))
		}
	}
}
