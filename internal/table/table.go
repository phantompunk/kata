package table

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phantompunk/kata/internal/models"
)

// Constants for default values
const (
	DefaultMaxHeight   = 10
	DefaultIDWidth     = 4
	DefaultNameWidth   = 30
	DefaultDiffWidth   = 14
	DefaultLangPadding = 2
	MinTableHeight     = 3
)

// TableConfig holds table configuration options
type TableConfig struct {
	MaxHeight   int
	IDWidth     int
	NameWidth   int
	DiffWidth   int
	LangPadding int
	Styles      *StyleConfig
}

// StyleConfig centralizes all table styling
type StyleConfig struct {
	Base       lipgloss.Style
	Center     lipgloss.Style
	Bold       lipgloss.Style
	Difficulty map[string]lipgloss.Style
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *TableConfig {
	return &TableConfig{
		MaxHeight:   DefaultMaxHeight,
		IDWidth:     DefaultIDWidth,
		NameWidth:   DefaultNameWidth,
		DiffWidth:   DefaultDiffWidth,
		LangPadding: DefaultLangPadding,
		Styles:      DefaultStyles(),
	}
}

// DefaultStyles returns the default style configuration
func DefaultStyles() *StyleConfig {
	base := lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("240"))
	center := lipgloss.NewStyle().Align(lipgloss.Center)
	bold := lipgloss.NewStyle().Bold(true)

	return &StyleConfig{
		Base:   base,
		Center: center,
		Bold:   bold,
		Difficulty: map[string]lipgloss.Style{
			"Easy":   center.PaddingLeft(2).Foreground(lipgloss.Color("2")), // Green
			"Medium": center.PaddingLeft(1).Foreground(lipgloss.Color("3")), // Yellow
			"Hard":   center.PaddingLeft(2).Foreground(lipgloss.Color("1")), // Red
		},
	}
}

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
	// Note: We use default styles here since the model doesn't have access to config
	// In a future refactor, we could pass config to the model
	baseStyle := lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("240"))
	return baseStyle.Render(m.table.View()) + "\n"
}

// Render creates and displays an interactive table with the given questions and languages
// Uses default configuration
func Render(questions []models.QuestionStat, languages []string) error {
	return RenderWithConfig(questions, languages, DefaultConfig())
}

// RenderWithConfig creates and displays an interactive table with custom configuration
func RenderWithConfig(questions []models.QuestionStat, languages []string, config *TableConfig) error {
	tableModel, err := NewTable(questions, languages, config)
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}

	return RunInteractiveTable(tableModel)
}

// NewTable creates a new table model without running the interactive program
func NewTable(questions []models.QuestionStat, languages []string, config *TableConfig) (table.Model, error) {
	if err := validateInputs(questions, languages, config); err != nil {
		return table.Model{}, err
	}

	columns := createColumns(languages, config)
	rows := createRows(questions, languages, config)

	return createTable(columns, rows, config), nil
}

// RunInteractiveTable runs the TUI program with the given table model
func RunInteractiveTable(tableModel table.Model) error {
	m := model{table: tableModel}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("running interactive table: %w", err)
	}

	return nil
}

// validateInputs checks if the provided inputs are valid
func validateInputs(questions []models.QuestionStat, languages []string, config *TableConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if len(questions) == 0 {
		return errors.New("questions slice cannot be empty")
	}

	if len(languages) == 0 {
		return errors.New("languages slice cannot be empty")
	}

	if config.MaxHeight < MinTableHeight {
		return fmt.Errorf("max height must be at least %d, got %d", MinTableHeight, config.MaxHeight)
	}

	return nil
}

func createTable(columns []table.Column, rows []table.Row, config *TableConfig) table.Model {
	height := calculateTableHeight(len(rows), config.MaxHeight)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Selected = config.Styles.Bold
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	t.SetStyles(s)

	return t
}

// calculateTableHeight determines the appropriate height for the table
func calculateTableHeight(rowCount, maxHeight int) int {
	// Add 1 for header row
	height := rowCount + 1
	if height > maxHeight {
		return maxHeight
	}
	if height < MinTableHeight {
		return MinTableHeight
	}
	return height
}

func createRows(questions []models.QuestionStat, tracks []string, config *TableConfig) []table.Row {
	var rows []table.Row
	for _, question := range questions {
		row := []string{
			fmt.Sprint(question.ID),
			truncateText(question.Title, config.NameWidth),
			colorize(question.Difficulty, config.Styles),
		}

		for _, lang := range tracks {
			status := "❌"
			if question.LangStatus[lang] {
				status = "✅"
			}
			row = append(row, config.Styles.Center.Width(len(lang)).Render(status))
		}
		rows = append(rows, row)
	}
	return rows
}

// truncateText truncates text to fit within the specified width
func truncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}

// calculateLangWidth calculates the width needed for a language column
func calculateLangWidth(lang string, padding int) int {
	return len(lang) + padding
}

func createColumns(languages []string, config *TableConfig) []table.Column {
	columns := []table.Column{
		{Title: config.Styles.Center.Render("ID"), Width: config.IDWidth},
		{Title: "Name", Width: config.NameWidth},
		{Title: "Difficulty", Width: config.DiffWidth},
	}

	for _, lang := range languages {
		langWidth := calculateLangWidth(lang, config.LangPadding)
		columns = append(columns, table.Column{
			Title: config.Styles.Center.Render(lang),
			Width: langWidth,
		})
	}
	return columns
}

// colorize applies styling to difficulty text based on its value
func colorize(difficulty string, styles *StyleConfig) string {
	if style, exists := styles.Difficulty[difficulty]; exists {
		return style.Render(difficulty)
	}
	// Return with center alignment for unknown difficulties
	return styles.Center.Render(difficulty)
}
