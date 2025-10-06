package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/phantompunk/kata/internal/browser"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/db"
	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/editor"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/repository"
)

type AppOptions struct {
	Problem   string
	Language  string
	Workspace string
	Open      bool
	Force     bool
}

type App struct {
	Config   *config.Config
	Repo     *repository.Queries
	lcs      *leetcode.Service
	Download *DownloadService
	Settings *config.ConfigService
}

func New() (*App, error) {
	cfgSerice, _ := config.New()

	cfg, err := cfgSerice.EnsureConfig()
	if err != nil {
		return nil, err
	}

	db, err := db.EnsureDB()
	if err != nil {
		fmt.Println("Failed db")
		return nil, err
	}

	repo := repository.New(db)

	lcs, err := leetcode.New(leetcode.WithSession2(cfg.Session))
	if err != nil {
		return nil, fmt.Errorf("failed to create leetcode service: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	client := leetcode.NewLC(leetcode.WithSession(cfg.Session))
	renderer, err := render.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create file renderer: %w", err)
	}
	download := NewDownloadService(repo, client, renderer)

	return &App{
		Config:   cfg,
		Repo:     repo,
		lcs:      lcs,
		Download: download,
		Settings: cfgSerice,
	}, nil
}

func ConvertToSlug(name string) string {
	if id, err := strconv.Atoi(name); err == nil {
		return MapIDtoSlug[id]
	}

	return name
}

// CheckSession checks if the session is valid by pinging the leetcode service
func (app *App) CheckSession() error {
	if !app.Config.HasValidSession() {
		app.Settings.ClearSession()
		return ErrInvalidSession
	}

	valid, err := app.lcs.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping leetcode service: %w", err)
	}

	if !valid {
		app.Settings.ClearSession()
		return ErrInvalidSession
	}
	return nil
}

func (app *App) ValidateSession() error {
	username, err := app.lcs.GetUsername()
	if err != nil {
		return fmt.Errorf("failed to ping leetcode service: %w", err)
	}

	if username == "" {
		app.Settings.ClearSession()
		return ErrInvalidSession
	}

	return app.Settings.SaveUsername(username)
}

// RefreshCookies fetches the session and csrf cookies from the browser and updates the app's config.
func (app *App) RefreshCookies() error {
	cookies, err := browser.GetCookies()
	if err != nil {
		return fmt.Errorf("failed to get browser cookies: %v", err)
	}

	sessionID, hasSession := cookies["LEETCODE_SESSION"]
	csrfToken, hasCSRF := cookies["csrftoken"]

	if !hasSession || !hasCSRF {
		return fmt.Errorf("required leetcode cookies not found in browser")
	}

	app.lcs.SetCookies(sessionID, csrfToken)
	return app.Settings.UpdateSession(sessionID, csrfToken)
}

func (app *App) GetAllQuestionsWithStatus(ctx context.Context) ([]domain.QuestionStat, error) {
	stats, err := app.Repo.GetAllWithStatus(ctx, app.Config.Tracks)
	if err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, ErrNoQuestions
	}

	return stats, nil
}

func (app *App) OpenQuestionInEditor(problem *domain.Problem) error {
	return editor.Open(problem.SolutionPath())
}
