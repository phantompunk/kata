package ui

import (
	"errors"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/leetcode"
)

// FormatError converts known error types into user-friendly messages
func FormatError(err error) string {
	switch {
	case errors.Is(err, leetcode.ErrQuestionNotFound):
		return "No matching question found. Please check the problem slug"
	case errors.Is(err, leetcode.ErrUnauthorized):
		return "Session is invalid or expired. Sign in to https://leetcode.com then run 'kata login'"
	case errors.Is(err, leetcode.ErrNotAuthenticated):
		return "Not authenticated. Please run 'kata login'"
	case errors.Is(err, app.ErrDuplicateProblem):
		return "Problem already exists, use --force to overwrite"
	case errors.Is(err, app.ErrCookiesNotFound):
		return "Session not found. Please sign in to https://leetcode.com then run 'kata login' again"
	case errors.Is(err, app.ErrInvalidSession):
		return "Session expired. Please sign in to https://leetcode.com then run 'kata login' again"
	case errors.Is(err, app.ErrNoQuestions):
		return "No questions found in the database. Please run `kata get` to fetch questions"
	case errors.Is(err, config.ErrUnsupportedLanguage):
		return "Supported languages: cpp, golang, java, python3, javascript"
	default:
		return "An unexpected error occurred. Please try again"
	}
}
