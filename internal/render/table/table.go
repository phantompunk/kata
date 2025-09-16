package table

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phantompunk/kata/internal/models"
)

var (
	base   = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("240"))
	center = lipgloss.NewStyle().Align(lipgloss.Center)
	bold   = lipgloss.NewStyle().Bold(true)
)

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return base.Render(m.table.View()) + "\n"
}

func Render(questions []models.Question, languages []string) error {
	columns := createColumns(languages)
	rows := createRows(questions, languages)

	table := createTable(columns, rows)
	m := model{table}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}

func createTable(columns []table.Column, rows []table.Row) table.Model {
	height := min(len(rows)+1, 10)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	s := table.DefaultStyles()
	s.Selected = bold
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	t.SetStyles(s)

	return t
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
			row = append(row, center.Width(len(lang)).Render(status))
		}
		rows = append(rows, row)
	}
	return rows
}

func createColumns(languages []string) []table.Column {
	columns := []table.Column{
		{Title: center.Render("ID"), Width: 4},
		{Title: "Name", Width: 30},
		{Title: "Difficulty", Width: 14},
	}

	for _, lang := range languages {
		columns = append(columns, table.Column{
			Title: center.Render(lang),
			Width: len(lang) + 2,
		})
	}
	return columns
}

var difficultyStyles = map[string]lipgloss.Style{
	"Easy":   center.PaddingLeft(2).Foreground(lipgloss.Color("2")), // Green
	"Medium": center.PaddingLeft(1).Foreground(lipgloss.Color("3")), // Yellow
	"Hard":   center.PaddingLeft(2).Foreground(lipgloss.Color("1")), // Red
}

func colorize(difficulty string) string {
	if style, exists := difficultyStyles[difficulty]; exists {
		return style.Render(difficulty)
	}
	return difficulty
}
