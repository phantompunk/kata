package app

import "errors"

var (
	ErrCookiesNotFound  = errors.New("session cookies not found")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrInvalidSession   = errors.New("session is not valid")
	ErrDuplicateProblem = errors.New("question has already been downloaded")
	ErrNoQuestions      = errors.New("no questions found in the database")
	ErrQuestionNotFound = errors.New("question not found")
	ErrSolutionFailed   = errors.New("solution failed")
)
