package cmd

import (
	"fmt"
	"os"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/render/templates"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "get",
	Short: "Download and stub a Leetcode problem",
	RunE:  HandleErrors(DownloadFunc),
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

	question, err := kata.GetQuestion(opts.Problem, opts.Language, opts.Force)
	if err != nil {
		return fmt.Errorf("fetching question %q: %w", opts.Problem, err)
	}
	ui.PrintSuccess(fmt.Sprintf("Fetched problem: %s", question.Title))

	if err := kata.Stub(question, opts); err != nil {
		return fmt.Errorf("stubbing problem %q: %w", opts.Problem, err)
	}

	if err := kata.Renderer.RenderOutput(os.Stdout, templates.CliGet, question); err != nil {
		return fmt.Errorf("rendering next steps: %w", err)
	}

	return nil
}
