package renderer

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/templates"
	"github.com/spf13/afero"
)

type Renderer struct {
	FileSystem afero.Fs
	Error      error
}

func New() Renderer {
	return Renderer{FileSystem: afero.NewOsFs()}
}

func (r *Renderer) Render(w io.Writer, problem *models.Problem, templateType string) error {
	if r.Error != nil {
		return r.Error
	}

	templ, err := template.New(templateType).ParseFS(templates.Files, "*.gohtml")
	if err != nil {
		return err
	}

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
		var buf bytes.Buffer
		err = templ.ExecuteTemplate(&buf, langBlock, problem)
		if err != nil {
			return err
		}
		problem.Code = buf.String()
	}

	if templateType == "readme" {
		markdown, err := htmltomarkdown.ConvertString(problem.Content)
		if err != nil {
			return err
		}

		problem.Content = markdown
	}

	if err = templ.ExecuteTemplate(w, templateType, problem); err != nil {
		fmt.Println("execTempl err")
		return err
	}
	return nil
}

func langTemplates(lang string) (string, string) {
	var solBlock string
	var testBlock string
	switch lang {
	case "go", "golang":
		solBlock = "golang"
		testBlock = "gotest"
	case "python", "python3":
		solBlock = "python"
		testBlock = "pytest"
	default:
		solBlock = lang
		testBlock = lang
	}
	return solBlock, testBlock
}
