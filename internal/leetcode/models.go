package leetcode

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Question struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	TestCaseList []string      `json:"exampleTestcaseList"`
	Metadata     QuestionMeta  `json:"metadata"`
	LangStatus   map[string]bool
	CreatedAt    string
}

// InternalQuestion is used for unmarshaling the raw metadata field
type InternalQuestion struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	TestCaseList []string      `json:"exampleTestcaseList"`
	RawMetadata  string        `json:"metadata"`
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type QuestionMeta struct {
	Name string `json:"name"`
}

func (q *Question) UnmarshalJSON(data []byte) error {
	var tmp InternalQuestion
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.ID == "" {
		*q = Question{}
		return ErrQuestionNotFound
	}

	if tmp.RawMetadata == "" {
		*q = Question{}
		return ErrMetadataMissing
	}

	q.ID = tmp.ID
	q.Title = tmp.Title
	q.TitleSlug = tmp.TitleSlug
	q.Difficulty = tmp.Difficulty
	q.Content = tmp.Content
	q.CodeSnippets = tmp.CodeSnippets
	q.TestCaseList = tmp.TestCaseList

	if err := json.Unmarshal([]byte(tmp.RawMetadata), &q.Metadata); err != nil {
		return err
	}

	return nil
}

type QuestionReponse struct {
	Data struct {
		Question Question `json:"question"`
	} `json:"data"`
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

type SubmitResponse struct {
	InterpretID  string `json:"interpret_id"`
	SubmissionID int64  `json:"submission_id"`
	Testcase     string `json:"test_case"`
}

func (r SubmitResponse) GetSubmissionID() string {
	return fmt.Sprintf("%d", r.SubmissionID)
}

type SubmissionResponse struct {
	InterpretID       string       `json:"interpret_id"`
	SubmissionID      SubmissionID `json:"submission_id"`
	QuestionID        string       `json:"question_id"`
	State             string       `json:"state"`
	StatusMsg         string       `json:"status_msg"`
	Correct           bool         `json:"correct_answer"`
	RuntimePercentile float64      `json:"runtime_percentile"`
	StatusRuntime     string       `json:"status_runtime"`
	MemoryPercentile  float64      `json:"memory_percentile"`
	StatusMemory      string       `json:"status_memory"`
	TotalCorrect      int          `json:"total_correct"`
	TotalTestcases    int          `json:"total_testcases"`
}

func (s SubmissionResponse) ToResult() *SubmissionResult {
	return &SubmissionResult{
		Answer:     s.Correct,
		State:      s.State,
		Result:     s.StatusMsg,
		Runtime:    s.StatusRuntime,
		RuntimeMsg: "Great",
		Memory:     s.StatusMemory,
		MemoryMsg:  "Good",
	}
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

type SubmissionResult struct {
	State      string
	Answer     bool
	Result     string
	Runtime    string
	RuntimeMsg string
	Memory     string
	MemoryMsg  string
}
