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
	// kata, _ := app.New()
	// questions, err := kata.Questions.GetAll()

	goProblem := models.Question{ID: "1", Difficulty: "easy", Content: "demo", TitleSlug: "demo", Title: "Demo", CodeSnippets: []models.CodeSnippet{{Code: "func sample()", LangSlug: "go"}}}
	pyProblem := models.Question{ID: "2", Difficulty: "easy", Content: "sample", TitleSlug: "sample", Title: "Sample", CodeSnippets: []models.CodeSnippet{{Code: "def sample()", LangSlug: "python"}}}
	questions := []models.Question{goProblem, pyProblem}

	fmt.Println(setupTable(questions))

	return nil
}

func setupTable(questions []models.Question) *table.Table {
	var rows [][]string
	for _, task := range questions {
		rows = append(rows, []string{
			task.ID,
			task.Title,
			task.Difficulty,
			"y",
			"x",
		})
	}
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	headers := []string{"#", "Name", "Difficulty", "Go", "Python"}
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(headers...).
		Width(45).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			return baseStyle.Foreground(lipgloss.Color("252"))
		})
	return t
}
