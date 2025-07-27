package models

import (
	"fmt"
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
	TestCase    string `json:"test_case"`
	State       string `json:"state"`
	Correct     bool   `json:"correct_answer"`
}

type StreakCounter struct {
	DaysSkipped         int  `json:"daysSkipped"`
	CurrentDayCompleted bool `json:"currentDayCompleted"`
}

type Data struct {
	Question      *Question      `json:"question"`
	StreakCounter *StreakCounter `json:"streakCounter"`
}

func (r *Response) GetQuestion(language string) (*Question, error) {
	if r != nil && r.Data != nil && r.Data.Question != nil {
		var selected CodeSnippet

		for _, snippet := range r.Data.Question.CodeSnippets {
			if snippet.LangSlug == LangName[language] {
				selected = snippet
				r.Data.Question.CodeSnippets = []CodeSnippet{selected}
				return r.Data.Question, nil
			}
		}
	}
	return nil, fmt.Errorf("Code snippet for %q not found", language)
}
