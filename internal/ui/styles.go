package ui

import (
	"bytes"
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
