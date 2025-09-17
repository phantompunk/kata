package templates

import (
	"embed"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/phantompunk/kata/internal/models"
)

//go:embed *.gohtml  *.tmpl
var Files embed.FS

//go:embed config_template.yml
var ConfigTemplate embed.FS

type TemplateType string

const (
	TemplateTypeProblem TemplateType = "problem"
	Solution            TemplateType = "solution"
	Readme              TemplateType = "readme"
	Test                TemplateType = "test"
	TemplateTypeQuiz    TemplateType = "quiz"
)

type Renderer struct {
	templ *template.Template
}

func New() (*Renderer, error) {
	funcMap := template.FuncMap{
		"pascalCase": pascalCase,
	}

	templ, err := template.New("").Funcs(funcMap).ParseFS(Files, "*")
	if err != nil {
		return nil, err
	}
	return &Renderer{templ: templ}, nil
}

func (r *Renderer) RenderFile(w io.Writer, templateType TemplateType, problem *models.Problem) error {
	switch templateType {
	case Solution:
		sol, _ := langTemplates(problem.LangSlug)
		if sol != "" {
			return r.templ.ExecuteTemplate(w, sol, problem)
		}
		return nil
	case Test:
		_, test := langTemplates(problem.LangSlug)
		if test != "" {
			return r.templ.ExecuteTemplate(w, test, problem)
		}
		return nil
	default:
		return r.templ.ExecuteTemplate(w, string(templateType), problem)
	}
}

func (r *Renderer) RenderOutput(w io.Writer, templateType TemplateType, data any) error {
	return r.templ.ExecuteTemplate(w, string(templateType), data)
}

func langTemplates(lang string) (string, string) {
	switch lang {
	case "go", "golang":
		return "golang", "gotest"
	case "python", "python4":
		return "python", "pytest"
	default:
		return lang, lang
	}
}

func pascalCase(s string) string {
	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate capacity

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
