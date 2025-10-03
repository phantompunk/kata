package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/models"
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

	language, err := validateLanguage()
	if err != nil {
		return nil
	}

	opts := app.AppOptions{
		Problem:  problemName,
		Language: language,
		Open:     open,
		Force:    force,
	}

	question, err := kata.Download.GetQuestion(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrQuestionNotFound) {
			ui.PrintError("Problem %q not found", problemName)
			return nil
		}

		return fmt.Errorf("fetching question %q: %w", opts.Problem, err)
	}

	ui.PrintSuccess(fmt.Sprintf("Fetched problem: %s", question.Title))

	problem := question.ToProblem(kata.Config.WorkspacePath(), language)
	if isQuestionStubbed(problem) && !force {
		ui.PrintError("Problem %s already exists at:\n  %s", problem.TitleSlug, problem.DirPath)
		return nil
	}

	result, err := kata.Download.Stub(cmd.Context(), problem, opts)
	if err != nil {
		return fmt.Errorf("stubbing question %q: %w", opts.Problem, err)
	}
	displayRenderResults(result, question.TitleSlug, opts.Force)

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
		ui.PrintWarning("Updated files:")
		for _, file := range result.FilesUpdated {
			ui.Print(fmt.Sprintf("  • %s", file))
		}
	}

	if len(result.FilesSkipped) == 1 && result.FilesSkipped[0] == "All files" {
		ui.PrintInfo(fmt.Sprintf("Problem %s already exists\n", slug))
		if !force {
			fmt.Printf("To refresh files, run:\n  kata get %s --force\n", slug)
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

func isQuestionStubbed(problem *models.Problem) bool {
	exists, _ := PathExists(problem.DirPath)
	return exists
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}
