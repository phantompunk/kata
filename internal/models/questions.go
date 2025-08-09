package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type QuestionModel struct {
	DB      *sql.DB
	Client  *http.Client
	BaseUrl string
}

type Question struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	FunctionName string        `json:"funcName"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	TestCases    []string      `json:"exampleTestcaseList"`
	LangStatus   map[string]bool
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type Problem struct {
	QuestionID   string
	Content      string
	Code         string
	LangSlug     string
	TitleSlug    string
	FunctionName string
	SolutionPath string
	TestPath     string
	DirPath      string
	ReadmePath   string
}

func (m *QuestionModel) Exists(titleSlug string) (bool, error) {
	var exists bool

	err := m.DB.QueryRow(queryExists, titleSlug).Scan(&exists)
	return exists, err
}

// Get retrieves a Question from the local database based on its titleSlug.
func (m *QuestionModel) Get(titleSlug string) (*Question, error) {
	var question Question
	var codeJSON []byte

	row := m.DB.QueryRow(queryGetBySlug, titleSlug)
	err := row.Scan(&question.ID, &question.Title, &question.TitleSlug, &question.Content, &question.Difficulty, &question.FunctionName, &codeJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			fmt.Println("Unexpected error")
			return nil, err
		}
	}

	err = json.Unmarshal(codeJSON, &question.CodeSnippets)
	if err != nil {
		return nil, err
	}

	return &question, nil
}

func (m *QuestionModel) Insert(q *Question) (int, error) {
	snippetJSON, err := json.Marshal(q.CodeSnippets)
	if err != nil {
		return 0, err
	}

	result, err := m.DB.Exec(queryInsert, q.ID, q.Title, q.TitleSlug, q.Difficulty, q.FunctionName, q.Content, snippetJSON)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *QuestionModel) GetRandom() (Question, error) {
	var q Question
	var codeSnippetsJSON []byte
	err := m.DB.QueryRow(queryGetRandom).Scan(&q.ID, &q.Title, &q.TitleSlug, &q.Difficulty, &q.Content, &codeSnippetsJSON)
	if err != nil {
		return q, err
	}
	err = json.Unmarshal(codeSnippetsJSON, &q.CodeSnippets)
	if err != nil {
		return q, err
	}
	return q, nil
}

func (m *QuestionModel) GetAllWithStatus(languages []string) ([]Question, error) {
	stmt := buildSelectClause(languages) + buildFromClause(languages)
	fmt.Println("Executing query:", stmt)
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		q.LangStatus = make(map[string]bool)
		scanArgs := []any{&q.ID, &q.Title, &q.Difficulty}
		solvedValues := make([]int, len(languages))

		for i := range languages {
			scanArgs = append(scanArgs, &solvedValues[i])
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for i, lang := range languages {
			q.LangStatus[lang] = solvedValues[i] == 1
		}
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func (q *Question) ToProblem(workspace, language string) *Problem {
	var problem Problem
	problem.QuestionID = q.ID
	problem.Content = q.Content
	problem.FunctionName = q.FunctionName
	problem.TitleSlug = formatTitleSlug(q.TitleSlug)
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == LangName[language] {
			problem.Code = snippet.Code
			problem.LangSlug = snippet.LangSlug
		}
	}
	problem.SetPaths(workspace)
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

func (p *Problem) SetPaths(workspace string) {
	title := formatTitleSlug(p.TitleSlug)
	p.DirPath = filepath.Join(workspace, p.LangSlug, title)
	p.SolutionPath = filepath.Join(workspace, p.LangSlug, title, fmt.Sprintf("%s%s", title, p.Extension()))
	p.TestPath = filepath.Join(workspace, p.LangSlug, title, fmt.Sprintf("%s_test%s", title, p.Extension()))
	p.ReadmePath = filepath.Join(workspace, p.LangSlug, title, "readme.md")
}

var numberToString = map[string]string{"1": "one", "2": "two", "3": "three", "4": "four", "5": "five", "6": "six", "7": "seven", "8": "eight", "9": "nine", "0": "zero"}

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

func formatTitleSlug(name string) string {
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
