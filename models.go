package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type App struct {
}

type LeetCodeClient struct {
	BaseUrl    string
	fileSystem afero.Fs
	HttpClient *http.Client
}

type Problem struct {
	QuestionID string
	Content    string
	Code       string
	LangSlug   string
	TitleSlug  string
}

type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type Response struct {
	Data *Data `json:"data"`
}

type Data struct {
	Question *Question `json:"question"`
}

type Question struct {
	TitleSlug    string        `json:"titleSlug"`
	Content      string        `json:"content"`
	QuestionID   string        `json:"questionId"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

func (p *Response) GetQuestionID() string {
	if p == nil || p.Data == nil || p.Data.Question == nil {
		return ""
	}
	return p.Data.Question.QuestionID
}

func (p *Response) GetContent() string {
	if p == nil || p.Data == nil || p.Data.Question == nil {
		return ""
	}
	return p.Data.Question.Content
}

func (p *Response) GetCodeSnippets() []CodeSnippet {
	if p == nil || p.Data == nil || p.Data.Question == nil {
		return []CodeSnippet{}
	}
	return p.Data.Question.CodeSnippets
}

func (p *Response) GetLangCodeSnippet(lang string) string {
	for _, cs := range p.GetCodeSnippets() {
		if cs.LangSlug == lang {
			return cs.Code
		}
	}
	return ""
}

func (p *Response) ToProblem(language string) *Problem {
	var problem Problem
	if p != nil && p.Data != nil {
		problem.Content = p.Data.Question.Content
		problem.TitleSlug = formatProblemName(p.Data.Question.TitleSlug)
		if p.Data.Question != nil {
			problem.QuestionID = p.Data.Question.QuestionID

			for _, snippet := range p.Data.Question.CodeSnippets {
				if snippet.LangSlug == language {
					problem.Code = snippet.Code
					problem.LangSlug = snippet.LangSlug
				}
			}
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

var packageLangName = map[string]string{
	"golang":  "go",
	"python3": "python",
}

func (p *Problem) DirFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(cfg.Workspace, packageLangName[p.LangSlug], title)
}

func (p *Problem) SolutionFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(cfg.Workspace, packageLangName[p.LangSlug], title, fmt.Sprintf("%s%s", title, p.Extension()))
}

func (p *Problem) TestFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(cfg.Workspace, packageLangName[p.LangSlug], title, fmt.Sprintf("%s_test%s", title, p.Extension()))
}

func (p *Problem) ReadmeFilePath() string {
	title := formatProblemName(p.TitleSlug)
	return filepath.Join(cfg.Workspace, packageLangName[p.LangSlug], title, "readme.md")
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
