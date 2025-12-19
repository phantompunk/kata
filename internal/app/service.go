package app

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/repository"
	"github.com/spf13/afero"
)

type QuestionService struct {
	repo      *repository.Queries
	client    leetcode.Client
	renderer  render.Renderer
	extractor *Extractor
}

func NewQuestionService(repo *repository.Queries, client leetcode.Client, renderer render.Renderer) *QuestionService {
	return &QuestionService{
		repo:      repo,
		client:    client,
		renderer:  renderer,
		extractor: NewExtractor(),
	}
}

func (s *QuestionService) GetQuestion(ctx context.Context, opts AppOptions) (*domain.Problem, error) {
	question, err := s.repo.GetBySlug(ctx, opts.Problem)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get question from repository: %w", err)
	}

	if question.QuestionID != 0 && !opts.Force {
		return toProblem(question, opts) 
	}

	apiQuestion, err := s.client.FetchQuestion(ctx, opts.Problem)
	if err != nil {
		if errors.Is(err, leetcode.ErrQuestionNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("failed to fetch question %q: %w", opts.Problem, err)
	}

	// TODO: consider fire and forget
	createdQuestion, err := s.repo.Create(ctx, repository.ToRepoCreateParams(apiQuestion))
	if err != nil {
		return nil, fmt.Errorf("failed to create question in repository: %w", err)
	}

	return toProblem(createdQuestion, opts)
}

func (s *QuestionService) Stub(ctx context.Context, problem *domain.Problem, opts AppOptions) (*render.RenderResult, error) {
	return s.renderer.RenderProblem(ctx, problem, opts.Force, opts.Retry)
}

func (s *QuestionService) GetRandomQuestion(ctx context.Context, opts AppOptions) (*domain.Problem, error) {
	question, err := s.repo.GetRandom(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoQuestions
		}
		return nil, fmt.Errorf("failed to get random question: %w", err)
	}

	return question.ToProblem(opts.Workspace, opts.Language), nil
}

func (s *QuestionService) SubmitTest(ctx context.Context, problem *domain.Problem, opts AppOptions) (string, error) {
	snippet, err := s.extractor.ExtractSnippet(problem.SolutionPath())
	if err != nil {
		return "", err
	}
	submissionId, err := s.client.SubmitTest(ctx, problem, snippet)
	if err != nil {
		return "", err
	}
	return submissionId, err
}

func (s *QuestionService) SubmitSolution(ctx context.Context, problem *domain.Problem, opts AppOptions) (string, error) {
	snippet, err := s.extractor.ExtractSnippet(problem.SolutionPath())
	if err != nil {
		return "", err
	}

	submissionId, err := s.client.SubmitSolution(ctx, problem, snippet)
	if err != nil {
		return "", err
	}
	return submissionId, err
}

func (s *QuestionService) WaitForResult(ctx context.Context, problem *domain.Problem, submissionId string, maxWaitTime time.Duration) (*leetcode.SubmissionResult, error) {
	startTime := time.Now()
	pollInterval := 1 * time.Second

	for time.Since(startTime) < maxWaitTime {
		result, err := s.client.CheckSubmissionResult(ctx, submissionId)
		if err != nil {
			return nil, err
		}

		switch result.State {
		case "SUCCESS":
			now := time.Now().Format(time.RFC3339)
			s.repo.Submit(ctx, repository.SubmitParams{QuestionID: int64(problem.GetID()), LangSlug: problem.Language.Slug(), Solved: 1, LastAttempted: now})
			return result, nil
		case "PENDING", "STARTED", "EVALUATION":
			time.Sleep(pollInterval)
		case "FAILED":
			return result, ErrSolutionFailed
		default:
			return nil, fmt.Errorf("unexpected submission state: %s", result.State)
		}
	}

	return nil, fmt.Errorf("submission timed out after %v", maxWaitTime)
}

func (s *QuestionService) GetBySlug(ctx context.Context, opts AppOptions) (*domain.Problem, error) {
	question, err := s.repo.GetBySlug(ctx, opts.Problem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	return question.ToProblem(opts.Workspace, opts.Language)
}

func (s *QuestionService) GetAllQuestionsWithStatus(ctx context.Context, opts AppOptions) ([]domain.QuestionStat, error) {
	stats, err := s.repo.GetAllWithStatus(ctx, opts.Tracks)
	if err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, ErrNoQuestions
	}

	return stats, nil
}

func (s *QuestionService) GetStats(ctx context.Context) (repository.GetStatsRow, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return repository.GetStatsRow{}, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}

func toProblem(question repository.Question, opts AppOptions) (*domain.Problem, error) {
	return question.ToProblem(opts.Workspace, opts.Language)
}

type Extractor struct {
	fs afero.Fs
}

func NewExtractor() *Extractor {
	return &Extractor{fs: afero.NewOsFs()}
}

func (e Extractor) ExtractSnippet(path string) (string, error) {
	file, err := e.fs.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	startMarker := fmt.Sprint("// ::KATA START::")
	endMarker := fmt.Sprint("// ::KATA END::")

	var builder strings.Builder
	scanner := bufio.NewScanner(file)

	inSnippet := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == startMarker {
			inSnippet = true
			continue
		}

		if strings.TrimSpace(line) == endMarker {
			break
		}

		if inSnippet {
			builder.WriteString(line + "\n")
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(builder.String()), nil
}
