package renderer

import (
	"bytes"
	"io"
	"strings"
	"text/template"
	"unicode"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/templates"
)

type Renderer struct {
	templ *template.Template
	Error error
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
