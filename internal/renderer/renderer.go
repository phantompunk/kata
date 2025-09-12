package renderer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/templates"
)

type Renderer struct {
	templ *template.Template
	Error error
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// Define the Bubble Tea model
type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func New() (*Renderer, error) {
	funcMap := template.FuncMap{
		"pascalCase": pascalCase,
	}
	templ, err := template.New("").Funcs(funcMap).ParseFS(templates.Files, "*.gohtml")
	if err != nil {
		return nil, err
	}
	return &Renderer{templ: templ}, nil
}

func (r *Renderer) Render(w io.Writer, problem *models.Problem, templateType string) error {
	var langBlock string
	if templateType == "solution" || templateType == "test" {
		sol, test := langTemplates(problem.LangSlug)
		if templateType == "solution" {
			langBlock = sol
		} else {
			langBlock = test
		}
	}

	if langBlock != "" {
		type TemplateData struct {
			Problem *models.Problem
			Code    string
		}
		var buf bytes.Buffer
		err := r.templ.ExecuteTemplate(&buf, langBlock, problem)
		if err != nil {
			return err
		}
		code := buf.String()
		return r.templ.ExecuteTemplate(w, templateType, &TemplateData{problem, code})
	}

	if templateType == "readme" {
		markdown, err := htmltomarkdown.ConvertString(problem.Content)
		if err != nil {
			return err
		}

		problem.Content = markdown
	}

	return r.templ.ExecuteTemplate(w, templateType, problem)
}

func langTemplates(lang string) (string, string) {
	switch lang {
	case "go", "golang":
		return "golang", "gotest"
	case "python", "python3":
		return "python", "pytest"
	default:
		return lang, lang
	}
}

func pascalCase(s string) string {
	var result strings.Builder
	nextUpper := true
	for _, r := range s {
		if unicode.IsSpace(r) {
			nextUpper = true
		} else if nextUpper {
			result.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (r *Renderer) QuestionsAsTable(questions []models.Question, languages []string) error {
	data := convertQuestions(languages, questions)
	printTable(languages, data)
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
	// re := lipgloss.NewRenderer(os.Stdout)
	// baseStyle := re.NewStyle().Padding(0, 1)
	// headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)

	// Define table headers
	// t := table.New().
	// 	Border(lipgloss.DoubleBorder()).
	// 	Headers(headers...).
	// 	Rows(questions...).
	// 	StyleFunc(func(row, col int) lipgloss.Style {
	// 		if row == table.HeaderRow {
	// 			return headerStyle
	// 		}
	// 		return baseStyle
	// 	})
	//
	// // Render table
	// fmt.Println(t.Render())
}

func (r *Renderer) ProblemsTable(questions []models.Question, languages []string) error {
	data := convert(languages, questions)
	showTable(languages, data)
	return nil
}

func convert(tracks []string, questions []models.Question) []table.Row {
	var rows []table.Row
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
		rows = append(rows, row)
	}
	return rows
}

func showTable(languages []string, rows []table.Row) {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 30},
		{Title: "Difficulty", Width: 15},
	}

	for _, lang := range languages {
		columns = append(columns, table.Column{Title: lang, Width: 4})
	}

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
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
