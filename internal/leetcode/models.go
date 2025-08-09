package leetcode

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/repository"
)

type Question struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	TestCases    string        `json:"exampleTestcaseList"`
	FunctionName string        `json:"funcName"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	LangStatus   map[string]bool
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type Problem struct {
	QuestionID   int64
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

func (q *Question) ToProblem(workspace, language string) *Problem {
	var problem Problem
	id, _ := strconv.ParseInt(q.ID, 10, 64)
	problem.QuestionID = id
	problem.Content = q.Content
	problem.FunctionName = q.FunctionName
	problem.TitleSlug = formatTitleSlug(q.TitleSlug)
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == models.LangName[language] {
			problem.Code = snippet.Code
			problem.LangSlug = snippet.LangSlug
		}
	}
	problem.setPaths(workspace)
	return &problem
}

func (q *Question) GetCodeSnippet(language string) string {
	for _, snippet := range q.CodeSnippets {
		if snippet.LangSlug == models.LangName[language] {
			return snippet.Code
		}
	}
	return ""
}

func GetCodeSnippet(q *repository.Question, language string) string {
	fmt.Println("Code Snippets:", q.Codesnippets)
	// for _, snippet := range q.Codesnippets {
	// 	if snippet.LangSlug == models.LangName[language] {
	// 		return snippet.Code
	// 	}
	// }
	return ""
}

func ConvertToProblem(q *repository.Question, workspace, language string) *Problem {
	problem := &Problem{
		QuestionID:   q.Questionid,
		Content:      q.Content,
		FunctionName: q.Functionname,
		TitleSlug:    formatTitleSlug(q.Titleslug),
	}
	// for _, snippet := range q.Codesnippets {
	// 	if snippet.LangSlug == models.LangName[language] {
	// 		problem.Code = snippet.Code
	// 		problem.LangSlug = snippet.LangSlug
	// 	}
	// }

	return problem
}

func (q *Question) AsParams() repository.CreateParams {
	// p := q.ToProblem(app.Config.Workspace, language)
	// functionName := app.GetFunctionName(p)
	// q.FunctionName = functionName
	snippetJSON, _ := json.Marshal(q.CodeSnippets)
	return repository.CreateParams{
		Title:        q.Title,
		Titleslug:    q.TitleSlug,
		Difficulty:   q.Difficulty,
		Functionname: q.FunctionName,
		Content:      q.Content,
		Codesnippets: string(snippetJSON),
	}
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
	title := formatTitleSlug(p.TitleSlug)
	p.DirPath = filepath.Join(workspace, p.LangSlug, title)
	p.SolutionPath = filepath.Join(p.DirPath, fmt.Sprintf("%s%s", title, p.Extension()))
	p.TestPath = filepath.Join(p.DirPath, fmt.Sprintf("%s_test%s", title, p.Extension()))
	p.ReadmePath = filepath.Join(p.DirPath, "readme.md")
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
	Status_Msg  string `json:"status_msg"`
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
			if snippet.LangSlug == models.LangName[language] {
				selected = snippet
				r.Data.Question.CodeSnippets = []CodeSnippet{selected}
				return r.Data.Question, nil
			}
		}
	}
	return nil, fmt.Errorf("Code snippet for %q not found", language)
}
