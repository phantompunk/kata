package app

import (
	"errors"
	"strconv"

	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/db"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/repository"
)

var (
	ErrCookiesNotFound  = errors.New("session cookies not found")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrInvalidSession   = errors.New("session is not valid")
	ErrDuplicateProblem = errors.New("question has already been downloaded")
	ErrNoQuestions      = errors.New("no questions found in the database")
	ErrQuestionNotFound = errors.New("question not found")
	ErrSolutionFailed   = errors.New("solution failed")
)

type AppOptions struct {
	Problem   string
	Language  string
	Tracks    []string
	Workspace string
	Open      bool
	Force     bool
}

type App struct {
	Config   *config.Config
	Question *QuestionService
	Setting  *config.ConfigService
	Session  *SessionService
}

func New() (*App, error) {
	settings, err := config.New()
	if err != nil {
		return nil, err
	}

	cfg, err := settings.EnsureConfig()
	if err != nil {
		return nil, err
	}

	db, err := db.EnsureDB()
	if err != nil {
		return nil, err
	}

	renderer, err := render.New()
	if err != nil {
		return nil, err
	}

	repo := repository.New(db)
	client := leetcode.NewClient(leetcode.WithSession(cfg.Session))

	download := NewQuestionService(repo, client, renderer)
	session := NewSessionService(cfg, client, settings)

	return &App{
		Config:   cfg,
		Question: download,
		Setting:  settings,
		Session:  session,
	}, nil
}

func ConvertToSlug(name string) string {
	if id, err := strconv.Atoi(name); err == nil {
		return MapIDtoSlug[id]
	}

	return name
}
