package ui

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/repository"
)

var quizTmpl = `
Problem: {{.Title}}
Difficulty: {{.Difficulty}}
Last attempted: {{.LastAttempted}}
Status: {{.Status}}

Next steps:
  • Start solving: kata solve {{.Slug}}
  • View details: kata show {{.Slug}}
  • Submit later: kata submit {{.Slug}}
`

func RenderQuizResult(problem *domain.Problem) error {
	t := template.Must(template.New("Quiz").Parse(quizTmpl))
	if err := t.Execute(os.Stdout, problem); err != nil {
		return err
	}
	return nil
}

var loginTmpl = `
Account:	{{.Username}}
Problems:	{{.Attempted}} attempted, {{.Completed}} completed

You're all set! 🎉

Next steps:
  • Stub problem:     kata get two-sum
  • Test solution:    kata test two-sum
  • Submit solution:  kata submit two-sum
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
