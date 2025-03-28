package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const API_URL = "https://leetcode.com/graphql"

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

func (m *QuestionModel) Fetch(name string) (*Question, error) {
	variables := map[string]any{"titleSlug": name}
	body, err := json.Marshal(Request{Query: gQLQueryQuestion, Variables: variables})
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

	// convert response to a question
	body, err = io.ReadAll(res.Body)
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Data.Question, nil
}

func (m *QuestionModel) Ping(session, token string) (bool, error) {
	body, err := json.Marshal(Request{Query: gQLQueryStreak})
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}

	req.Header.Set("referer", "https://leetcode.com/u/phantompunk/")
	req.Header.Set("origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")
	cookies := []*http.Cookie{
		{Name: "csrftoken", Value: token},
		{Name: "LEETCODE_SESSION", Value: session},
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return false, fmt.Errorf("Error creating cookie jar: %v", err)
	}

	jar.SetCookies(req.URL, cookies)
	m.Client.Jar = jar
	res, err := m.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	// convert response to a question
	if res.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil

	// body, err = io.ReadAll(res.Body)
	// var response Response
	// err = json.Unmarshal(body, &response)
	// if err != nil {
	// 	return nil, fmt.Errorf("Error unmarshalling response: %w", err)
	// }
	// fmt.Println("Hmm", response.Data.StreakCounter.CurrentDayCompleted)
	// return response.Data.StreakCounter, nil
}

func (m *QuestionModel) FetchQuestion(name string) (*Question, error) {
	// check if question has been saved before
	exists, err := m.Exists(name)
	if err != nil {
		return nil, err
	}

	if exists {
		return m.Get(name)
	}

	// fetch the question from leetcode
	question, err := m.Fetch(name)
	if err != nil {
		return nil, err
	}

	// save question to database
	_, err = m.Insert(question)
	if err != nil {
		return nil, err
	}

	return question, nil
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
	stmt := `INSERT OR REPLACE INTO questions (questionId, title, titleSlug, difficulty, functionName, content, codeSnippets) VALUES (?, ?, ?, ?, ?, ?, ?);`

	result, err := m.DB.Exec(stmt, q.ID, q.Title, q.TitleSlug, q.Difficulty, q.FunctionName, q.Content, snippetJSON)
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
	err := m.DB.QueryRow("SELECT questionId, title, titleSlug, difficulty, content, codeSnippets FROM questions ORDER BY RANDOM() LIMIT 1;").Scan(&q.ID, &q.Title, &q.TitleSlug, &q.Difficulty, &q.Content, &codeSnippetsJSON)
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
	stmt := `SELECT q.questionId, q.title, q.difficulty`
	for _, lang := range languages {
		stmt += fmt.Sprintf(", COALESCE(%s.solved, 0) AS %sSolved", lang, lang)
	}
	stmt += ` FROM questions q`
	for _, lang := range languages {
		stmt += fmt.Sprintf(" LEFT JOIN status %s ON q.questionId = %s.questionId AND %s.langSlug = '%s'", lang, lang, lang, lang)
	}

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

var packageLangName = map[string]string{
	"golang":  "go",
	"python3": "python",
}

func (q *Question) ToProblem(workspace, language string) *Problem {
	var problem Problem
	problem.QuestionID = q.ID
	problem.Content = q.Content
	problem.FunctionName = q.FunctionName
	problem.TitleSlug = formatProblemName(q.TitleSlug)
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == GetLangName(language) {
			problem.Code = snippet.Code
			problem.LangSlug = snippet.LangSlug
		}
	}
	problem.setPaths(workspace)
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

func (p *Problem) setPaths(workspace string) {
	title := formatProblemName(p.TitleSlug)
	p.DirPath = filepath.Join(workspace, packageLangName[p.LangSlug], title)
	p.SolutionPath = filepath.Join(workspace, packageLangName[p.LangSlug], title, fmt.Sprintf("%s%s", title, p.Extension()))
	p.TestPath = filepath.Join(workspace, packageLangName[p.LangSlug], title, fmt.Sprintf("%s_test%s", title, p.Extension()))
	p.ReadmePath = filepath.Join(workspace, packageLangName[p.LangSlug], title, "readme.md")
}

func (q *Question) usesLang(lang string) string {
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == GetLangName(lang) {
			return "y"
		}
	}
	return ""
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
