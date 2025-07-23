package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/phantompunk/kata/internal/models"
	"github.com/spf13/cobra"
)

func ListFunc(cmd *cobra.Command, args []string) error {

	// Display as Table or Markdown Table<D-y>
	questions, _ := kata.Questions.GetAllWithStatus(kata.Config.Tracks)
	// statuses := kata.Questions.GetStatuses

	// fmt.Println(kata.PrintStatuses())
	qStr := convertQuestions(kata.Config.Tracks, questions)

	printTable(kata.Config.Tracks, qStr)
	return nil
}

func convertQuestions(tracks []string, questions []models.Question) [][]string {
	var results [][]string
	for _, question := range questions {
		row := []string{
			fmt.Sprint(question.ID),
			question.Title,
			colorizeDifficulty(question.Difficulty),
		}

		for _, lang := range tracks {
			status := "❌"
			if question.LangStatus[lang] {
				status = "✅"
			}
			row = append(row, status)
		}
		results = append(results, row)
	}
	return results
}

func colorizeDifficulty(difficulty string) string {
	styles := map[string]lipgloss.Style{
		"Easy":   lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // Green
		"Medium": lipgloss.NewStyle().Foreground(lipgloss.Color("3")), // Yellow
		"Hard":   lipgloss.NewStyle().Foreground(lipgloss.Color("1")), // Red
	}

	if style, exists := styles[difficulty]; exists {
		return style.Render(difficulty)
	}
	return difficulty
}

// printTable displays a formatted table of questions using table.Table.
func printTable(languages []string, questions [][]string) {
	headers := []string{"ID", "Name", "Difficulty"}
	headers = append(headers, languages...)
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)

	// Define table headers
	t := table.New().
		Border(lipgloss.DoubleBorder()).
		Headers(headers...).
		Rows(questions...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return baseStyle
		})

	// Render table
	fmt.Println(t.Render())
}
