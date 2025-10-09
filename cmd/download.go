package cmd

import (
	"errors"
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
	downloadCmd.Flags().BoolVarP(&open, "open", "o", false, "Open problem with $EDITOR")
	downloadCmd.Flags().BoolVarP(&force, "force", "f", false, "Force download even if problem already exists")
	downloadCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func DownloadFunc(cmd *cobra.Command, args []string) error {
	problemName := app.ConvertToSlug(args[0])

	if err := validateLanguage(); err != nil {
		ui.PrintError("language %q not supported", language)
		return err
	}

	opts := app.AppOptions{
		Problem:   problemName,
		Language:  language,
		Workspace: kata.Config.WorkspacePath(),
		Open:      open,
		Force:     force,
	}

	problem, err := kata.Question.GetQuestion(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrQuestionNotFound) {
			ui.PrintError("Problem %q not found", problemName)
			return nil
		}

		return fmt.Errorf("fetching question %q: %w", opts.Problem, err)
	}

	ui.PrintSuccess(fmt.Sprintf("Fetched problem: %s", problem.Title))

	if problem.DirectoryPath.Exists() && !force {
		ui.PrintError("Problem %s already exists at:\n  %s", problem.Title, problem.DirectoryPath.DisplayPath())
		ui.Print("\nTo refresh files, run:\n  kata get two-sum --force")
		return nil
	}

	result, err := kata.Question.Stub(cmd.Context(), problem, opts)
	if err != nil {
		return fmt.Errorf("stubbing question %q: %w", opts.Problem, err)
	}
	displayRenderResults(result, problem.Slug, opts.Force)

	return nil
}

func displayRenderResults(result *render.RenderResult, slug string, force bool) {
	if result.DirectoryCreated != "" {
		ui.PrintSuccess(fmt.Sprintf("Created directory: %s", result.DirectoryCreated))
	}

	if len(result.FilesCreated) > 0 {
		ui.PrintSuccess("Generated files:")
		for _, file := range result.FilesCreated {
			ui.Print(fmt.Sprintf("  • %s", file))
		}
	}

	if len(result.FilesUpdated) > 0 {
		ui.PrintInfo("Updated files:")
		for _, file := range result.FilesUpdated {
			ui.Print(fmt.Sprintf("  • %s", file))
		}
	}

	if len(result.FilesSkipped) == 1 && result.FilesSkipped[0] == "All files" {
		ui.PrintInfo(fmt.Sprintf("Problem %s already exists\n", slug))
		if !force {
			ui.Print(fmt.Sprintf("To refresh files, run:\n  kata get %s --force\n", slug))
		}
		return
	}

	if len(result.FilesSkipped) > 0 {
		ui.PrintWarning("Skipped files:")
		for _, file := range result.FilesSkipped {
			ui.PrintInfo(fmt.Sprintf("  • %s", file))
		}
		if !force {
			ui.PrintInfo("Use --force to overwrite existing files")
		}
	}

	ui.PrintNextSteps(slug)
}
