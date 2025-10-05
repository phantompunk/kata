package app

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/repository"
	"github.com/spf13/afero"
)

type DownloadService struct {
	repo      *repository.Queries
	client    leetcode.Client
	renderer  render.Renderer
	extractor *Extractor
}

func NewDownloadService(repo *repository.Queries, client leetcode.Client, renderer render.Renderer) *DownloadService {
	return &DownloadService{
		repo:      repo,
		client:    client,
		renderer:  renderer,
		extractor: NewExtractor(),
	}
}

func (s *DownloadService) GetQuestion(ctx context.Context, opts AppOptions) (*domain.Problem, error) {
	question, err := s.repo.GetBySlug(ctx, opts.Problem)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get question from repository: %w", err)
	}

	if question.QuestionID != 0 && !opts.Force {
		return toProblem(question, opts), nil
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

	return toProblem(createdQuestion, opts), nil
}

func (s *DownloadService) Stub(ctx context.Context, problem *domain.Problem, opts AppOptions) (*render.RenderResult, error) {
	return s.renderer.RenderProblem(ctx, problem, opts.Force)
}

func (s *DownloadService) GetRandomQuestion(ctx context.Context, opts AppOptions) (*domain.Problem, error) {
	question, err := s.repo.GetRandom(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoQuestions
		}
		return nil, fmt.Errorf("failed to get random question: %w", err)
	}

	return question.ToProblem(opts.Workspace, opts.Language), nil
}

func (s *DownloadService) SubmitQuestion(ctx context.Context, problem *domain.Problem, opts AppOptions) {
	snippet := s.extractor.ExtractSnippet(problem.SolutionPath())
	s.client.SubmitTest(ctx, problem, snippet)
}

func (s *DownloadService) GetBySlug(ctx context.Context, opts AppOptions) (*domain.Problem, error) {
	question, err := s.repo.GetBySlug(ctx, opts.Problem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	return question.ToDProblem(opts.Workspace, opts.Language), nil
}

func toProblem(question repository.Question, opts AppOptions) *domain.Problem {
	return question.ToDProblem(opts.Workspace, opts.Language)
}

type Extractor struct {
	fs afero.Fs
}

func NewExtractor() *Extractor {
	return &Extractor{fs: afero.NewOsFs()}
}

func (e Extractor) ExtractSnippet(path string) string {
	file, _ := e.fs.Open(path)
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
		// todo return error
		return ""
	}
	return strings.TrimSpace(builder.String())
}
