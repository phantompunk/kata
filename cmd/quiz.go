package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/phantompunk/kata/internal/editor"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var (
	open     bool
	language string
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
	if language == "" {
		language = kata.Config.Language
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	question, err := kata.Repo.GetRandom(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("✗ No eligible problems to quiz on")
			fmt.Println("ℹ You need at least one attempted solution\n\tTo get started, run: 'kata get two-sum'")
			return fmt.Errorf("no attempted problems found")
		}
		return fmt.Errorf("failed to get random question: %w", err)
	}

	fmt.Println("✓ Selected a random problem from your history")

	fmt.Print(ui.RenderQuizResult(&question))

	if open || kata.Config.OpenInEditor {
		problem := question.ToProblem(kata.Config.Workspace, language)
		if err := editor.OpenWithEditor(problem.SolutionPath); err != nil {
			return fmt.Errorf("failed to open solution file in editor: %w", err)
		}
	}

	return nil
}
