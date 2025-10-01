package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
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
		if errors.Is(err, leetcode.ErrQuestionNotFound) {
			return nil, ErrQuestionNotFound
		}
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
	return s.renderer.RenderQuestion(ctx, problem, opts.Force)
}
