package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/models"
	"github.com/spf13/cobra"
)

func ListFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	questions, err := kata.Questions.GetAll()
	qStr := convertQuestions(questions)

	printTable(qStr)
	return nil
}

func convertQuestions(questions []models.Question) [][]string {
	var results [][]string
	for _, task := range questions {
		hasGo, hasPython := "❌", "❌"

		for _, snippet := range task.CodeSnippets {
			if snippet.LangSlug == "golang" {
				hasGo = "✅"
			}
			if snippet.LangSlug == "python2" || snippet.LangSlug == "python3" {
				hasPython = "✅"
			}
		}

		results = append(results, []string{
			task.ID,
			task.Title,
			colorizeDifficulty(task.Difficulty),
			hasGo,
			hasPython,
		})
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
func printTable(questions [][]string) {
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)

	// Define table headers
	t := table.New().
		Border(lipgloss.DoubleBorder()).
		Headers("ID", "Name", "Difficulty", "Go", "Python").
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
