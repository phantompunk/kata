package render

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/phantompunk/kata/internal/domain"
	"github.com/spf13/afero"
)

//go:embed templates/*.gohtml templates/config_template.yml
var Files embed.FS

type Renderer interface {
	RenderProblem(ctx context.Context, problem *domain.Problem, force bool) (*RenderResult, error)
}

type QuestionRenderer struct {
	fs    afero.Fs
	templ *template.Template
}

func New() (*QuestionRenderer, error) {
	funcMap := template.FuncMap{
		"pascalCase": pascalCase,
		"snakeCase":  snakeCase,
	}

	templ, err := template.New("new").Funcs(funcMap).ParseFS(Files, "templates/*")
	if err != nil {
		return nil, err
	}
	return &QuestionRenderer{fs: afero.NewOsFs(), templ: templ}, nil
}

func (r *QuestionRenderer) RenderProblem(ctx context.Context, problem *domain.Problem, force bool) (*RenderResult, error) {
	result := NewRenderResult()

	directoryCreated, err := r.ensureDirectory(problem.DirectoryPath)
	if err != nil {
		return result, err
	}

	if directoryCreated {
		result.RecordDirectoryCreated(problem.DirectoryPath)
	}

	if !directoryCreated && !force {
		result.RecordAllSkipped()
		return result, nil
	}

	for _, file := range problem.FileSet {
		if err := r.renderProblemFile(ctx, problem, file, force, result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *QuestionRenderer) renderProblemFile(ctx context.Context, problem *domain.Problem, problemFile domain.ProblemFile, force bool, result *RenderResult) error {
	fileExists := problemFile.Path.Exists()

	if fileExists && !force {
		result.RecordFileSkipped(problemFile.Path)
		return nil
	}

	file, err := r.fs.Create(problemFile.Path.String())
	if err != nil {
		return fmt.Errorf("failed creating file %q: %w", file.Name(), err)
	}
	defer file.Close()

	if err := r.renderFileContent(file, problem, problemFile); err != nil {
		return err
	}

	if fileExists {
		result.RecordFileUpdated(problemFile.Path)
	} else {
		result.RecordFileCreated(problemFile.Path)
	}
	return nil
}

func (r *QuestionRenderer) renderFileContent(w io.Writer, problem *domain.Problem, fileInfo domain.ProblemFile) error {
	switch fileInfo.Type {
	case domain.SolutionFile:
		return r.templ.ExecuteTemplate(w, problem.Language.TemplateName(), problem)

	case domain.TestFile:
		return r.templ.ExecuteTemplate(w, problem.Language.TestTemplate(), problem)

	case domain.ReadmeFile:
		markdown, err := htmltomarkdown.ConvertString(problem.Content)
		if err != nil {
			return fmt.Errorf("failed converting to markdown: %w", err)
		}

		mdProblem := *problem
		mdProblem.Content = markdown
		return r.templ.ExecuteTemplate(w, string(fileInfo.Type), mdProblem)
	}
	return nil
}

func (r *QuestionRenderer) ensureDirectory(problemDirectory domain.Path) (bool, error) {
	exists := problemDirectory.Exists()
	if err := r.fs.MkdirAll(problemDirectory.String(), os.ModePerm); err != nil {
		return false, fmt.Errorf("failed creating directory: %w", err)
	}
	return !exists, nil
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

func snakeCase(s string) string {
	return strings.ReplaceAll(string(s), "-", "_")
}
