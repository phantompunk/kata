package leetcode

import (
	"fmt"

	"github.com/phantompunk/kata/internal/models"
)

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Response struct {
	Data *Data `json:"data"`
}

type TestResponse struct {
	InterpretID string `json:"interpret_id"`
	State       string `json:"state"`
	StatusMsg   string `json:"status_msg"`
	Correct     bool   `json:"correct_answer"`
}

type SubmitTestResponse struct {
	InterpretID string `json:"interpret_id"`
	State       string `json:"state"`
	StatusMsg   string `json:"status_msg"`
	Correct     bool   `json:"correct_answer"`
}

type StreakCounter struct {
	DaysSkipped         int  `json:"daysSkipped"`
	CurrentDayCompleted bool `json:"currentDayCompleted"`
}

type Data struct {
	Question      *models.Question `json:"question"`
	StreakCounter *StreakCounter   `json:"streakCounter"`
}

func (r *Response) GetQuestion(language string) (*models.Question, error) {
	if r != nil && r.Data != nil && r.Data.Question != nil {
		var selected models.CodeSnippet

		for _, snippet := range r.Data.Question.CodeSnippets {
			if snippet.LangSlug == models.LangName[language] {
				selected = snippet
				r.Data.Question.CodeSnippets = []models.CodeSnippet{selected}
				return r.Data.Question, nil
			}
		}
	}
	return nil, fmt.Errorf("Code snippet for %q not found", language)
}
