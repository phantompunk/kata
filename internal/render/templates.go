package templates

import (
	"embed"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/phantompunk/kata/internal/models"
)

//go:embed templates/*.gohtml templates/*.txt
var Files embed.FS

//go:embed templates/config_template.yml
var ConfigTemplate embed.FS

type TemplateType string

type Renderer interface {
	RenderFile(w io.Writer, templateType TemplateType, problem *models.Problem) error
}

const (
	TemplateTypeProblem TemplateType = "problem"
	Solution            TemplateType = "solution"
	Readme              TemplateType = "readme"
	Test                TemplateType = "test"
	CliQuiz             TemplateType = "cliQuiz"
	CliGet              TemplateType = "cliGet"
	CliTest             TemplateType = "cliTest"
	CliSubmit           TemplateType = "cliSubmit"
	CliLogin            TemplateType = "cliLogin"
)

type FileRenderer struct {
	templ *template.Template
}

func New() (*FileRenderer, error) {
	funcMap := template.FuncMap{
		"pascalCase": pascalCase,
	}

	templ, err := template.New("").Funcs(funcMap).ParseFS(Files, "*")
	if err != nil {
		return nil, err
	}
	return &FileRenderer{templ: templ}, nil
}

func (r *FileRenderer) RenderFile(w io.Writer, templateType TemplateType, problem *models.Problem) error {
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

func (r *FileRenderer) RenderOutput(w io.Writer, templateType TemplateType, data any) error {
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
