package cmd

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Select a random problem to complete",
	RunE:  HandleErrors(QuizFunc),
}

func QuizFunc(cmd *cobra.Command, args []string) error {
	if language == "" {
		language = kata.Config.Language
	}

	question, err := kata.GetRandomQuestion(cmd.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("✗ No eligible problems to quiz on")
			fmt.Println("ℹ You need at least one attempted solution\n\tTo get started, run: 'kata get two-sum'")
			return err
		}
		return err
	}

	fmt.Println("✓ Selected a random problem from your history")

	fmt.Print(ui.RenderQuizResult(question))

	if open || kata.Config.OpenInEditor {
		if err := kata.OpenQuestionInEditor(question); err != nil {
			return fmt.Errorf("failed to open solution file in editor: %w", err)
		}
	}

	return nil
}
