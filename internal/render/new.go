package render

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/render/templates"
	"github.com/spf13/afero"
)

type Renderer interface {
	RenderQuestion(ctx context.Context, problem *models.Problem) (*RenderResult, error)
}

type RenderResult struct {
	DirectoryCreated string
	FilesCreated     []string
	FilesUpdated     []string
}

type QuestionRenderer struct {
	fs    afero.Fs
	templ *template.Template
}

func New() (*QuestionRenderer, error) {
	funcMap := template.FuncMap{
		"pascalCase": pascalCase,
	}

	templ, err := template.New("new").Funcs(funcMap).ParseFS(templates.Files, "*")
	if err != nil {
		return nil, err
	}
	return &QuestionRenderer{fs: afero.NewOsFs(), templ: templ}, nil
}

func (r *QuestionRenderer) RenderQuestion(ctx context.Context, problem *models.Problem) (*RenderResult, error) {
	result := &RenderResult{
		FilesCreated: []string{},
		FilesUpdated: []string{},
	}

	dirExists := pathExists(problem.DirPath)
	if err := r.fs.MkdirAll(problem.DirPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed creating dirctory: %w", err)
	}

	if !dirExists {
		result.DirectoryCreated = formatPathForDisplay(problem.DirPath)
	}

	typeMapping := templateTypeMapping(problem)
	for fileType, filePath := range typeMapping {
		fileExists := pathExists(filePath)
		file, err := r.fs.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed creating file %q: %w", file.Name(), err)
		}
		defer file.Close()

		if err := r.renderFile(file, fileType, problem); err != nil {
			return nil, fmt.Errorf("failed rendering file %q: %w", file.Name(), err)
		}

		if fileExists {
			result.FilesUpdated = append(result.FilesUpdated, filepath.Base(file.Name()))
		} else {
			result.FilesCreated = append(result.FilesCreated, filepath.Base(file.Name()))
		}

	}

	return result, nil
}

func (r *QuestionRenderer) renderFile(w io.Writer, templateType templates.TemplateType, problem *models.Problem) error {
	switch templateType {
	case templates.Solution:
		sol, _ := langTemplates(problem.LangSlug)
		if sol != "" {
			return r.templ.ExecuteTemplate(w, sol, problem)
		}
		return nil
	case templates.Test:
		_, test := langTemplates(problem.LangSlug)
		if test != "" {
			return r.templ.ExecuteTemplate(w, test, problem)
		}
		return nil
	default:
		return r.templ.ExecuteTemplate(w, string(templateType), problem)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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

func templateTypeMapping(problem *models.Problem) map[templates.TemplateType]string {
	return map[templates.TemplateType]string{
		templates.Solution: problem.SolutionPath,
		templates.Test:     problem.TestPath,
		templates.Readme:   problem.ReadmePath,
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

func formatPathForDisplay(path string) string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}

	return path
}
