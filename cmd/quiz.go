package cmd

import (
	"errors"
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Select a random problem to complete",
	RunE:  HandleErrors(QuizFunc),
}

func init() {
	quizCmd.Flags().BoolVarP(&open, "open", "o", false, "Open problem with $EDITOR")
	quizCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func QuizFunc(cmd *cobra.Command, args []string) error {
	if err := validateLanguage(); err != nil {
		ui.PrintError("language %q not supported", language)
		return err
	}

	opts := app.AppOptions{
		Workspace: kata.Config.WorkspacePath(),
		Language:  language,
		Open:      open,
	}

	problem, err := kata.Download.GetRandomQuestion(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrNoQuestions) {
			ui.PrintError("No eligible problems to quiz on")
			ui.PrintInfo("ℹ You need at least one attempted solution\n\tTo get started, run: 'kata get two-sum'")
			return nil
		}

		return err
	}

	ui.PrintSuccess("Selected a random problem from your history")
	if err := ui.RenderQuizResult(problem); err != nil {
		return err
	}

	if opts.Open || kata.Config.OpenInEditor {
		if err := kata.OpenQuestionInEditor(problem); err != nil {
			return fmt.Errorf("failed to open solution file in editor: %w", err)
		}
	}

	return nil
}
