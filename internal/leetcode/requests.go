package leetcode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/phantompunk/kata/internal/models"
)

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Response struct {
	Data Data `json:"data"`
}

type TestResponse struct {
	InterpretID  string       `json:"interpret_id"`
	SubmissionID SubmissionID `json:"submission_id"`
	QuestionID   string       `json:"question_id"`
	State        string       `json:"state"`
	StatusMsg    string       `json:"status_msg"`
	Correct      bool         `json:"correct_answer"`
}

type AuthResponse struct {
	Data struct {
		UserStatus UserStatus `json:"userStatus"`
	} `json:"data"`
}

type UserStatus struct {
	IsSignedIn bool   `json:"isSignedIn"`
	Username   string `json:"username"`
}

type SubmissionID string

// Custom unmarshal to handle both int and string
func (s *SubmissionID) UnmarshalJSON(data []byte) error {
	// If it's quoted, it's a string
	if len(data) > 0 && data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = SubmissionID(str)
		return nil
	}

	// Otherwise, try integer
	var num int64
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*s = SubmissionID(strconv.FormatInt(num, 10))
	return nil
}

type StreakCounter struct {
	DaysSkipped         int  `json:"daysSkipped"`
	CurrentDayCompleted bool `json:"currentDayCompleted"`
}

type Data struct {
	Question *models.Question `json:"question"`
	Auth     *AuthResponse    `json:"userStatus"`
}

func (r *Response) GetQuestion(language string) (*models.Question, error) {
	if r != nil && r.Data.Question != nil {
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
