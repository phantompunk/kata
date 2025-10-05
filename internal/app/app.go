package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/phantompunk/kata/internal/browser"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/db"
	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/editor"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/render/templates"
	"github.com/phantompunk/kata/internal/repository"
)

const (
	LoginUrl = "https://leetcode.com/accounts/login/"
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
	Renderer *templates.FileRenderer
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

	renderer, err := templates.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	client := leetcode.NewLC(leetcode.WithSession(cfg.Session))
	frenderer, err := render.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create file renderer: %w", err)
	}
	download := NewDownloadService(repo, client, frenderer)

	return &App{
		Config:   cfg,
		Repo:     repo,
		lcs:      lcs,
		Renderer: renderer,
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

func (app *App) Test(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.LanguageName()
	}

	opts.Problem = ConvertToSlug(opts.Problem)

	if err := app.CheckSession(); err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	return app.TestSolution(opts.Problem, opts.Language)
}

func (app *App) Submit(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.LanguageName()
	}

	opts.Problem = ConvertToSlug(opts.Problem)

	if err := app.CheckSession(); err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	return app.SubmitSolution(opts.Problem, opts.Language)
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

// TestSolution tests the solution for a given problem name and language.
func (app *App) TestSolution(name, language string) error {
	exists, err := app.Repo.Exists(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("question %q not found", name)
	}

	repoQuestion, err := app.Repo.GetBySlug(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to get question details: %w", err)
	}

	problem := repoQuestion.ToProblem(app.Config.WorkspacePath(), language)
	snippet := "" //app.extractSnippet(problem.SolutionPath)

	testStatusUrl, err := app.lcs.Test(problem, language, snippet)
	if err != nil {
		return fmt.Errorf("failed to submit solution for testing: %w", err)
	}

	if testStatusUrl == "" {
		return fmt.Errorf("empty testStatusUrl received from server")
	}

	res, err := app.lcs.PollTestStatus(testStatusUrl)
	if err != nil {
		return fmt.Errorf("failed to poll test status: %w", err)
	}

	if res.Correct {
		fmt.Println()
		fmt.Println("✓ All test cases passed")
		if err := app.Renderer.RenderOutput(os.Stdout, templates.CliTest, problem); err != nil {
			return fmt.Errorf("failed to render quiz: %w", err)
		}
		return nil
	}

	fmt.Println()
	fmt.Println("✗ Some test cases failed")
	if err := app.Renderer.RenderOutput(os.Stdout, templates.CliTest, problem); err != nil {
		return fmt.Errorf("failed to render quiz: %w", err)
	}
	return nil
}

// SubmitSolution tests the solution for a given problem name and language.
func (app *App) SubmitSolution(name, language string) error {
	exists, err := app.Repo.Exists(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("question %q not found", name)
	}

	repoQuestion, err := app.Repo.GetBySlug(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to get question details: %w", err)
	}

	problem := repoQuestion.ToProblem(app.Config.WorkspacePath(), language)
	snippet := "" // app.extractSnippet(problem.SolutionPath)

	testStatusUrl, err := app.lcs.Submit(problem, language, snippet)
	if err != nil {
		return fmt.Errorf("failed to submit solution for testing: %w", err)
	}

	if testStatusUrl == "" {
		return fmt.Errorf("empty testStatusUrl received from server")
	}

	res, err := app.lcs.PollTestStatus(testStatusUrl)
	if err != nil {
		return fmt.Errorf("failed to poll test status: %w", err)
	}

	if res.Correct || res.StatusMsg == "Accepted" {
		fmt.Println()
		fmt.Println("✓ Submission accepted!")

		if err := app.Renderer.RenderOutput(os.Stdout, templates.CliSubmit, res); err != nil {
			return fmt.Errorf("failed to render quiz: %w", err)
		}

		app.Repo.Submit(context.Background(), repository.SubmitParams{
			QuestionID: repoQuestion.QuestionID,
			LangSlug:   language,
			Solved:     1,
		})
		return nil
	}

	fmt.Println("Failed")
	return fmt.Errorf("Failed: %s.", res.StatusMsg)
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

func (app *App) GetAllQuestionsWithStatus(ctx context.Context) ([]models.QuestionStat, error) {
	stats, err := app.Repo.GetAllWithStatus(ctx, app.Config.Tracks)
	if err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, ErrNoQuestions
	}

	return stats, nil
}

func (app *App) GetRandomQuestion(ctx context.Context) (*repository.GetRandomRow, error) {
	question, err := app.Repo.GetRandom(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no attempted problems found")
		}
		return nil, fmt.Errorf("failed to get random question: %w", err)
	}

	return &question, nil
}

func (app *App) OpenQuestionInEditor(problem *domain.Problem) error {
	return editor.Open(problem.SolutionPath())
}
