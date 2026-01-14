package cmd

import (
	"errors"
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:     "get",
	Short:   "Download and stub a Leetcode problem",
	PreRunE: validateLanguagePreRun,
	RunE:    HandleErrors(DownloadFunc),
	Args:    cobra.ExactArgs(1),
}

func init() {
	downloadCmd.Flags().BoolVarP(&open, "open", "o", false, "Open problem with $EDITOR")
	downloadCmd.Flags().BoolVarP(&force, "force", "f", false, "Force download even if problem already exists")
	downloadCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
	downloadCmd.Flags().BoolVarP(&retry, "retry", "r", false, "Update only the solution file using the expected template")
}

func DownloadFunc(cmd *cobra.Command, args []string) error {
	problemName := app.ConvertToSlug(args[0])
	presenter := ui.NewPresenter()

	opts := app.AppOptions{
		Problem:   problemName,
		Language:  language,
		Workspace: kata.Config.WorkspacePath(),
		Open:      open,
		Force:     force,
		Retry:     retry,
	}

	problem, err := kata.Question.GetQuestion(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrQuestionNotFound) {
			presenter.ShowProblemNotFound(problemName)
			return nil
		}

		return fmt.Errorf("fetching question %q: %w", opts.Problem, err)
	}

	presenter.ShowProblemFetched(problem.Title)

	if problem.DirectoryPath.Exists() && !force && !retry {
		presenter.ShowProblemAlreadyExists(problem.Title, problem.DirectoryPath.DisplayPath(), problem.Slug)
		return nil
	}

	if retry && !problem.DirectoryPath.Exists() {
		presenter.ShowProblemDoesNotExist(problem.Title, problem.DirectoryPath.DisplayPath(), problem.Slug)
		return nil
	}

	result, err := kata.Question.Stub(cmd.Context(), problem, opts)
	if err != nil {
		return fmt.Errorf("stubbing question %q: %w", opts.Problem, err)
	}
	presenter.ShowRenderResults(result, problem.Slug, opts.Force)

	return nil
}

