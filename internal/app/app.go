package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/datastore"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/renderer"
	"github.com/spf13/afero"
)

type App struct {
	Config    config.Config
	Questions models.QuestionModel
	Renderer  renderer.Renderer
	fs        afero.Fs
}

func New() (*App, error) {
	cfg, err := config.EnsureConfig()
	if err != nil {
		fmt.Println("Failed cfg")
		return nil, err
	}

	db, err := datastore.EnsureDB(datastore.GetDbPath())
	if err != nil {
		fmt.Println("Failed db")
		return nil, err
	}

	return &App{
		Config:    cfg,
		Questions: models.QuestionModel{DB: db, Client: http.DefaultClient},
		Renderer:  renderer.New(),
		fs:        afero.NewOsFs(),
	}, nil
}

func (app *App) StubProblem(problem *models.Problem) error {
	if err := app.fs.MkdirAll(filepath.Join(app.Config.Workspace, problem.DirFilePath()), os.ModePerm); err != nil {
		return fmt.Errorf("failed creating problem directory: %w", err)
	}

	file, err := app.fs.Create(filepath.Join(app.Config.Workspace, problem.SolutionFilePath()))
	if err != nil {
		return fmt.Errorf("failed creating problem solution file: %w", err)
	}

	test, err := app.fs.Create(filepath.Join(app.Config.Workspace, problem.TestFilePath()))
	if err != nil {
		return fmt.Errorf("failed create problem test file: %w", err)
	}

	readme, err := app.fs.Create(filepath.Join(app.Config.Workspace, problem.ReadmeFilePath()))
	if err != nil {
		return fmt.Errorf("failed creating readme file: %w", err)
	}

	app.Renderer.Render(file, problem, "solution")
	app.Renderer.Render(test, problem, "test")
	app.Renderer.Render(readme, problem, "readme")
	if app.Renderer.Error != nil {
		// return app.Renderer.Error
		return fmt.Errorf("failed to render file %w", app.Renderer.Error)
	}

	return nil
}
