package ui

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/phantompunk/kata/internal/repository"
)

var quizTmpl = `
Problem: {{.Title}}
Difficulty: {{.Difficulty}}
Last attempted: {{.LastAttempted}}
Status: {{.Status}}

Next steps:
  • Start solving: kata solve {{.TitleSlug}}
  • View details: kata show {{.TitleSlug}}
  • Submit later: kata submit {{.TitleSlug}}
`

func RenderQuizResult(question *repository.GetRandomRow) string {
	var buf bytes.Buffer
	t := template.Must(template.New("Quiz").Parse(quizTmpl))
	err := t.Execute(&buf, question)
	if err != nil {
		return ""
	}

	return buf.String()
}

var loginTmpl = `
Account:	{{.Username}}
Problems:	{{.Attempted}} attempted, {{.Completed}} completed

You're all set! 🎉

Next steps:
• Try a random quiz: kata quiz
• Browse problems: kata list
• Open dashboard: kata dashboard
`

func RenderLoginResult(username string, stats repository.GetStatsRow) string {
	var buf bytes.Buffer
	t := template.Must(template.New("Login").Parse(loginTmpl))
	err := t.Execute(&buf, map[string]string{
		"Attempted": fmt.Sprint(stats.Attempted),
		"Completed": fmt.Sprint(stats.Completed),
		"Username":  username,
	})
	if err != nil {
		return ""
	}

	return buf.String()
}
