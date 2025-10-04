package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all completed Leetcode problems",
	RunE:  HandleErrors(ListFunc),
}

func ListFunc(cmd *cobra.Command, args []string) error {
	questions, err := kata.GetAllQuestionsWithStatus(cmd.Context())
	if err != nil {
		return fmt.Errorf("listing questions: %w", err)
	}

	if err := table.Render(questions, kata.Config.Tracks); err != nil {
		return fmt.Errorf("rendering questions as table: %w", err)
	}

	return nil
}
