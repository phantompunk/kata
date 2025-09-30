package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/render/templates"
	"github.com/phantompunk/kata/internal/repository"
)

type DownloadService struct {
	repo     *repository.Queries
	client   leetcode.Client
	renderer render.Renderer
}

func NewDownloadService(repo *repository.Queries, client leetcode.Client, renderer render.Renderer) *DownloadService {
	return &DownloadService{
		repo:     repo,
		client:   client,
		renderer: renderer,
	}
}

func (s *DownloadService) GetQuestion(ctx context.Context, opts AppOptions) (*repository.Question, error) {
	question, err := s.repo.GetBySlug(ctx, opts.Problem)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get question from repository: %w", err)
	}

	if question.QuestionID != 0 && !opts.Force {
		return &question, nil
	}

	apiQuestion, err := s.client.FetchQuestion(ctx, opts.Problem)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch question %q: %w", opts.Problem, err)
	}

	createdQuestion, err := s.repo.Create(ctx, repository.ToRepoCreateParams(apiQuestion))
	if err != nil {
		return nil, fmt.Errorf("failed to create question in repository: %w", err)
	}

	return &createdQuestion, nil
}

func (s *DownloadService) Stub(ctx context.Context, question *repository.Question, opts AppOptions, workspace string) (*render.RenderResult, error) {
	problem := question.ToProblem(workspace, opts.Language)
	return s.renderer.RenderQuestion(ctx, problem)
}

func (s *DownloadService) GetQuizQuestion(ctx context.Context) string {
	s.repo.GetRandom(ctx)
	return ""
}

// /////////////
func (app *App) Tests(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	opts.Problem = ConvertToSlug(opts.Problem)

	if err := app.CheckSession(); err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	return app.TestSolution(opts.Problem, opts.Language)
}

// TestSolution tests the solution for a given problem name and language.
func (app *App) TestSolutions(name, language string) error {
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

	problem := repoQuestion.ToProblem(app.Config.Workspace, language)
	snippet := app.extractSnippet(problem.SolutionPath)

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

func (app *App) Submits(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	opts.Problem = ConvertToSlug(opts.Problem)

	if err := app.CheckSession(); err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	return app.SubmitSolution(opts.Problem, opts.Language)
}

// SubmitSolution tests the solution for a given problem name and language.
func (app *App) SubmitSolutions(name, language string) error {
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

	problem := repoQuestion.ToProblem(app.Config.Workspace, language)
	snippet := app.extractSnippet(problem.SolutionPath)

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
