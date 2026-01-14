package cmd

import (
	"errors"
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/phantompunk/kata/pkg/editor"
	"github.com/spf13/cobra"
)

func newQuizCmd(kata *app.App) *cobra.Command {
	var open bool
	var language string

	cmd := &cobra.Command{
		Use:     "quiz",
		Short:   "Select a random problem to complete",
		PreRunE: validateLanguagePreRun(kata, &language),
		RunE:    handleErrors(kata, quizFunc(kata, &open, &language)),
	}

	cmd.Flags().BoolVarP(&open, "open", "o", false, "Open problem with $EDITOR")
	cmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")

	return cmd
}

func quizFunc(kata *app.App, open *bool, language *string) CommandFunc {
	return func(cmd *cobra.Command, args []string) error {
		opts := app.AppOptions{
			Workspace: kata.Config.WorkspacePath(),
			Language:  *language,
			Open:      *open,
		}

		presenter := ui.NewPresenter()

		problem, err := kata.Question.GetRandomQuestion(cmd.Context(), opts)
		if err != nil {
			if errors.Is(err, app.ErrNoQuestions) {
				presenter.ShowNoEligibleProblems()
				return nil
			}

			return err
		}

		if err := presenter.ShowQuizResult(problem); err != nil {
			return err
		}

		if opts.Open || kata.Config.OpenInEditor {
			if err := editor.Open(problem.SolutionPath()); err != nil {
				return fmt.Errorf("failed to open solution file in editor: %w", err)
			}
		}

		return nil
	}
}
