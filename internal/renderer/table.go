package renderer

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phantompunk/kata/internal/models"
)

func ProblemsTable(questions []models.Question, languages []string) error {
	columns := createColumns(languages)
	rows := createRows(questions, languages)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}

func createRows(questions []models.Question, tracks []string) []table.Row {
	var rows []table.Row
	for _, question := range questions {
		row := []string{
			fmt.Sprint(question.ID),
			question.Title,
			colorize(question.Difficulty),
		}

		for _, lang := range tracks {
			status := "❌"
			if question.LangStatus[lang] {
				status = "✅"
			}
			row = append(row, centeredStyle.Width(len(lang)).Render(status))
		}
		rows = append(rows, row)
	}
	return rows
}

func createColumns(languages []string) []table.Column {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 30},
		{Title: "Difficulty", Width: 13},
	}

	for _, lang := range languages {
		columns = append(columns, table.Column{
			Title: centeredStyle.Render(lang),
			Width: len(lang) + 2,
		})
	}
	return columns
}

func colorize(difficulty string) string {
	styles := map[string]lipgloss.Style{
		"Easy":   lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("2")), // Green
		"Medium": lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("3")), // Yellow
		"Hard":   lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("1")), // Red
	}

	if style, exists := styles[difficulty]; exists {
		return style.Render(difficulty)
	}
	return difficulty
}
