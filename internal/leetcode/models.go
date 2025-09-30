package leetcode

import (
	"encoding/json"
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

type QuestionReponse struct {
	Data struct {
		Question Question `json:"question"`
	} `json:"data"`
}

func (q *Question) UnmarshalJSON(data []byte) error {
	var tmp InternalQuestion
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.RawMetadata == "" {
		*q = Question{}
		return ErrQuestionNotFound
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
