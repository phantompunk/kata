package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all completed Leetcode problems",
	RunE:  HandleErrors(ListFunc),
}

func ListFunc(cmd *cobra.Command, args []string) error {
	opts := app.AppOptions{
		Tracks: kata.Config.Tracks,
	}

	questions, err := kata.Question.GetAllQuestionsWithStatus(cmd.Context(), opts)
	if err != nil {
		return fmt.Errorf("listing questions: %w", err)
	}

	if err := table.Render(questions, kata.Config.Tracks); err != nil {
		return fmt.Errorf("rendering questions as table: %w", err)
	}

	return nil
}
