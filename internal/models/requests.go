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

type Data struct {
	Question *Question `json:"question"`
}

func GetLangName(language string) string {
	var commonName string
	switch language {
	case "go":
		commonName = "golang"
	case "python":
		commonName = "python3"
	case "c++":
		commonName = "cpp"
	case "c#":
		commonName = "csharp"
	default:
		commonName = language
	}
	return commonName
}

func (r *Response) GetQuestion(language string) (*Question, error) {
	if r != nil && r.Data != nil && r.Data.Question != nil {
		var selected CodeSnippet

		for _, snippet := range r.Data.Question.CodeSnippets {
			if snippet.LangSlug == GetLangName(language) {
				selected = snippet
				r.Data.Question.CodeSnippets = []CodeSnippet{selected}
				return r.Data.Question, nil
			}
		}
	}
	return nil, fmt.Errorf("Code snippet for %q not found", language)
}
