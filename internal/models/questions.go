package models

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Question struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	FunctionName string        `json:"funcName"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	TestCaseList []string      `json:"exampleTestcaseList"`
	Testcases    string        `json:"testCases"`
	LangStatus   map[string]bool
	CreatedAt    string
}

type QuestionStat struct {
	ID         string
	Title      string
	Difficulty string
	LangStatus map[string]bool
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type Problem struct {
	QuestionID    string
	Content       string
	Code          string
	LangSlug      string
	TitleSlug     string
	Slug          string
	FunctionName  string
	TestCases     string
	LastAttempted string
	SolutionPath  string
	TestPath      string
	DirPath       string
	ReadmePath    string
}

func (q *Question) ToProblem(workspace, language string) *Problem {
	var problem Problem
	problem.QuestionID = q.ID
	problem.Content = q.Content
	problem.FunctionName = q.FunctionName
	problem.LastAttempted = q.CreatedAt
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

func (p *Problem) ID() int {
	id, _ := strconv.Atoi(p.QuestionID)
	return id
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
	p.DirPath = filepath.Join(workspace, p.LangSlug, p.Slug)
	p.SolutionPath = filepath.Join(workspace, p.LangSlug, p.Slug, fmt.Sprintf("%s%s", p.Slug, p.Extension()))
	p.TestPath = filepath.Join(workspace, p.LangSlug, p.Slug, fmt.Sprintf("%s_test%s", p.Slug, p.Extension()))
	p.ReadmePath = filepath.Join(workspace, p.LangSlug, p.Slug, "readme.md")
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
