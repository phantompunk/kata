package models

import "github.com/charmbracelet/lipgloss"

type Quiz struct {
	Title         string
	TitleSlug     string
	Difficulty    string
	LastAttempted string
	SolutionPath  string
	Status        string
}

func (q *Quiz) Style() {
	q.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(q.Title)
}
