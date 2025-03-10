package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/phantompunk/kata/internal/datastore"
)

const API_URL = "https://leetcode.com/graphql"

type QuestionModel struct {
	DB      *datastore.Datastore
	Client  *http.Client
	BaseUrl string
}

type Question struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type Problem struct {
	QuestionID string
	Content    string
	Code       string
	LangSlug   string
	TitleSlug  string
}

func (m *QuestionModel) FetchQuestion(name, lang string) (*Question, error) {
	query := `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    content
    titleSlug
		title
		difficulty
    codeSnippets {
      langSlug
      code
    }
  }
}`

	variables := map[string]any{"titleSlug": name}
	body, err := json.Marshal(Request{Query: query, Variables: variables})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := m.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		// if err != nil || response.Data == nil || response.Data.Question == nil {
		// return nil, fmt.Errorf("problem not found")
		return nil, err
	}
	// return response.GetQuestion(lang)
	return response.Data.Question, nil
}

func (m *QuestionModel) GetAll() ([]Question, error) {
	return nil, nil
}

var packageLangName = map[string]string{
	"golang":  "go",
	"python3": "python",
}

func (q *Question) ToProblem(language string) *Problem {
	var problem Problem
	problem.QuestionID = q.ID
	problem.Content = q.Content
	problem.TitleSlug = formatProblemName(q.TitleSlug)
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == GetLangName(language) {
			problem.Code = snippet.Code
			problem.LangSlug = snippet.LangSlug
		}
	}
	return &problem
}

func (p *Problem) Extension() string {
	extMap := map[string]string{
		"python":  ".py",
		"python3": ".py",
		"go":      ".go",
		"golang":  ".go",
	}
	return extMap[p.LangSlug]
}

func (p *Problem) DirFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(packageLangName[p.LangSlug], title)
}

func (p *Problem) SolutionFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(packageLangName[p.LangSlug], title, fmt.Sprintf("%s%s", title, p.Extension()))
}

func (p *Problem) TestFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(packageLangName[p.LangSlug], title, fmt.Sprintf("%s_test%s", title, p.Extension()))
}

func (p *Problem) ReadmeFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(packageLangName[p.LangSlug], title, "readme.md")
}

func (q *Question) usesLang(lang string) string {
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == GetLangName(lang) {
			return "y"
		}
	}
	return ""
}

func (q *Question) UsesGo() string {
	return q.usesLang("go")
}

func (q *Question) UsesPython() string {
	return q.usesLang("python")
}

func (q *Question) DirPath() string {
	return filepath.Join()
}

var numberToString = map[string]string{"1": "one", "2": "two", "3": "three", "4": "four"}

func convertNumberToWritten(name string) string {
	letters := strings.Split(name, "")
	for i, letter := range letters {
		if hasNumber(letter) {
			written := numberToString[letter]
			letters[i] = written
		}
	}
	return strings.Join(letters, "")
}

func formatProblemName(name string) string {
	if hasNumber(name) {
		return convertNumberToWritten(name)
	}
	return strings.ReplaceAll(name, "-", "_")
}

func hasNumber(name string) bool {
	for _, char := range name {
		if '0' <= char && char <= '9' {
			return true
		}
	}
	return false
}
